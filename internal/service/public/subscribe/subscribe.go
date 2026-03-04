package subscribe

import (
	"context"
	"strconv"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/subscribe/v1"
	subscribeBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscribe"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// SubscribeService Public Subscribe服务实现
type SubscribeService struct {
	v1.UnimplementedSubscribeServer
	uc *subscribeBiz.SubscribeUseCase
}

// NewSubscribeService 创建Public Subscribe服务
func NewSubscribeService(uc *subscribeBiz.SubscribeUseCase) *SubscribeService {
	return &SubscribeService{uc: uc}
}

// QuerySubscribeList 查询订阅列表
func (s *SubscribeService) QuerySubscribeList(ctx context.Context, req *v1.QuerySubscribeListRequest) (*v1.SubscribeListReply, error) {

	// 调用业务层
	subscribes, total, err := s.uc.QuerySubscribeList(ctx, req.Language)
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.SubscribeItem, 0, len(subscribes))
	for _, sub := range subscribes {
		item := &v1.SubscribeItem{
			Id:             strconv.FormatInt(sub.ID, 10),
			Name:           sub.Name,
			Language:       sub.Language,
			Description:    sub.Description,
			UnitPrice:      strconv.FormatInt(sub.UnitPrice, 10),
			UnitTime:       sub.UnitTime,
			Replacement:    strconv.FormatInt(sub.Replacement, 10),
			Inventory:      strconv.FormatInt(sub.Inventory, 10),
			Traffic:        strconv.FormatInt(sub.Traffic, 10),
			SpeedLimit:     strconv.FormatInt(sub.SpeedLimit, 10),
			DeviceLimit:    strconv.FormatInt(sub.DeviceLimit, 10),
			Quota:          strconv.FormatInt(sub.Quota, 10),
			Nodes:          convertIntSliceToInt32Slice(sub.Nodes),
			NodeTags:       sub.NodeTags,
			Show:           sub.Show,
			Sell:           sub.Sell,
			Sort:           strconv.FormatInt(sub.Sort, 10),
			DeductionRatio: strconv.FormatInt(sub.DeductionRatio, 10),
			AllowDeduction: sub.AllowDeduction,
			ResetCycle:     sub.ResetCycle,
			CreatedAt:      strconv.FormatInt(sub.CreatedAt, 10),
			UpdatedAt:      strconv.FormatInt(sub.UpdatedAt, 10),
		}

		// 转换折扣信息
		if len(sub.Discount) > 0 {
			discounts := make([]*v1.SubscribeDiscount, 0, len(sub.Discount))
			for _, d := range sub.Discount {
				discounts = append(discounts, &v1.SubscribeDiscount{
					Quantity:   d.Quantity,
					Percentage: d.Percentage,
				})
			}
			item.Discount = discounts
		}

		list = append(list, item)
	}

	return &v1.SubscribeListReply{
		Code:    int32(responsecode.SubscribeQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.SubscribeQuerySuccess],
		Data: &v1.SubscribeListData{
			List:  list,
			Total: int32(total),
		},
	}, nil
}

// convertIntSliceToInt32Slice converts []int to []int32
func convertIntSliceToInt32Slice(input []int) []int32 {
	if input == nil {
		return nil
	}
	result := make([]int32, len(input))
	for i, v := range input {
		result[i] = int32(v)
	}
	return result
}
