package ads

import (
	"context"
	"strconv"
	"time"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/ads/v1"
	adsbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/ads"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"

	"github.com/go-kratos/kratos/v2/log"
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

	if req.Status != nil {
		status := int32ToInt8Ptr(req.Status)
		filter.Status = status
	}

	total, list, err := s.uc.GetAdsListByPage(ctx, int(req.Page), int(req.Size), filter)
	if err != nil {
		return nil, err
	}

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
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	ads, err := s.uc.GetAdsByID(ctx, id)
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
		Status:      int32ToInt8(req.Status),
	}

	if req.StartTime > 0 {
		ads.StartTime = time.Unix(req.StartTime, 0)
	}
	if req.EndTime > 0 {
		ads.EndTime = time.Unix(req.EndTime, 0)
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
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	ads := &adsbiz.Ads{
		ID:          id,
		Title:       req.Title,
		Type:        req.Type,
		Content:     req.Content,
		Description: req.Description,
		TargetURL:   req.TargetUrl,
		Status:      int32ToInt8(req.Status),
	}

	if req.StartTime > 0 {
		ads.StartTime = time.Unix(req.StartTime, 0)
	}
	if req.EndTime > 0 {
		ads.EndTime = time.Unix(req.EndTime, 0)
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
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	err = s.uc.DeleteAds(ctx, id)
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
		Id:          strconv.FormatInt(ads.ID, 10),
		Title:       ads.Title,
		Type:        ads.Type,
		Content:     ads.Content,
		Description: ads.Description,
		TargetUrl:   ads.TargetURL,
		Status:      int8ToInt32(ads.Status),
	}

	if !ads.StartTime.IsZero() {
		pbAds.StartTime = ads.StartTime.Unix()
	}
	if !ads.EndTime.IsZero() {
		pbAds.EndTime = ads.EndTime.Unix()
	}
	if !ads.CreatedAt.IsZero() {
		pbAds.CreatedAt = ads.CreatedAt.Unix()
	}
	if !ads.UpdatedAt.IsZero() {
		pbAds.UpdatedAt = ads.UpdatedAt.Unix()
	}

	return pbAds
}

func int32ToInt8(i int32) int8 {
	return int8(i)
}

func int8ToInt32(i int8) int32 {
	return int32(i)
}

func int32ToInt8Ptr(i *int32) *int8 {
	if i == nil {
		return nil
	}
	result := int32ToInt8(*i)
	return &result
}
