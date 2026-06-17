package subscribe

import (
	"context"

	v1 "github.com/OmnTeam/npanel-pro/api/public/subscribe/v1"
	subscribeBiz "github.com/OmnTeam/npanel-pro/internal/biz/public/subscribe"
	appmiddleware "github.com/OmnTeam/npanel-pro/internal/pkg/middleware"
)

type SubscribeService struct {
	v1.UnimplementedPublicSubscribeServer
	uc *subscribeBiz.SubscribeUseCase
}

func NewSubscribeService(uc *subscribeBiz.SubscribeUseCase) *SubscribeService {
	return &SubscribeService{uc: uc}
}

func (s *SubscribeService) QuerySubscribeList(ctx context.Context, req *v1.QuerySubscribeListRequest) (*v1.QuerySubscribeListReply, error) {
	subscribes, total, err := s.uc.QuerySubscribeList(ctx, req.Language)
	if err != nil {
		return nil, err
	}

	list := make([]*v1.Subscribe, 0, len(subscribes))
	for _, sub := range subscribes {
		discounts := make([]*v1.SubscribeDiscount, 0, len(sub.Discount))
		for _, discount := range sub.Discount {
			if discount == nil {
				continue
			}
			discounts = append(discounts, &v1.SubscribeDiscount{
				Quantity: discount.Quantity,
				Discount: discount.Discount,
			})
		}

		trafficLimits := make([]*v1.TrafficLimit, 0, len(sub.TrafficLimit))
		for _, limit := range sub.TrafficLimit {
			if limit == nil {
				continue
			}
			trafficLimits = append(trafficLimits, &v1.TrafficLimit{
				StatType:     limit.StatType,
				StatValue:    limit.StatValue,
				TrafficUsage: limit.TrafficUsage,
				SpeedLimit:   int32(limit.SpeedLimit),
			})
		}

		list = append(list, &v1.Subscribe{
			Id:                sub.ID,
			Name:              sub.Name,
			Language:          sub.Language,
			Description:       sub.Description,
			UnitPrice:         sub.UnitPrice,
			UnitTime:          sub.UnitTime,
			Discount:          discounts,
			Replacement:       sub.Replacement,
			Inventory:         int32(sub.Inventory),
			Traffic:           sub.Traffic,
			SpeedLimit:        int32(sub.SpeedLimit),
			DeviceLimit:       int32(sub.DeviceLimit),
			Quota:             int32(sub.Quota),
			Nodes:             convertIntSliceToInt64Slice(sub.Nodes),
			NodeTags:          sub.NodeTags,
			NodeGroupIds:      sub.NodeGroupIds,
			NodeGroupId:       sub.NodeGroupId,
			TrafficLimit:      trafficLimits,
			Show:              sub.Show,
			Sell:              sub.Sell,
			Sort:              int32(sub.Sort),
			DeductionRatio:    int32(sub.DeductionRatio),
			AllowDeduction:    sub.AllowDeduction,
			ResetCycle:        int32(sub.ResetCycle),
			RenewalReset:      sub.RenewalReset,
			ShowOriginalPrice: sub.ShowOriginalPrice,
			CreatedAt:         sub.CreatedAt,
			UpdatedAt:         sub.UpdatedAt,
		})
	}

	return &v1.QuerySubscribeListReply{List: list, Total: total}, nil
}

