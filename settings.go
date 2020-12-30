package breaker

import "time"

const (
	defaultInterval = time.Duration(0) * time.Second
	defaultTimeout  = time.Duration(60) * time.Second
)

type Settings struct {
	Name          string
	MaxRequest    uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts Counts) bool
	OnStateChange func(name string, from, to State)
}

func NewBreaker(st Settings) *Breaker {
	b := new(Breaker)

	b.name = st.Name
	b.onStateChange = st.OnStateChange

	if st.MaxRequest == 0 {
		b.maxRequest = 1
	} else {
		b.maxRequest = st.MaxRequest
	}

	if st.Interval <= 0 {
		b.interval = defaultInterval
	} else {
		b.interval = st.Interval
	}

	if st.Timeout <= 0 {
		b.timeout = defaultTimeout
	} else {
		b.timeout = st.Timeout
	}

	if st.ReadyToTrip == nil {
		b.readyToTrip = defaultReadyToTrip
	} else {
		b.readyToTrip = st.ReadyToTrip
	}

	b.newVersion(time.Now())

	return b
}

func defaultReadyToTrip(counts Counts) bool {
	return counts.ContinuityFailNum > 5
}
