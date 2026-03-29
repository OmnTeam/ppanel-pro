package responsecode

// User module error code aliases
const (
	ErrUserNotExist   = ErrUserNotFound   // 用户不存在（别名）
	ErrEmailExist     = ErrDuplicateEmail // 邮箱已存在（别名）
	ErrTelephoneExist = 90012             // 手机号已存在
)

func init() {
	// 添加新的错误码消息到映射表
	CodeMessages[ErrTelephoneExist] = "手机号已存在"
}
