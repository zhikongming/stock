package model

// 事件响应结构
type EventResp struct {
	ID      uint         `json:"id"`
	Date    string       `json:"date"`
	Event   string       `json:"event"`
	Comment string       `json:"comment"`
	Stocks  []*CodeBasic `json:"stocks"`
}

// 创建事件请求
type CreateEventReq struct {
	Date    string `json:"date"`
	Event   string `json:"event"`
	Comment string `json:"comment"`
	Stocks  string `json:"stocks"`
}

// 更新事件请求
type UpdateEventReq struct {
	ID      uint   `json:"id"`
	Date    string `json:"date"`
	Event   string `json:"event"`
	Comment string `json:"comment"`
	Stocks  string `json:"stocks"`
}

// 删除事件请求
type DeleteEventReq struct {
	ID uint `json:"id"`
}

// 时间轴事件响应
type TimelineEventResp struct {
	Date   string       `json:"date"`
	Events []*EventResp `json:"events"`
}
