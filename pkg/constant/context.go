package constant

type CtxKey string

const (
	LoginType         CtxKey = "loginType"
	CtxKeyUser         CtxKey = "user"
	CtxKeySessionID    CtxKey = "sessionId"
	CtxKeyRequestHost  CtxKey = "requestHost"
	CtxKeyPlatform     CtxKey = "platform"
	CtxKeyPayment      CtxKey = "payment"
)
