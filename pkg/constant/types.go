package constant

import "encoding/json"

// Int64 用于类型转换标识
var Int64 int64

// VerifyType is the type of verification code
type VerifyType int32

const (
	Register VerifyType = iota + 1
	Security
)

func ParseVerifyType(i int32) VerifyType {
	return VerifyType(i)
}

func (v VerifyType) String() string {
	switch v {
	case Register:
		return "Register"
	case Security:
		return "Login"
	default:
		return "Unknown"
	}
}

// TempOrderCacheKey Redis临时订单缓存key模板
// 使用方式: fmt.Sprintf(TempOrderCacheKey, orderNo)
// TTL: 15分钟
const TempOrderCacheKey = "temp_order:%s"

// SessionIdKey Redis会话ID缓存key前缀
// 使用方式: fmt.Sprintf("%s:%s", SessionIdKey, sessionId)
// 存储值: userID
const SessionIdKey = "auth:session_id"

// TemporaryOrderInfo 临时订单信息
// 用于Portal模块，存储未注册用户的临时认证信息
type TemporaryOrderInfo struct {
	OrderNo    string `json:"order_no"`              // 订单号
	Identifier string `json:"identifier"`            // 认证标识符（邮箱/Telegram ID等）
	AuthType   string `json:"auth_type"`             // 认证类型（email/telegram等）
	Password   string `json:"password"`              // 用户密码（已加密）
	InviteCode string `json:"invite_code,omitempty"` // 邀请码（可选）
}

// Marshal 序列化为JSON
func (t *TemporaryOrderInfo) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

// Unmarshal 从JSON反序列化
func (t *TemporaryOrderInfo) Unmarshal(data []byte) error {
	return json.Unmarshal(data, t)
}
