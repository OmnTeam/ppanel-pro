package model

// UserSubscribeStatus 用户订阅状态常量
type UserSubscribeStatus uint8

const (
	UserSubscribeStatusPending UserSubscribeStatus = 0 // 待激活
	UserSubscribeStatusActive  UserSubscribeStatus = 1 // 激活
	UserSubscribeStatusFinish  UserSubscribeStatus = 2 // 完成
	UserSubscribeStatusExpired UserSubscribeStatus = 3 // 过期
	UserSubscribeStatusDeduct  UserSubscribeStatus = 4 // 已扣除
)

// AuthType 认证类型常量
const (
	AuthTypeEmail    = "email"
	AuthTypeMobile   = "mobile"
	AuthTypeApple    = "apple"
	AuthTypeGoogle   = "google"
	AuthTypeGithub   = "github"
	AuthTypeFacebook = "facebook"
	AuthTypeTelegram = "telegram"
)

// Subscribe subscribe model for data layer
type Subscribe struct {
	ID                int64
	Name              string
	Language          string
	Description       string
	UnitPrice         int64
	UnitTime          string
	Discount          string // JSON string
	Replacement       int64
	Inventory         int64
	Traffic           int64
	SpeedLimit        int64
	DeviceLimit       int64
	Quota             int64
	Nodes             string // Comma-separated int64 IDs
	NodeTags          string // Comma-separated tags
	NodeGroupIDs      []int64
	NodeGroupID       int64
	TrafficLimit      string // JSON string
	Show              bool
	Sell              bool
	Sort              int64
	DeductionRatio    int64
	AllowDeduction    bool
	ResetCycle        int64
	RenewalReset      bool
	ShowOriginalPrice bool
}

// SubscribeDiscount discount configuration
type SubscribeDiscount struct {
	Quantity int64 `json:"quantity"` // Number of months
	Discount int64 `json:"discount"` // Discount value
}

// SubscribeListParams subscribe list query parameters
type SubscribeListParams struct {
	Page        int
	Size        int
	Language    string
	Search      string
	NodeGroupID int64
	IDs         []int64 // For filtering by specific IDs
}

// SubscribeGroup subscribe group model
type SubscribeGroup struct {
	ID                  int64
	Name                string
	Description         string
	IsExpiredGroup      bool
	ExpiredDaysLimit    int64
	MaxTrafficGBExpired int64
	SpeedLimit          int64
}

// TrafficLimit traffic limit configuration
type TrafficLimit struct {
	StatType     string `json:"stat_type"`
	StatValue    int64  `json:"stat_value"`
	TrafficUsage int64  `json:"traffic_usage"`
	SpeedLimit   int64  `json:"speed_limit"`
}
