package responsecode

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
)

// KratosErrorConverter Kratos错误转换器
// 将响应码转换为Kratos错误

// NewKratosError 创建带响应码的Kratos错误
// 错误消息从 codes.go 的映射表中自动获取
func NewKratosError(code int) error {
	reason := getKratosReason(code)
	message := getCodeMessage(code)

	// 根据错误码确定HTTP状态码
	var httpCode int
	switch {
	case code >= 4000000 && code < 5000000:
		// 权限错误
		if code >= 4006003000 && code < 4006003100 {
			// 认证错误 -> 401 Unauthorized
			httpCode = 401
		} else {
			// 授权错误 -> 403 Forbidden
			httpCode = 403
		}
	case code >= 3000000 && code < 4000000:
		// 业务错误 -> 400 Bad Request
		httpCode = 400
	case code >= 5000000 && code < 6000000:
		// 系统错误 -> 500 Internal Server Error
		httpCode = 500
	case code >= 6000000 && code < 7000000:
		// 网络错误 -> 503 Service Unavailable
		httpCode = 503
	case code >= 7000000 && code < 8000000:
		// 数据错误 -> 500 Internal Server Error
		httpCode = 500
	case code >= 8000000 && code < 9000000:
		// 配置错误 -> 500 Internal Server Error
		httpCode = 500
	case code >= 9000000 && code < 10000000:
		// 第三方错误 -> 500 Internal Server Error
		httpCode = 500
	default:
		// 默认为内部错误
		httpCode = 500
	}

	// 创建错误，并将自定义错误码放在metadata中
	return errors.New(httpCode, reason, message).WithMetadata(map[string]string{
		"custom_code": fmt.Sprintf("%d", code),
	})
}