func (s *SubscribeService) QueryUserSubscribeNodeList(ctx context.Context, req *v1.QueryUserSubscribeNodeListRequest) (*v1.QueryUserSubscribeNodeListReply, error) {
	userID := appmiddleware.GetUserID(ctx)
	list, err := s.uc.QueryUserSubscribeNodeList(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.UserSubscribeInfo, 0, len(list))
	for _, item := range list {
		nodes := make([]*v1.UserSubscribeNodeInfo, 0, len(item.Nodes))
		for _, node := range item.Nodes {
			if node == nil {
				continue
			}
			nodes = append(nodes, &v1.UserSubscribeNodeInfo{
				Id:              node.ID,
				Name:            node.Name,
				Uuid:            node.Uuid,
				Protocol:        node.Protocol,
				Protocols:       node.Protocols,
				Port:            node.Port,
				Address:         node.Address,
				Tags:            node.Tags,
				Country:         node.Country,
				City:            node.City,
				Longitude:       node.Longitude,
				Latitude:        node.Latitude,
				LatitudeCenter:  node.LatitudeCenter,
				LongitudeCenter: node.LongitudeCenter,
				CreatedAt:       node.CreatedAt,
				Sni:             node.SNI,

				OmniflowCarrier:                    node.OmniflowCarrier,
				OmniflowPath:                       node.OmniflowPath,
				OmniflowContentType:                node.OmniflowContentType,
				OmniflowProfilePath:                node.OmniflowProfilePath,
				OmniflowProfileJson:                node.OmniflowProfileJson,
				OmniflowServerHost:                 node.OmniflowServerHost,
				OmniflowServerPort:                 int32(node.OmniflowServerPort),
				OmniflowCaCertPath:                 node.OmniflowCaCertPath,
				OmniflowTargetMeta:                 node.OmniflowTargetMeta,
				OmniflowSpkiPin:                    node.OmniflowSpkiPin,
				OmniflowH3FallbackEnabled:          node.OmniflowH3FallbackEnabled,
				OmniflowH3FallbackPolicy:           node.OmniflowH3FallbackPolicy,
				OmniflowH3FallbackTimeoutMs:        int32(node.OmniflowH3FallbackTimeoutMs),
				OmniflowH3FallbackRetryBudget:      int32(node.OmniflowH3FallbackRetryBudget),
				OmniflowH3FallbackSmokeEnabled:     node.OmniflowH3FallbackSmokeEnabled,
				OmniflowH3FallbackSmokeIntervalSec: int32(node.OmniflowH3FallbackSmokeIntervalSec),
				OmniflowH3FallbackSmokeTimeoutMs:   int32(node.OmniflowH3FallbackSmokeTimeoutMs),
				OmniflowMaxAgeSec:                  int32(node.OmniflowMaxAgeSec),
				OmniflowIdleTimeoutSec:             int32(node.OmniflowIdleTimeoutSec),
				OmniflowMaxConnections:             int32(node.OmniflowMaxConnections),
				OmniflowAdaptiveTlsEnabled:         node.OmniflowAdaptiveTlsEnabled,
				OmniflowTlsFingerprint:             node.OmniflowTlsFingerprint,
				OmniflowSniMode:                    node.OmniflowSniMode,
				OmniflowPaddingMode:                node.OmniflowPaddingMode,
				OmniflowTrafficShapingEnabled:      node.OmniflowTrafficShapingEnabled,
				OmniflowAfEnabled:                  node.OmniflowAfEnabled,
				OmniflowAfPathMode:                 node.OmniflowAfPathMode,
				OmniflowAfPathPrefix:               node.OmniflowAfPathPrefix,
				OmniflowAfPathSuffix:               node.OmniflowAfPathSuffix,
				OmniflowAfPathRotationSecs:         int32(node.OmniflowAfPathRotationSecs),
				OmniflowAfPathSkewSlots:            int32(node.OmniflowAfPathSkewSlots),
				OmniflowFallbackEnabled:            node.OmniflowFallbackEnabled,
				OmniflowFallbackTargetScheme:       node.OmniflowFallbackTargetScheme,
				OmniflowFallbackTargetHost:         node.OmniflowFallbackTargetHost,
				OmniflowFallbackTargetPort:         int32(node.OmniflowFallbackTargetPort),
				OmniflowFallbackHostHeader:         node.OmniflowFallbackHostHeader,
				OmniflowFallbackTlsSni:             node.OmniflowFallbackTLSSNI,
				OmniflowFallbackCarrierEnabled:     node.OmniflowFallbackCarrierEnabled,
				OmniflowFallbackConnectTunnel:      node.OmniflowFallbackConnectTunnel,
				OmniflowFallbackWssEnabled:         node.OmniflowFallbackWssEnabled,
			})
		}

		items = append(items, &v1.UserSubscribeInfo{
			Id:          item.ID,
			UserId:      item.UserID,
			OrderId:     item.OrderID,
			SubscribeId: item.SubscribeID,
			StartTime:   item.StartTime,
			ExpireTime:  item.ExpireTime,
			FinishedAt:  item.FinishedAt,
			ResetTime:   item.ResetTime,
			Traffic:     item.Traffic,
			Download:    item.Download,
			Upload:      item.Upload,
			Token:       item.Token,
			Status:      uint32(item.Status),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			IsTryOut:    item.IsTryOut,
			Nodes:       nodes,
		})
	}

	return &v1.QueryUserSubscribeNodeListReply{List: items}, nil
}

func convertIntSliceToInt64Slice(input []int) []int64 {
	if len(input) == 0 {
		return []int64{}
	}
	result := make([]int64, 0, len(input))
	for _, item := range input {
		result = append(result, int64(item))
	}
	return result
}
