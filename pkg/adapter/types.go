package adapter

// Server 服务器信息
type Server struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	Config     string `json:"config"`
	Status     bool   `json:"status"`
	Tags       string `json:"tags"`
	Country    string `json:"country"`
	RelayMode  string `json:"relay_mode"`
	RelayNode  string `json:"relay_node"`
	ServerAddr string `json:"server_addr"`
}

// RuleGroup 规则组
type RuleGroup struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	Behavior string `json:"behavior"`
	Rules    string `json:"rules"`
	Tags     string `json:"tags"`
	Default  bool   `json:"default"`
}

// NodeRelay 节点中继
type NodeRelay struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Prefix string `json:"prefix"`
}

// 规则组类型常量
const (
	RuleGroupTypeReject = "reject"
	RuleGroupTypeDirect = "direct"
)

// 中继模式常量
const (
	RelayModeAll    = "all"
	RelayModeRandom = "random"
)
