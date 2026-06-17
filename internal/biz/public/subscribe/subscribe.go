package subscribe

import (
	"context"
)

// SubscribeRepo Public Subscribe数据仓库接口
type SubscribeRepo interface {
	// QuerySubscribeList 查询订阅列表
	QuerySubscribeList(ctx context.Context, language string) ([]*Subscribe, int32, error)
	QueryUserSubscribeNodeList(ctx context.Context, userID int64) ([]*UserSubscribeInfo, error)
}

// Subscribe 订阅信息
type Subscribe struct {
	ID                int64
	Name              string
	Language          string
	Description       string
	UnitPrice         int64
	UnitTime          string
	Discount          []*SubscribeDiscount
	Replacement       int64
	Inventory         int64
	Traffic           int64
	SpeedLimit        int64
	DeviceLimit       int64
	Quota             int64
	Nodes             []int
	NodeTags          []string
	NodeGroupIds      []int64
	NodeGroupId       int64
	TrafficLimit      []*TrafficLimit
	Show              bool
	Sell              bool
	Sort              int64
	DeductionRatio    int64
	AllowDeduction    bool
	ResetCycle        int64
	RenewalReset      bool
	ShowOriginalPrice bool
	CreatedAt         int64
	UpdatedAt         int64
}

// SubscribeDiscount 订阅折扣
type SubscribeDiscount struct {
	Quantity int64
	Discount float64
}

type TrafficLimit struct {
	StatType     string
	StatValue    int64
	TrafficUsage int64
	SpeedLimit   int64
}

type UserSubscribeInfo struct {
	ID          int64
	UserID      int64
	OrderID     int64
	SubscribeID int64
	StartTime   int64
	ExpireTime  int64
	FinishedAt  int64
	ResetTime   int64
	Traffic     int64
	Download    int64
	Upload      int64
	Token       string
	Status      int32
	CreatedAt   int64
	UpdatedAt   int64
	IsTryOut    bool
	Nodes       []*UserSubscribeNodeInfo
}

type UserSubscribeNodeInfo struct {
	ID              int64
	Name            string
	Uuid            string
	Protocol        string
	Protocols       string
	Port            uint32
	Address         string
	Tags            []string
	Country         string
	City            string
	Longitude       string
	Latitude        string
	LatitudeCenter  string
	LongitudeCenter string
	CreatedAt       int64

	SNI                        string
	OmniflowCarrier            string
	OmniflowPath               string
	OmniflowContentType        string
	OmniflowProfileJson        string
	OmniflowCaCertPath         string
	OmniflowTargetMeta         string
	OmniflowSpkiPin            string
	OmniflowAdaptiveTlsEnabled bool
	OmniflowTlsFingerprint     string
	OmniflowSniMode            string
	OmniflowPaddingMode        string
	OmniflowAfEnabled          bool
	OmniflowAfPathMode         string
	OmniflowAfPathPrefix       string
	OmniflowAfPathSuffix       string
	OmniflowAfPathRotationSecs int
	OmniflowAfPathSkewSlots    int
}

// SubscribeUseCase Public Subscribe用例
type SubscribeUseCase struct {
	repo SubscribeRepo
}

// NewSubscribeUseCase 创建Public Subscribe用例
func NewSubscribeUseCase(repo SubscribeRepo) *SubscribeUseCase {
	return &SubscribeUseCase{repo: repo}
}

// QuerySubscribeList 查询订阅列表
func (uc *SubscribeUseCase) QuerySubscribeList(ctx context.Context, language string) ([]*Subscribe, int32, error) {
	return uc.repo.QuerySubscribeList(ctx, language)
}

func (uc *SubscribeUseCase) QueryUserSubscribeNodeList(ctx context.Context, userID int64) ([]*UserSubscribeInfo, error) {
	return uc.repo.QueryUserSubscribeNodeList(ctx, userID)
}