// getKratosReason 根据响应码获取Kratos错误原因
func getKratosReason(code int) string {
	reasons := map[int]string{
		// 认证错误
		ErrMissingAuthToken:     "MISSING_AUTH_TOKEN",
		ErrInvalidAuthToken:     "INVALID_AUTH_TOKEN",
		ErrAuthTokenExpired:     "AUTH_TOKEN_EXPIRED",
		ErrInvalidCredentials:   "INVALID_CREDENTIALS",
		ErrUserNotAuthenticated: "USER_NOT_AUTHENTICATED",
		ErrPasswordIncorrect:    "PASSWORD_INCORRECT",
		ErrAccountLocked:        "ACCOUNT_LOCKED",
		ErrAccountDisabled:      "ACCOUNT_DISABLED",

		// 授权错误
		ErrPermissionDenied:       "PERMISSION_DENIED",
		ErrInsufficientPermission: "INSUFFICIENT_PERMISSION",
		ErrResourceAccessDenied:   "RESOURCE_ACCESS_DENIED",
		ErrOperationNotAllowed:    "OPERATION_NOT_ALLOWED",
		ErrTenantAccessDenied:     "TENANT_ACCESS_DENIED",
		ErrCrossTenantOperation:   "CROSS_TENANT_OPERATION",
		ErrNotResourceOwner:       "NOT_RESOURCE_OWNER",

		// 参数验证错误
		ErrInvalidUserID:        "INVALID_USER_ID",
		ErrInvalidTenantID:      "INVALID_TENANT_ID",
		ErrInvalidOrderID:       "INVALID_ORDER_ID",
		ErrInvalidSubscribeID:   "INVALID_SUBSCRIBE_ID",
		ErrInvalidPaymentID:     "INVALID_PAYMENT_ID",
		ErrInvalidServerID:      "INVALID_SERVER_ID",
		ErrInvalidNodeID:        "INVALID_NODE_ID",
		ErrInvalidCouponCode:    "INVALID_COUPON_CODE",
		ErrMissingRequiredParam: "MISSING_REQUIRED_PARAM",
		ErrInvalidParamFormat:   "INVALID_PARAM_FORMAT",

		// 数据不存在错误
		ErrUserNotFound:                 "USER_NOT_FOUND",
		ErrOrderNotFound:                "ORDER_NOT_FOUND",
		ErrSubscribeNotFound:            "SUBSCRIBE_NOT_FOUND",
		ErrPaymentNotFound:              "PAYMENT_NOT_FOUND",
		ErrServerNotFound:               "SERVER_NOT_FOUND",
		ErrNodeNotFound:                 "NODE_NOT_FOUND",
		ErrCouponNotFound:               "COUPON_NOT_FOUND",
		ErrDeviceNotFound:               "DEVICE_NOT_FOUND",
		ErrAuthMethodNotFound:           "AUTH_METHOD_NOT_FOUND",
		ErrAnnouncementNotFound:         "ANNOUNCEMENT_NOT_FOUND",
		ErrDocumentNotFound:             "DOCUMENT_NOT_FOUND",
		ErrAdsNotFound:                  "ADS_NOT_FOUND",
		ErrSystemNotFound:               "SYSTEM_NOT_FOUND",
		ErrSubscribeApplicationNotFound: "SUBSCRIBE_APPLICATION_NOT_FOUND",
		ErrServerGroupNotFound:          "SERVER_GROUP_NOT_FOUND",
		ErrTicketNotFound:               "TICKET_NOT_FOUND",

		// 数据冲突错误
		ErrUserAlreadyExists:                 "USER_ALREADY_EXISTS",
		ErrOrderAlreadyExists:                "ORDER_ALREADY_EXISTS",
		ErrSubscribeAlreadyExists:            "SUBSCRIBE_ALREADY_EXISTS",
		ErrPaymentAlreadyExists:              "PAYMENT_ALREADY_EXISTS",
		ErrServerAlreadyExists:               "SERVER_ALREADY_EXISTS",
		ErrNodeAlreadyExists:                 "NODE_ALREADY_EXISTS",
		ErrCouponAlreadyExists:               "COUPON_ALREADY_EXISTS",
		ErrDuplicateEmail:                    "DUPLICATE_EMAIL",
		ErrDuplicateUsername:                 "DUPLICATE_USERNAME",
		ErrTelephoneExist:                    "TELEPHONE_EXIST",
		ErrAnnouncementAlreadyExists:         "ANNOUNCEMENT_ALREADY_EXISTS",
		ErrDocumentAlreadyExists:             "DOCUMENT_ALREADY_EXISTS",
		ErrSystemAlreadyExists:               "SYSTEM_ALREADY_EXISTS",
		ErrAuthMethodAlreadyExists:           "AUTH_METHOD_ALREADY_EXISTS",
		ErrSubscribeApplicationAlreadyExists: "SUBSCRIBE_APPLICATION_ALREADY_EXISTS",

		// 业务逻辑错误
		ErrOrderCannotCancel:       "ORDER_CANNOT_CANCEL",
		ErrOrderCannotComplete:     "ORDER_CANNOT_COMPLETE",
		ErrOrderCannotClose:        "ORDER_CANNOT_CLOSE",
		ErrCouponExpired:           "COUPON_EXPIRED",
		ErrCouponNotAvailable:      "COUPON_NOT_AVAILABLE",
		ErrCouponUsedUp:            "COUPON_USED_UP",
		ErrCouponUserLimitExceeded: "COUPON_USER_LIMIT_EXCEEDED",
		ErrInsufficientBalance:     "INSUFFICIENT_BALANCE",
		ErrDeviceLimitExceeded:     "DEVICE_LIMIT_EXCEEDED",
		ErrSubscribeExpired:        "SUBSCRIBE_EXPIRED",
		ErrTrafficExceeded:         "TRAFFIC_EXCEEDED",
		ErrSubscribeInUse:          "SUBSCRIBE_IN_USE",
		ErrInvalidOrderStatus:      "INVALID_ORDER_STATUS",
		ErrInvalidParameter:        "INVALID_PARAMETER",
		ErrTitleRequired:           "TITLE_REQUIRED",
		ErrTypeRequired:            "TYPE_REQUIRED",
		ErrInvalidTimeRange:        "INVALID_TIME_RANGE",
		ErrInvalidTicketStatus:     "INVALID_TICKET_STATUS",
		ErrInvalidTicketPriority:   "INVALID_TICKET_PRIORITY",

		// 系统错误
		ErrDatabaseConnection:  "DATABASE_CONNECTION_FAILED",
		ErrDatabaseQuery:       "DATABASE_QUERY_FAILED",
		ErrDatabaseUpdate:      "DATABASE_UPDATE_FAILED",
		ErrDatabaseInsert:      "DATABASE_INSERT_FAILED",
		ErrDatabaseDelete:      "DATABASE_DELETE_FAILED",
		ErrDatabaseTransaction: "DATABASE_TRANSACTION_FAILED",
		ErrCacheConnection:     "CACHE_CONNECTION_FAILED",
		ErrCacheGet:            "CACHE_GET_FAILED",
		ErrCacheSet:            "CACHE_SET_FAILED",
		ErrInternalError:       "INTERNAL_ERROR",
		ErrServiceUnavailable:  "SERVICE_UNAVAILABLE",
		ErrConfigurationError:  "CONFIGURATION_ERROR",

		// 外部服务错误
		ErrIPGeolocationFailed: "IP_GEOLOCATION_FAILED",
		ErrPaymentGatewayError: "PAYMENT_GATEWAY_ERROR",
		ErrEmailSendFailed:     "EMAIL_SEND_FAILED",
		ErrSMSSendFailed:       "SMS_SEND_FAILED",
		ErrThirdPartyAPIError:  "THIRD_PARTY_API_ERROR",
	}

	if reason, exists := reasons[code]; exists {
		return reason
	}
	return "UNKNOWN_ERROR"
}

