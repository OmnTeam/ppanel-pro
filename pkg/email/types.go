package email

// Task 任务信息
type Task struct {
	Id      int64  `json:"id"`
	Status  uint8  `json:"status"`
	Scope   string `json:"scope"`
	Content string `json:"content"`
	Errors  string `json:"errors"`
	Current uint64 `json:"current"`
}

// EmailScope 邮件任务范围
type EmailScope struct {
	Recipients []string `json:"recipients"`
	Additional []string `json:"additional"`
	Interval   int      `json:"interval"`
}

// EmailContent 邮件内容
type EmailContent struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}
