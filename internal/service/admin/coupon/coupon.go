package coupon

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/coupon/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/coupon"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// CouponService coupon service implementation
type CouponService struct {
	v1.UnimplementedCouponServiceServer

	uc *coupon.CouponUseCase
}

// NewCouponService create coupon service
func NewCouponService(uc *coupon.CouponUseCase) *CouponService {
	return &CouponService{
		uc: uc,
	}
}

// CreateCoupon 创建优惠券
func (s *CouponService) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponReply, error) {
	if err := s.uc.CreateCoupon(ctx, 0, req.Name, req.Code, req.Count, req.Type, req.Discount, req.StartTime, req.ExpireTime, req.UserLimit, req.Subscribe, req.Enable); err != nil {
		return nil, err
	}
	return &v1.CreateCouponReply{
		Code:    int32(responsecode.AdminCreateCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateCouponSuccess],
	}, nil
}

// UpdateCoupon 更新优惠券
func (s *CouponService) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponReply, error) {
	if err := s.uc.UpdateCoupon(ctx, 0, req.Id, req.Name, req.Code, req.Count, req.Type, req.Discount, req.StartTime, req.ExpireTime, req.UserLimit, req.Subscribe, req.Enable); err != nil {
		return nil, err
	}
	return &v1.UpdateCouponReply{
		Code:    int32(responsecode.AdminUpdateCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateCouponSuccess],
	}, nil
}

// DeleteCoupon 删除优惠券
func (s *CouponService) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponReply, error) {
	if err := s.uc.DeleteCoupon(ctx, 0, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteCouponReply{
		Code:    int32(responsecode.AdminDeleteCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteCouponSuccess],
	}, nil
}

// BatchDeleteCoupon 批量删除优惠券
func (s *CouponService) BatchDeleteCoupon(ctx context.Context, req *v1.BatchDeleteCouponRequest) (*v1.BatchDeleteCouponReply, error) {
	if err := s.uc.BatchDeleteCoupon(ctx, 0, req.Ids); err != nil {
		return nil, err
	}
	return &v1.BatchDeleteCouponReply{
		Code:    int32(responsecode.AdminBatchDeleteCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteCouponSuccess],
	}, nil
}

// GetCouponList 获取优惠券列表
func (s *CouponService) GetCouponList(ctx context.Context, req *v1.GetCouponListRequest) (*v1.GetCouponListReply, error) {
	list, total, err := s.uc.GetCouponList(ctx, 0, req.Page, req.Size, req.Subscribe, req.Search)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	items := make([]*v1.CouponItem, 0, len(list))
	for _, c := range list {
		items = append(items, &v1.CouponItem{
			Id:         c.ID,
			Name:       c.Name,
			Code:       c.Code,
			Count:      c.Count,
			Type:       int32(c.Type),
			Discount:   c.Discount,
			StartTime:  c.StartTime.Unix(),
			ExpireTime: c.EndTime.Unix(),
			UserLimit:  0, // Field doesn't exist in schema
			Subscribe:  nil, // Field doesn't exist in schema
			UsedCount:  0, // Field doesn't exist in schema
			Enable:     c.Status == 1,
			CreatedAt:  c.CreatedAt.Unix(),
			UpdatedAt:  c.UpdatedAt.Unix(),
		})
	}

	return &v1.GetCouponListReply{
		Code:    int32(responsecode.AdminGetCouponListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetCouponListSuccess],
		Data: &v1.GetCouponListData{
			List:  items,
			Total: total,
		},
	}, nil
}