// ==== 便捷的Kratos错误创建方法 ====

// CreateKratosErrorFromCode 根据响应码创建Kratos错误
func CreateKratosErrorFromCode(code int) error {
	return NewKratosError(code)
}

// ==== 常用错误快捷方法 ====

// ErrUnauthorized 未认证错误
func ErrUnauthorized() error {
	return NewKratosError(ErrUserNotAuthenticated)
}

// ErrForbidden 无权限错误
func ErrForbidden() error {
	return NewKratosError(ErrPermissionDenied)
}

// ==== Order相关错误 ====

// ErrOrderIDRequired 订单ID必需
func ErrOrderIDRequired() error {
	return NewKratosError(ErrInvalidOrderID)
}

// ErrOrderCreateFailed 订单创建失败
func ErrOrderCreateFailed() error {
	return NewKratosError(ErrDatabaseInsert)
}

// ErrOrderUpdateFailed 订单更新失败
func ErrOrderUpdateFailed() error {
	return NewKratosError(ErrDatabaseUpdate)
}

// ErrOrderListFailed 订单列表获取失败
func ErrOrderListFailed() error {
	return NewKratosError(ErrDatabaseQuery)
}

// ==== 通用参数验证错误 ====

// ErrTenantIDRequired 租户ID必需
func ErrTenantIDRequired() error {
	return NewKratosError(ErrInvalidTenantID)
}

// ErrUserIDRequired 用户ID必需
func ErrUserIDRequired() error {
	return NewKratosError(ErrInvalidUserID)
}

// ErrInvalidParam 无效参数
func ErrInvalidParam() error {
	return NewKratosError(ErrInvalidParameter)
}

// ==== Task相关错误 ====

// ErrTaskCreateFailed 任务创建失败
func ErrTaskCreateFailed() error {
	return NewKratosError(ErrDatabaseInsert)
}

// ErrTaskUpdateFailed 任务更新失败
func ErrTaskUpdateFailed() error {
	return NewKratosError(ErrDatabaseUpdate)
}

// ErrTaskListFailed 任务列表获取失败
func ErrTaskListFailed() error {
	return NewKratosError(ErrDatabaseQuery)
}

// ErrTaskQueryFailed 任务查询失败
func ErrTaskQueryFailed() error {
	return NewKratosError(ErrDatabaseQuery)
}
