package responsecode

type ResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewError(code int) *ResponseData {
	return &ResponseData{
		Code:    code,
		Message: getCodeMessage(code),
	}
}

func NewSuccess(code int, value ...interface{}) *ResponseData {
	return &ResponseData{
		Code:    code,
		Message: getCodeMessage(code),
		Data:    value,
	}
}
