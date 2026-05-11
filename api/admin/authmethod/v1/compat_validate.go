package v1

// Compatibility helpers for stale validate stubs.
// The project currently does not regenerate *.pb.validate.go in `make api`,
// so these shims keep builds green while the proto contract has been updated.

func (x *PlatformListData) GetPlatforms() []*Platform {
	if x == nil {
		return nil
	}
	return x.GetList()
}

type TestSendData struct {
	Success       bool
	ResultMessage string
}

func (x *TestSendData) GetSuccess() bool {
	if x == nil {
		return false
	}
	return x.Success
}

func (x *TestSendData) GetResultMessage() string {
	if x == nil {
		return ""
	}
	return x.ResultMessage
}

type TestSendReply struct {
	Code    int32
	Message string
	Data    *TestSendData
}

func (x *TestSendReply) GetCode() int32 {
	if x == nil {
		return 0
	}
	return x.Code
}

func (x *TestSendReply) GetMessage() string {
	if x == nil {
		return ""
	}
	return x.Message
}

func (x *TestSendReply) GetData() *TestSendData {
	if x == nil {
		return nil
	}
	return x.Data
}
