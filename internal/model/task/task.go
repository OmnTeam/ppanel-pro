package task

import "encoding/json"

// TaskType 任务类型
type TaskType int8

const (
	TypeEmail TaskType = 0 // 邮件任务
	TypeQuota TaskType = 1 // 配额任务
)

// TaskStatus 任务状态
type TaskStatus int8

const (
	StatusPending    TaskStatus = 0 // 待处理
	StatusInProgress TaskStatus = 1 // 处理中
	StatusCompleted  TaskStatus = 2 // 已完成
	StatusFailed     TaskStatus = 3 // 失败
)

// ScopeType 范围类型
type ScopeType int8

const (
	ScopeAll     ScopeType = 1 // 所有用户
	ScopeActive  ScopeType = 2 // 活跃用户
	ScopeExpired ScopeType = 3 // 过期用户
	ScopeNone    ScopeType = 4 // 无订阅用户
	ScopeSkip    ScopeType = 5 // 跳过用户过滤
)

// ========== Email Task Models ==========

// EmailScope 邮件任务范围
type EmailScope struct {
	Type              int8     `json:"type"`                // 范围类型
	RegisterStartTime int64    `json:"register_start_time"` // 注册开始时间(毫秒时间戳)
	RegisterEndTime   int64    `json:"register_end_time"`   // 注册结束时间(毫秒时间戳)
	Recipients        []string `json:"recipients"`          // 收件人邮箱列表
	Additional        []string `json:"additional"`          // 额外的邮箱地址
	Scheduled         int64    `json:"scheduled"`           // 计划发送时间(秒时间戳)
	Interval          uint8    `json:"interval"`            // 发送间隔(秒)
	Limit             uint64   `json:"limit"`               // 每日发送限制
}

// MarshalEmailScope 序列化邮件任务范围
func MarshalEmailScope(scope *EmailScope) (string, error) {
	bytes, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UnmarshalEmailScope 反序列化邮件任务范围
func UnmarshalEmailScope(data string) (*EmailScope, error) {
	var scope EmailScope
	err := json.Unmarshal([]byte(data), &scope)
	if err != nil {
		return nil, err
	}
	return &scope, nil
}

// EmailContent 邮件任务内容
type EmailContent struct {
	Subject string `json:"subject"` // 邮件主题
	Content string `json:"content"` // 邮件内容
}

// MarshalEmailContent 序列化邮件任务内容
func MarshalEmailContent(content *EmailContent) (string, error) {
	bytes, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UnmarshalEmailContent 反序列化邮件任务内容
func UnmarshalEmailContent(data string) (*EmailContent, error) {
	var content EmailContent
	err := json.Unmarshal([]byte(data), &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// ========== Quota Task Models ==========

// QuotaScope 配额任务范围
type QuotaScope struct {
	Subscribers []int64 `json:"subscribers"` // 订阅ID列表
	IsActive    *bool   `json:"is_active"`   // 是否仅活跃订阅
	StartTime   int64   `json:"start_time"`  // 开始时间过滤(毫秒时间戳)
	EndTime     int64   `json:"end_time"`    // 结束时间过滤(毫秒时间戳)
	Objects     []int64 `json:"recipients"`  // 用户订阅ID列表(实际影响的对象)
}

// MarshalQuotaScope 序列化配额任务范围
func MarshalQuotaScope(scope *QuotaScope) (string, error) {
	bytes, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UnmarshalQuotaScope 反序列化配额任务范围
func UnmarshalQuotaScope(data string) (*QuotaScope, error) {
	var scope QuotaScope
	err := json.Unmarshal([]byte(data), &scope)
	if err != nil {
		return nil, err
	}
	return &scope, nil
}

// QuotaContent 配额任务内容
type QuotaContent struct {
	ResetTraffic bool   `json:"reset_traffic"`        // 是否重置流量
	Days         uint64 `json:"days,omitempty"`       // 增加天数
	GiftType     uint8  `json:"gift_type,omitempty"`  // 赠送类型: 1:Fixed, 2:Ratio
	GiftValue    uint64 `json:"gift_value,omitempty"` // 赠送值
}

// MarshalQuotaContent 序列化配额任务内容
func MarshalQuotaContent(content *QuotaContent) (string, error) {
	bytes, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UnmarshalQuotaContent 反序列化配额任务内容
func UnmarshalQuotaContent(data string) (*QuotaContent, error) {
	var content QuotaContent
	err := json.Unmarshal([]byte(data), &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
