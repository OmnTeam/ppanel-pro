package types

const (
	// ScheduledBatchSendEmail 批量发送邮件任务（定时）
	ScheduledBatchSendEmail = "scheduled:email:batch"

	// SchedulerCheckSubscription 定时检查订阅状态任务
	SchedulerCheckSubscription = "scheduler:check:subscription"

	// SchedulerTotalServerData 定时获取服务器总数据任务
	SchedulerTotalServerData = "scheduler:total:server"

	// SchedulerResetTraffic 定时重置流量任务
	SchedulerResetTraffic = "scheduler:reset:traffic"

	// SchedulerTrafficStat 定时流量统计任务
	SchedulerTrafficStat = "scheduler:traffic:stat"

	// SchedulerExchangeRate 定时获取汇率任务
	SchedulerExchangeRate = "scheduler:exchange:rate"

	// SchedulerRecalculateGroup 定时分组重算任务
	SchedulerRecalculateGroup = "scheduler:recalculate:group"

	// ForthwithQuotaTask 配额任务（立即执行）
	ForthwithQuotaTask = "forthwith:quota:task"

	// ForthwithSendEmail 立即发送邮件
	ForthwithSendEmail = "forthwith:email:send"

	// ForthwithSendSms 立即发送短信
	ForthwithSendSms = "forthwith:sms:send"

	// ForthwithActivateOrder 立即激活订单任务
	ForthwithActivateOrder = "forthwith:order:activate"

	// DeferCloseOrder 延迟关闭订单任务（15分钟后执行）
	DeferCloseOrder = "defer:order:close"

	// ForthwithTrafficStatistics 立即流量统计
	ForthwithTrafficStatistics = "forthwith:traffic:statistics"
)

const (
	EmailTypeVerify        = "verify"
	EmailTypeMaintenance   = "maintenance"
	EmailTypeExpiration    = "expiration"
	EmailTypeTrafficExceed = "traffic_exceed"
	EmailTypeCustom        = "custom"
)

// ForthwithActivateOrderPayload 立即激活订单任务负载
type ForthwithActivateOrderPayload struct {
	OrderNo string `json:"order_no"` // 订单号
}

// SendEmailPayload 邮件发送任务负载
type SendEmailPayload struct {
	Type    string                 `json:"type"`
	Email   string                 `json:"to"`
	Subject string                 `json:"subject"`
	Content map[string]interface{} `json:"content"`
}

// SendSmsPayload 短信发送任务负载
type SendSmsPayload struct {
	Type          int32  `json:"type"`
	Telephone     string `json:"telephone"`
	TelephoneArea string `json:"area"`
	Content       string `json:"content"`
}

// DeferCloseOrderPayload 延迟关闭订单任务负载
type DeferCloseOrderPayload struct {
	OrderNo string `json:"order_no"` // 订单号
}

// ============================================================================
// 流量统计相关类型（复刻老项目）
// ============================================================================

// UserTraffic 用户流量统计
type UserTraffic struct {
	SID      int64 `json:"uid"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

// TrafficStatistics 流量统计
type TrafficStatistics struct {
	ServerID int64         `json:"server_id"`
	Protocol string        `json:"protocol"`
	Logs     []UserTraffic `json:"logs"`
}

// OnlineUser 在线用户
type OnlineUser struct {
	UID int64  `json:"uid"`
	IP  string `json:"ip"`
}

// ServerStatus 服务器状态
type ServerStatus struct {
	CPU       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	Disk      float64 `json:"disk"`
	UpdatedAt int64   `json:"updated_at"`
}

// NodeStatus 节点状态
type NodeStatus struct {
	OnlineUsers []OnlineUser `json:"online_users"`
	Status      ServerStatus `json:"status"`
	LastAt      int64        `json:"last_at"`
}

// ServerTrafficCount 服务器流量统计
type ServerTrafficCount struct {
	ServerID  int64  `json:"server_id"`
	Name      string `json:"name"`
	Today     int64  `json:"today"`
	Yesterday int64  `json:"yesterday"`
}
