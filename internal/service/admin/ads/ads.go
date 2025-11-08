package ads

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/ads/v1"
	adsbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/ads"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdsService 广告服务
type AdsService struct {
	v1.UnimplementedAdsServiceServer

	uc  *adsbiz.AdsUsecase
	log *log.Helper
}

// NewAdsService 创建广告服务
func NewAdsService(uc *adsbiz.AdsUsecase, logger log.Logger) *AdsService {
	return &AdsService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// GetAdsList 获取广告列表
func (s *AdsService) GetAdsList(ctx context.Context, req *v1.GetAdsListRequest) (*v1.GetAdsListReply, error) {
	filter := adsbiz.AdsFilter{
		Search: req.Search,
	}

	// 处理可选的状态过滤 - 转换 bool 到 int8
	if req.Status != nil {
		status := boolToInt8Ptr(req.Status)
		filter.Status = status
	}

	total, list, err := s.uc.GetAdsListByPage(ctx, 0, req.Page, req.Size, filter)
	if err != nil {
		return nil, err
	}

	// 转换为proto对象
	pbList := make([]*v1.Ads, len(list))
	for i, ads := range list {
		pbList[i] = s.bizAdsToProto(ads)
	}

	return &v1.GetAdsListReply{
		Code:    int32(responsecode.AdminGetAdsListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetAdsListSuccess],
		Data: &v1.GetAdsListData{
			Total: total,
			List:  pbList,
		},
	}, nil
}

// GetAds 获取广告详情
func (s *AdsService) GetAds(ctx context.Context, req *v1.GetAdsRequest) (*v1.GetAdsReply, error) {
	ads, err := s.uc.GetAdsByID(ctx, 0, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetAdsReply{
		Code:    int32(responsecode.AdminGetAdsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetAdsSuccess],
		Data: &v1.GetAdsData{
			Ads: s.bizAdsToProto(ads),
		},
	}, nil
}

// CreateAds 创建广告
func (s *AdsService) CreateAds(ctx context.Context, req *v1.CreateAdsRequest) (*v1.CreateAdsReply, error) {
	ads := &adsbiz.Ads{
		Title:       req.Title,
		Type:        req.Type,
		Content:     req.Content,
		Description: req.Description,
		TargetURL:   req.TargetUrl,
		Status:      boolToInt8(req.Status), // 转换 bool 到 int8
	}

	// 处理时间字段
	if req.StartTime != nil {
		ads.StartTime = req.StartTime.AsTime()
	}
	if req.EndTime != nil {
		ads.EndTime = req.EndTime.AsTime()
	}

	result, err := s.uc.CreateAds(ctx, ads)
	if err != nil {
		return nil, err
	}

	return &v1.CreateAdsReply{
		Code:    int32(responsecode.AdminCreateAdsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateAdsSuccess],
		Data: &v1.CreateAdsData{
			Ads: s.bizAdsToProto(result),
		},
	}, nil
}

// UpdateAds 更新广告
func (s *AdsService) UpdateAds(ctx context.Context, req *v1.UpdateAdsRequest) (*v1.UpdateAdsReply, error) {
	ads := &adsbiz.Ads{
		ID:          req.Id,
		Title:       req.Title,
		Type:        req.Type,
		Content:     req.Content,
		Description: req.Description,
		TargetURL:   req.TargetUrl,
		Status:      boolToInt8(req.Status), // 转换 bool 到 int8
	}

	// 处理时间字段
	if req.StartTime != nil {
		ads.StartTime = req.StartTime.AsTime()
	}
	if req.EndTime != nil {
		ads.EndTime = req.EndTime.AsTime()
	}

	result, err := s.uc.UpdateAds(ctx, ads)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateAdsReply{
		Code:    int32(responsecode.AdminUpdateAdsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateAdsSuccess],
		Data: &v1.UpdateAdsData{
			Ads: s.bizAdsToProto(result),
		},
	}, nil
}

// DeleteAds 删除广告
func (s *AdsService) DeleteAds(ctx context.Context, req *v1.DeleteAdsRequest) (*v1.DeleteAdsReply, error) {
	err := s.uc.DeleteAds(ctx, 0, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteAdsReply{
		Code:    int32(responsecode.AdminDeleteAdsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteAdsSuccess],
		Data: &v1.DeleteAdsData{
			Success: true,
		},
	}, nil
}

// bizAdsToProto 将业务对象转换为proto对象
func (s *AdsService) bizAdsToProto(ads *adsbiz.Ads) *v1.Ads {
	if ads == nil {
		return nil
	}

	pbAds := &v1.Ads{
		Id:          ads.ID,
		Title:       ads.Title,
		Type:        ads.Type,
		Content:     ads.Content,
		Description: ads.Description,
		TargetUrl:   ads.TargetURL,
		Status:      int8ToBool(ads.Status), // 转换 int8 到 bool
	}

	// 处理时间字段
	if !ads.StartTime.IsZero() {
		pbAds.StartTime = timestamppb.New(ads.StartTime)
	}
	if !ads.EndTime.IsZero() {
		pbAds.EndTime = timestamppb.New(ads.EndTime)
	}
	if !ads.CreatedAt.IsZero() {
		pbAds.CreatedAt = timestamppb.New(ads.CreatedAt)
	}
	if !ads.UpdatedAt.IsZero() {
		pbAds.UpdatedAt = timestamppb.New(ads.UpdatedAt)
	}

	return pbAds
}

// boolToInt8 converts bool to int8 (false -> 0, true -> 1)
func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// int8ToBool converts int8 to bool (0 -> false, non-zero -> true)
func int8ToBool(i int8) bool {
	return i != 0
}

// boolToInt8Ptr converts *bool to *int8
func boolToInt8Ptr(b *bool) *int8 {
	if b == nil {
		return nil
	}
	result := boolToInt8(*b)
	return &result
}
