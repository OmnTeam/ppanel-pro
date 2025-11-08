package responsecode

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// NewUnsupportedPlatformError 创建不支持的支付平台错误
func NewUnsupportedPlatformError() *errors.Error {
	return errors.New(ErrUnsupportedPlatform, "UNSUPPORTED_PLATFORM", CodeMessages[ErrUnsupportedPlatform])
}

// NewPaymentNotFoundError 创建支付方式不存在错误
func NewPaymentNotFoundError() *errors.Error {
	return errors.New(ErrPaymentNotFound, "PAYMENT_NOT_FOUND", CodeMessages[ErrPaymentNotFound])
}
