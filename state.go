package breaker

import "fmt"

/**
 * 熔断器状态
 *
 * 关闭状态：当前状态不拦截请求，请求直接透过
 * 半开状态：允许部分请求透过，如果执行成功，则状态置位关闭，如果请求失败，状态置位开放，等待定时器到达等待时间后重试
 * 开放状态：不允许任何请求透过
 */

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknow state: %d", s)
	}
}
