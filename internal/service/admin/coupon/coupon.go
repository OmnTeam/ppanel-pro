package coupon

import (
	"context"
	"strconv"
	"strings"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/coupon/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/coupon"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func requireString(value string) error {
	if strings.TrimSpace(value) == "" {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return nil
}

func parseRequiredInt64(s string) (int64, error) {
	if err := requireString(s); err != nil {
		return 0, err
	}
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return val, nil
}

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
	if err := requireString(req.Name); err != nil {
		return nil, err
	}
	if req.Type == 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	discount, err := parseRequiredInt64(req.Discount)
	if err != nil {
		return nil, err
	}
	startTime, err := parseRequiredInt64(req.StartTime)
	if err != nil {
		return nil, err
	}
	expireTime, err := parseRequiredInt64(req.ExpireTime)
	if err != nil {
		return nil, err
	}
	subscribeInt := make([]int, len(req.Subscribe))
	for i, v := range req.Subscribe {
		subscribeInt[i] = int(v)
	}
	if err := s.uc.CreateCoupon(ctx, req.Name, req.Code, int(req.Count), int32(req.Type), discount, startTime, expireTime, int64(req.UserLimit), subscribeInt, req.Enable); err != nil {
		return nil, err
	}
	return &v1.CreateCouponReply{
		Code:    int32(responsecode.AdminCreateCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateCouponSuccess],
	}, nil
}

// UpdateCoupon 更新优惠券
func (s *CouponService) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponReply, error) {
	id, err := parseRequiredInt64(req.Id)
	if err != nil {
		return nil, err
	}
	if err := requireString(req.Name); err != nil {
		return nil, err
	}
	if req.Type == 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	discount, err := parseRequiredInt64(req.Discount)
	if err != nil {
		return nil, err
	}
	startTime, err := parseRequiredInt64(req.StartTime)
	if err != nil {
		return nil, err
	}
	expireTime, err := parseRequiredInt64(req.ExpireTime)
	if err != nil {
		return nil, err
	}
	subscribeInt := make([]int, len(req.Subscribe))
	for i, v := range req.Subscribe {
		subscribeInt[i] = int(v)
	}
	if err := s.uc.UpdateCoupon(ctx, int(id), req.Name, req.Code, int(req.Count), int32(req.Type), discount, startTime, expireTime, int64(req.UserLimit), subscribeInt, req.Enable); err != nil {
		return nil, err
	}
	return &v1.UpdateCouponReply{
		Code:    int32(responsecode.AdminUpdateCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateCouponSuccess],
	}, nil
}

// DeleteCoupon 删除优惠券
func (s *CouponService) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponReply, error) {
	id, err := parseRequiredInt64(req.Id)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteCoupon(ctx, int(id)); err != nil {
		return nil, err
	}
	return &v1.DeleteCouponReply{
		Code:    int32(responsecode.AdminDeleteCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteCouponSuccess],
	}, nil
}

// BatchDeleteCoupon 批量删除优惠券
func (s *CouponService) BatchDeleteCoupon(ctx context.Context, req *v1.BatchDeleteCouponRequest) (*v1.BatchDeleteCouponReply, error) {
	if len(req.Ids) == 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	idsInt := make([]int, len(req.Ids))
	for i, v := range req.Ids {
		idsInt[i] = int(v)
	}
	if err := s.uc.BatchDeleteCoupon(ctx, idsInt); err != nil {
		return nil, err
	}
	return &v1.BatchDeleteCouponReply{
		Code:    int32(responsecode.AdminBatchDeleteCouponSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteCouponSuccess],
	}, nil
}

// GetCouponList 获取优惠券列表
func (s *CouponService) GetCouponList(ctx context.Context, req *v1.GetCouponListRequest) (*v1.GetCouponListReply, error) {
	list, total, err := s.uc.GetCouponList(ctx, int64(req.Page), int64(req.Size), int64(req.Subscribe), req.Search)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	items := make([]*v1.CouponItem, 0, len(list))
	for _, c := range list {
		subscribeList := tool.StringToInt64Slice(c.Subscribe)

		items = append(items, &v1.CouponItem{
			Id:         formatInt64(int64(c.ID)),
			Name:       c.Name,
			Code:       c.Code,
			Count:      int32(c.Count),
			Type:       int32(c.Type),
			Discount:   formatInt64(int64(c.Discount)),
			StartTime:  formatInt64(c.StartTime),
			ExpireTime: formatInt64(c.ExpireTime),
			UserLimit:  int32(c.UserLimit),
			Subscribe:  convertIntSliceToInt32Slice(subscribeList),
			UsedCount:  int32(c.UsedCount),
			Enable:     c.Enable,
			CreatedAt:  formatInt64(c.CreatedAt.Unix()),
			UpdatedAt:  formatInt64(c.UpdatedAt.Unix()),
		})
	}

	return &v1.GetCouponListReply{
		Code:    int32(responsecode.AdminGetCouponListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetCouponListSuccess],
		Data: &v1.GetCouponListData{
			List:  items,
			Total: int32(total),
		},
	}, nil
}

// convertIntSliceToInt32Slice converts []int64 to []int32
func convertIntSliceToInt32Slice(input []int64) []int32 {
	if input == nil {
		return nil
	}
	result := make([]int32, len(input))
	for i, v := range input {
		result[i] = int32(v)
	}
	return result
}
