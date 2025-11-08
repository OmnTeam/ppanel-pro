package responsedata

type ResponseData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
	Show bool   `json:"show"`
}
