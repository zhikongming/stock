package model

type LarkMessage struct {
	MsgType string   `json:"msg_type"` // 固定为 "interactive"
	Card    LarkCard `json:"card"`     // 卡片内容
}

// Card 对应 card 字段
type LarkCard struct {
	Schema string     `json:"schema"` // 卡片版本，如 "2.0"
	Config LarkConfig `json:"config"` // 卡片配置
	Body   LarkBody   `json:"body"`   // 卡片主体
	Header LarkHeader `json:"header"` // 卡片头部
}

// Config 卡片配置
type LarkConfig struct {
	UpdateMulti bool  `json:"update_multi"` // 是否支持多端更新
	Style       Style `json:"style"`        // 样式配置
}

// Style 样式配置
type Style struct {
	TextSize TextSize `json:"text_size"` // 文字大小
}

// TextSize 文字大小配置
type TextSize struct {
	NormalV2 NormalV2 `json:"normal_v2"` // normal_v2 样式
}

// NormalV2 不同端的文字大小
type NormalV2 struct {
	Default string `json:"default"` // 默认大小
	Pc      string `json:"pc"`      // PC 端大小
	Mobile  string `json:"mobile"`  // 移动端大小
}

// Body 卡片主体
type LarkBody struct {
	Direction         string    `json:"direction"`          // 排列方向，如 "vertical"
	HorizontalSpacing string    `json:"horizontal_spacing"` // 水平间距
	VerticalSpacing   string    `json:"vertical_spacing"`   // 垂直间距
	HorizontalAlign   string    `json:"horizontal_align"`   // 水平对齐
	VerticalAlign     string    `json:"vertical_align"`     // 垂直对齐
	Padding           string    `json:"padding"`            // 内边距
	Elements          []Element `json:"elements"`           // 元素列表，注意这里是 Element 接口
}

// Element 是一个接口，用于表示不同类型的卡片元素
type Element interface {
	isElement() // 标记方法，仅用于类型约束
}

// MarkdownElement markdown 元素
type MarkdownElement struct {
	Tag       string `json:"tag"`        // 固定为 "markdown"
	Content   string `json:"content"`    // 内容
	TextAlign string `json:"text_align"` // 对齐方式
	TextSize  string `json:"text_size"`  // 文字大小
	Margin    string `json:"margin"`     // 外边距
}

func (m MarkdownElement) isElement() {}

// TableElement 表格元素
type TableElement struct {
	Tag         string                   `json:"tag"`          // 固定为 "table"
	Columns     []Column                 `json:"columns"`      // 列定义
	Rows        []map[string]interface{} `json:"rows"`         // 行数据
	RowHeight   string                   `json:"row_height"`   // 行高
	HeaderStyle HeaderStyle              `json:"header_style"` // 表头样式
	PageSize    int                      `json:"page_size"`    // 每页显示行数
	Margin      string                   `json:"margin"`       // 外边距
}

func (t TableElement) isElement() {}

type HrElement struct {
	Tag    string `json:"tag"`    // 固定为 "hr"
	Margin string `json:"margin"` // 外边距
}

func (h HrElement) isElement() {}

// Column 列定义
type Column struct {
	DataType        string  `json:"data_type"`             // 数据类型，如 "text", "options", "number"
	Name            string  `json:"name"`                  // 字段名
	DisplayName     string  `json:"display_name"`          // 显示名
	HorizontalAlign string  `json:"horizontal_align"`      // 水平对齐
	Width           string  `json:"width"`                 // 宽度，如 "auto"
	Format          *Format `json:"format,omitempty"`      // 数字格式（可选）
	DateFormat      string  `json:"date_format,omitempty"` // 日期格式（可选）
}

// Format 数字格式
type Format struct {
	Precision int `json:"precision"` // 小数位数
}

// OptionItem 选项项，用于 customer_scale 字段
type OptionItem struct {
	Text  string `json:"text"`  // 显示文本
	Color string `json:"color"` // 颜色，如 "blue", "red"
}

// HeaderStyle 表头样式
type HeaderStyle struct {
	BackgroundStyle string `json:"background_style"` // 背景样式，如 "none"
	Bold            bool   `json:"bold"`             // 是否加粗
	Lines           int    `json:"lines"`            // 行数
}

// Header 卡片头部
type LarkHeader struct {
	Title    LarkTitle `json:"title"`    // 标题
	Subtitle LarkTitle `json:"subtitle"` // 副标题（可能为空）
	Template string    `json:"template"` // 模板颜色，如 "blue"
	Padding  string    `json:"padding"`  // 内边距
}

// Title 标题结构
type LarkTitle struct {
	Tag     string `json:"tag"`     // 固定为 "plain_text"
	Content string `json:"content"` // 文本内容
}
