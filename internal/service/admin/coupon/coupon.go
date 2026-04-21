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

func parseStringIDs(input []string) ([]int, error) {
	result := make([]int, len(input))
	for i, v := range input {
		id, err := parseRequiredInt64(v)
		if err != nil {
			return nil, err
		}
		result[i] = int(id)
	}
	return result, nil
}

func optionalBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
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
	subscribeInt, err := parseStringIDs(req.Subscribe)
	if err != nil {
		return nil, err
	}
	if err := s.uc.CreateCoupon(ctx, req.Name, req.Code, int(req.Count), int32(req.Type), req.Discount, req.StartTime, req.ExpireTime, int64(req.UserLimit), subscribeInt, optionalBool(req.Enable)); err != nil {
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
	subscribeInt, err := parseStringIDs(req.Subscribe)
	if err != nil {
		return nil, err
	}
	if err := s.uc.UpdateCoupon(ctx, int(id), req.Name, req.Code, int(req.Count), int32(req.Type), req.Discount, req.StartTime, req.ExpireTime, int64(req.UserLimit), subscribeInt, optionalBool(req.Enable)); err != nil {
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
	idsInt, err := parseStringIDs(req.Ids)
	if err != nil {
		return nil, err
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
	var subscribe int64
	if strings.TrimSpace(req.Subscribe) != "" {
		var err error
		subscribe, err = parseRequiredInt64(req.Subscribe)
		if err != nil {
			return nil, err
		}
	}

	list, total, err := s.uc.GetCouponList(ctx, int64(req.Page), int64(req.Size), subscribe, req.Search)
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
			Count:      int64(c.Count),
			Type:       int32(c.Type),
			Discount:   int64(c.Discount),
			StartTime:  c.StartTime,
			ExpireTime: c.ExpireTime,
			UserLimit:  int64(c.UserLimit),
			Subscribe:  convertIntSliceToStringSlice(subscribeList),
			UsedCount:  int64(c.UsedCount),
			Enable:     c.Enable,
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

// convertIntSliceToStringSlice converts []int64 to []string
func convertIntSliceToStringSlice(input []int64) []string {
	if input == nil {
		return nil
	}
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = strconv.FormatInt(v, 10)
	}
	return result
}
