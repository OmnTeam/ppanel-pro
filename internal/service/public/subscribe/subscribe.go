package subscribe

import (
	"context"

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
			Id:             sub.ID,
			Name:           sub.Name,
			Language:       sub.Language,
			Description:    sub.Description,
			UnitPrice:      sub.UnitPrice,
			UnitTime:       sub.UnitTime,
			Replacement:    sub.Replacement,
			Inventory:      sub.Inventory,
			Traffic:        sub.Traffic,
			SpeedLimit:     sub.SpeedLimit,
			DeviceLimit:    sub.DeviceLimit,
			Quota:          sub.Quota,
			Nodes:          sub.Nodes,
			NodeTags:       sub.NodeTags,
			Show:           sub.Show,
			Sell:           sub.Sell,
			Sort:           sub.Sort,
			DeductionRatio: sub.DeductionRatio,
			AllowDeduction: sub.AllowDeduction,
			ResetCycle:     sub.ResetCycle,
			CreatedAt:      sub.CreatedAt,
			UpdatedAt:      sub.UpdatedAt,
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
			Total: total,
		},
	}, nil
}
