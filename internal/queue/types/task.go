package types

const (
	// ScheduledBatchSendEmail 批量发送邮件任务（定时）
	ScheduledBatchSendEmail = "scheduled:email:batch"

	// ScheduledCheckSubscription 定时检查订阅状态任务
	ScheduledCheckSubscription = "scheduled:subscription:check"

	// ScheduledResetTraffic 定时重置流量任务（支持三种重置模式）
	ScheduledResetTraffic = "scheduled:traffic:reset"

	// ForthwithQuotaTask 配额任务（立即执行）
	ForthwithQuotaTask = "forthwith:quota:task"

	// ForthwithSendEmail 立即发送邮件
	ForthwithSendEmail = "forthwith:email:send"

	// ForthwithSendSms 立即发送短信
	ForthwithSendSms = "forthwith:sms:send"

	// DeferCloseOrder 延迟关闭订单任务（15分钟后执行）
	DeferCloseOrder = "defer:order:close"
)

const (
	EmailTypeVerify        = "verify"
	EmailTypeMaintenance   = "maintenance"
	EmailTypeExpiration    = "expiration"
	EmailTypeTrafficExceed = "traffic_exceed"
	EmailTypeCustom        = "custom"
)

type (
	SendEmailPayload struct {
		TenantID int64                  `json:"tenant_id"` // 租户ID，用于查询租户配置
		Type     string                 `json:"type"`
		Email    string                 `json:"to"`
		Subject  string                 `json:"subject"`
		Content  map[string]interface{} `json:"content"`
	}

	SendSmsPayload struct {
		TenantID      int64  `json:"tenant_id"` // 租户ID，用于查询租户配置
		Type          int32  `json:"type"`
		Telephone     string `json:"telephone"`
		TelephoneArea string `json:"area"`
		Content       string `json:"content"`
	}

	// DeferCloseOrderPayload 延迟关闭订单任务负载
	DeferCloseOrderPayload struct {
		OrderNo string `json:"order_no"` // 订单号
	}
)

// ForthwithActivateOrder 立即激活订单任务
const ForthwithActivateOrder = "forthwith:order:activate"

// ForthwithActivateOrderPayload 立即激活订单任务负载
type ForthwithActivateOrderPayload struct {
	TenantID int64  `json:"tenant_id"`
	UserID   int64  `json:"user_id"`
	OrderNo  string `json:"order_no"`
}
