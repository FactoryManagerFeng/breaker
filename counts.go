package breaker

/**
 * 熔断器请求统计
 */

type Counts struct {
	RequestsNum          uint32 // 请求数
	SuccessNum           uint32 // 成功数
	FailNum              uint32 // 失败数
	ContinuitySuccessNum uint32 // 连续成功数
	ContinuityFailNum    uint32 // 连续失败数
}

// 添加请求次数
func (c *Counts) onRequest() {
	c.RequestsNum++
}

// 添加成功次数
func (c *Counts) onSuccess() {
	c.SuccessNum++
	c.ContinuitySuccessNum++
	c.ContinuityFailNum = 0
}

// 添加失败次数
func (c *Counts) onFail() {
	c.FailNum++
	c.ContinuityFailNum++
	c.ContinuitySuccessNum = 0
}

// 清空统计次数
func (c *Counts) clear() {
	c.RequestsNum = 0
	c.SuccessNum = 0
	c.FailNum = 0
	c.ContinuitySuccessNum = 0
	c.ContinuityFailNum = 0
}
