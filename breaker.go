package breaker

import (
	"sync"
	"time"
)

type Breaker struct {
	name          string
	maxRequest    uint32                            //请求数
	interval      time.Duration                     //closed状态下，多久清除一次计数统计
	timeout       time.Duration                     //open状态到 half-open状态切换的时间
	readyToTrip   func(counts Counts) bool          //readyToTrip熔断条件，当执行失败后，会根据readyToTrip来判断是否进入Open状态
	onStateChange func(name string, from, to State) //状态变更回调方法
	mutex         sync.Mutex                        //锁
	state         State                             //熔断状态
	version       uint64                            //递增至，用于记录当前熔断器状态切换次数，相当于一个乐观锁
	counts        Counts                            //Counts统计
	expiry        time.Time                         //超时时间，用于open状态到half-open状态的切换，超时则从open切换half-open
}

func (b *Breaker) Name() string {
	return b.name
}

func (b *Breaker) State() State {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	t := time.Now()
	state, _ := b.currentState(t)

	return state
}

// 执行
func (b *Breaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	version, err := b.beforeRequest()
	if err != nil {
		return nil, err
	}

	// 捕捉panic错误，避免因为函数错误造成熔断器panic
	defer func() {
		e := recover()
		if e != nil {
			b.afterRequest(version, false)
			panic(e)
		}
	}()
	result, err := req()
	b.afterRequest(version, err == nil)
	return result, err
}

// 请求之前调用
func (b *Breaker) beforeRequest() (uint64, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	t := time.Now()
	state, version := b.currentState(t)
	if state == StateOpen {
		// 如果状态是开启，直接返回错误
		return version, ErrOpenState
	} else if state == StateHalfOpen && b.counts.RequestsNum >= b.maxRequest {
		// 如果状态是半开启，判断请求量是否达到设置的请求数
		return version, ErrTooManyRequests
	}
	// 请求数+1
	b.counts.onRequest()
	return version, nil
}

// 请求之后调用
func (b *Breaker) afterRequest(before uint64, success bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	t := time.Now()
	state, version := b.currentState(t)
	// 校验是否是同一个请求
	if before != version {
		return
	}

	if success {
		b.onSuccess(state, t)
	} else {
		b.onFailure(state, t)
	}
}

// 请求成功，修改计数以及判断是否能将状态关闭
func (b *Breaker) onSuccess(state State, t time.Time) {
	switch state {
	case StateClosed:
		b.counts.onSuccess()
	case StateHalfOpen:
		b.counts.onSuccess()
		if b.counts.ContinuitySuccessNum >= b.maxRequest {
			b.setState(StateClosed, t)
		}
	}
}

// 请求失败，修改计数以及判断是否达到失败次数，达到则修改状态为开启
func (b *Breaker) onFailure(state State, t time.Time) {
	switch state {
	case StateClosed:
		b.counts.onFail()
		if b.readyToTrip(b.counts) {
			b.setState(StateOpen, t)
		}
	case StateHalfOpen:
		b.setState(StateOpen, t)
	}
}

// 当前状态
func (b *Breaker) currentState(t time.Time) (State, uint64) {
	switch b.state {
	case StateClosed:
		// 如果当前状态是关闭，且设置了超时时间，且超时了，则递增version到新一轮的计数
		if !b.expiry.IsZero() && b.expiry.Before(t) {
			b.newVersion(t)
		}
	case StateOpen:
		// 如果是开启状态，并且超时了，则设置为半开启状态
		if b.expiry.Before(t) {
			b.setState(StateHalfOpen, t)
		}
	}
	return b.state, b.version
}

// 设置状态
func (b *Breaker) setState(state State, t time.Time) {
	if b.state == state {
		return
	}
	preState := b.state
	b.state = state

	b.newVersion(t)

	if b.onStateChange != nil {
		b.onStateChange(b.name, preState, state)
	}
}

//设置超时时间，添加版本计数，清空统计数据，重置超时时间
func (b *Breaker) newVersion(t time.Time) {
	b.version++
	b.counts.clear()

	var zero time.Time
	switch b.state {
	case StateOpen:
		b.expiry = t.Add(b.timeout)
	case StateClosed:
		if b.interval == 0 {
			b.expiry = zero
		} else {
			b.expiry = t.Add(b.interval)
		}
	default:
		b.expiry = zero
	}
}
