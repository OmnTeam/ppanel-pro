package data

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxynode"
	"github.com/OmnTeam/npanel-pro/ent/proxyserver"
	"github.com/OmnTeam/npanel-pro/ent/proxyservergroup"
	"github.com/OmnTeam/npanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/npanel-pro/ent/proxysystem"
	"github.com/OmnTeam/npanel-pro/ent/proxyusersubscribe"
	subscribeBiz "github.com/OmnTeam/npanel-pro/internal/biz/public/subscribe"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/OmnTeam/npanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
)

type publicSubscribeRepo struct {
	data *Data
	log  *log.Helper
}

type publicSubscribeTrafficLimit struct {
	StatType     string `json:"stat_type"`
	StatValue    int64  `json:"stat_value"`
	TrafficUsage int64  `json:"traffic_usage"`
	SpeedLimit   int64  `json:"speed_limit"`
}

// NewPublicSubscribeRepo 创建Public Subscribe仓库
func NewPublicSubscribeRepo(data *Data, logger log.Logger) subscribeBiz.SubscribeRepo {
	return &publicSubscribeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// QuerySubscribeList 查询订阅列表
func (r *publicSubscribeRepo) QuerySubscribeList(ctx context.Context, language string) ([]*subscribeBiz.Subscribe, int32, error) {
	// 查询条件: sell=true
	query := r.data.db.ProxySubscribe.Query().
		Where(
			proxysubscribe.Sell(true),
		)

	// 语言过滤：与老项目 DefaultLanguage=true 行为保持一致
	if language != "" {
		query = query.Where(
			proxysubscribe.Or(
				proxysubscribe.Language(language),
				proxysubscribe.Language(""),
			),
		)
	} else {
		query = query.Where(proxysubscribe.Language(""))
	}

	// 查询
	subscribes, err := query.Order(ent.Asc(proxysubscribe.FieldSort)).All(ctx)
	if err != nil {
		r.log.Errorf("QuerySubscribeList query error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	total := int32(len(subscribes))
	result := make([]*subscribeBiz.Subscribe, 0, len(subscribes))

	for _, s := range subscribes {
		// 处理Description（指针类型）
		desc := ""
		if s.Description != nil {
			desc = *s.Description
		}

		// 处理ResetCycle和DeductionRatio（指针类型）
		resetCycle := int64(0)
		deductionRatio := int64(0)
		if s.ResetCycle != nil {
			resetCycle = int64(*s.ResetCycle)
		}
		if s.DeductionRatio != nil {
			deductionRatio = int64(*s.DeductionRatio)
		}

		// 处理Nodes（与老项目一致，按字符串列表解析）
		nodes64 := tool.StringToInt64Slice(s.Nodes)
		nodes := make([]int, 0, len(nodes64))
		for _, nodeID := range nodes64 {
			nodes = append(nodes, int(nodeID))
		}

		// 处理NodeTags（与老项目一致，按逗号分隔）
		var nodeTags []string
		if s.NodeTags != "" {
			for _, item := range strings.Split(s.NodeTags, ",") {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					nodeTags = append(nodeTags, trimmed)
				}
			}
		}

		nodeGroupID := int64(0)
		if s.NodeGroupID != nil {
			nodeGroupID = *s.NodeGroupID
		}

		var trafficLimit []*subscribeBiz.TrafficLimit
		if s.TrafficLimit != nil && strings.TrimSpace(*s.TrafficLimit) != "" {
			var limits []publicSubscribeTrafficLimit
			if err := json.Unmarshal([]byte(*s.TrafficLimit), &limits); err == nil {
				trafficLimit = make([]*subscribeBiz.TrafficLimit, 0, len(limits))
				for _, limit := range limits {
					trafficLimit = append(trafficLimit, &subscribeBiz.TrafficLimit{
						StatType:     limit.StatType,
						StatValue:    limit.StatValue,
						TrafficUsage: limit.TrafficUsage,
						SpeedLimit:   limit.SpeedLimit,
					})
				}
			}
		}

		item := &subscribeBiz.Subscribe{
			ID:                int64(s.ID),
			Name:              s.Name,
			Language:          s.Language,
			Description:       desc,
			UnitPrice:         s.UnitPrice,
			UnitTime:          s.UnitTime,
			Replacement:       int64(s.Replacement),
			Inventory:         int64(s.Inventory),
			Traffic:           s.Traffic,
			SpeedLimit:        int64(s.SpeedLimit),
			DeviceLimit:       int64(s.DeviceLimit),
			Quota:             int64(s.Quota),
			Nodes:             nodes,
			NodeTags:          nodeTags,
			NodeGroupIds:      append([]int64{}, s.NodeGroupIds...),
			NodeGroupId:       nodeGroupID,
			TrafficLimit:      trafficLimit,
			Show:              s.Show,
			Sell:              s.Sell,
			Sort:              int64(s.Sort),
			DeductionRatio:    deductionRatio,
			AllowDeduction:    s.AllowDeduction,
			ResetCycle:        resetCycle,
			RenewalReset:      s.RenewalReset,
			ShowOriginalPrice: s.ShowOriginalPrice,
			CreatedAt:         s.CreatedAt.UnixMilli(),
			UpdatedAt:         s.UpdatedAt.UnixMilli(),
		}

		// 解析Discount字段（指针类型）
		if s.Discount != nil && *s.Discount != "" {
			var discounts []*subscribeBiz.SubscribeDiscount
			if err := json.Unmarshal([]byte(*s.Discount), &discounts); err == nil {
				item.Discount = discounts
			}
		}

		result = append(result, item)
	}

	return result, total, nil
}

func (r *publicSubscribeRepo) QueryUserSubscribeNodeList(ctx context.Context, userID int64) ([]*subscribeBiz.UserSubscribeInfo, error) {
	userSubs, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.UserIDEQ(userID),
			proxyusersubscribe.StatusIn(0, 1, 2, 3),
		).
		Order(ent.Asc(proxyusersubscribe.FieldID)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	userSubs = filterLegacyUserSubscribes(userSubs, time.Now())
	groupEnabled := legacyGroupEnabled(ctx, r.data)
	list := make([]*subscribeBiz.UserSubscribeInfo, 0, len(userSubs))
	for _, userSub := range userSubs {
		subscribePlan, err := r.data.db.ProxySubscribe.Query().
			Where(proxysubscribe.IDEQ(userSub.SubscribeID)).
			Only(ctx)
		if err != nil {
			return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		nodes, err := legacyNodeList(ctx, r.data, userSub, subscribePlan, groupEnabled)
		if err != nil {
			return nil, err
		}
		item := &subscribeBiz.UserSubscribeInfo{
			ID:          userSub.ID,
			UserID:      userSub.UserID,
			OrderID:     userSub.OrderID,
			SubscribeID: userSub.SubscribeID,
			StartTime:   userSub.StartTime.Unix(),
			ExpireTime:  legacyUnix(userSub.ExpireTime),
			FinishedAt:  unixTime(userSub.FinishedAt),
			ResetTime:   0,
			Traffic:     publicSubscribeInt64Value(userSub.Traffic),
			Download:    publicSubscribeInt64Value(userSub.Download),
			Upload:      publicSubscribeInt64Value(userSub.Upload),
			Token:       stringValue(userSub.Token),
			Status:      int32(int8Value(userSub.Status)),
			CreatedAt:   userSub.CreatedAt.Unix(),
			UpdatedAt:   userSub.UpdatedAt.Unix(),
			IsTryOut:    r.data.AppConf() != nil && r.data.AppConf().Register != nil && r.data.AppConf().Register.EnableTrial && r.data.AppConf().Register.TrialSubscribe == userSub.SubscribeID,
			Nodes:       nodes,
		}
		item.ResetTime = legacyUserResetTime(item)
		list = append(list, item)
	}
	return list, nil
}

func filterLegacyUserSubscribes(items []*ent.ProxyUserSubscribe, now time.Time) []*ent.ProxyUserSubscribe {
	result := make([]*ent.ProxyUserSubscribe, 0, len(items))
	for _, item := range items {
		if !shouldKeepLegacyUserSubscribe(item, now) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func publicSubscribeInt64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func int8Value(value *int8) int8 {
	if value == nil {
		return 0
	}
	return *value
}

func unixTime(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.Unix()
}

func legacyUserResetTime(item *subscribeBiz.UserSubscribeInfo) int64 {
	if item == nil || item.ExpireTime == 0 {
		return 0
	}
	return item.ExpireTime
}

func legacyUnix(value *time.Time) int64 {
	if value == nil || value.Unix() == 0 {
		return 0
	}
	return value.Unix()
}

func legacyGroupEnabled(ctx context.Context, d *Data) bool {
	if d == nil || d.db == nil {
		return false
	}
	item, err := d.db.ProxySystem.Query().
		Where(
			proxysystem.CategoryEQ("group"),
			proxysystem.KeyIn("enabled", "Enabled"),
		).
		First(ctx)
	if err != nil {
		return false
	}
	return item.Value == "true" || item.Value == "1"
}

func legacyNodeList(ctx context.Context, d *Data, userSub *ent.ProxyUserSubscribe, subscribePlan *ent.ProxySubscribe, groupEnabled bool) ([]*subscribeBiz.UserSubscribeNodeInfo, error) {
	now := time.Now()
	if userSub.ExpireTime != nil && userSub.ExpireTime.Unix() != 0 && userSub.ExpireTime.Before(now) {
		return legacyExpiredNodes(ctx, d, userSub)
	}
	enabledNodes, err := d.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
		).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	selected := make([]*ent.ProxyNode, 0)
	seen := make(map[int64]struct{})
	if groupEnabled {
		nodeGroupID := publicSubscribeResolveNodeGroupID(userSub, subscribePlan)
		if nodeGroupID > 0 {
			groupInfo, err := legacyAccessibleNodeGroup(ctx, d, nodeGroupID)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}
			if groupInfo == nil {
				nodeGroupID = 0
			}
		}
		directNodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)
		for _, node := range enabledNodes {
			if len(node.NodeGroupIds) == 0 {
				if _, ok := seen[node.ID]; !ok {
					seen[node.ID] = struct{}{}
					selected = append(selected, node)
				}
				continue
			}
			if nodeGroupID != 0 && tool.Contains(node.NodeGroupIds, nodeGroupID) {
				if _, ok := seen[node.ID]; !ok {
					seen[node.ID] = struct{}{}
					selected = append(selected, node)
				}
			}
		}
		for _, node := range enabledNodes {
			if tool.Contains(directNodeIDs, node.ID) {
				if _, ok := seen[node.ID]; !ok {
					seen[node.ID] = struct{}{}
					selected = append(selected, node)
				}
			}
		}
	} else {
		nodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)
		tags := make([]string, 0)
		for _, item := range strings.Split(subscribePlan.NodeTags, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
		selectedIDs := make(map[int64]struct{})
		for _, node := range enabledNodes {
			if len(nodeIDs) == 0 && len(tags) == 0 {
				continue
			}
			matched := false
			if len(nodeIDs) > 0 && tool.Contains(nodeIDs, node.ID) {
				matched = true
			}
			if len(tags) > 0 && nodeMatchesTags(node.Tags, tags) {
				matched = true
			}
			if !matched {
				continue
			}
			if _, ok := selectedIDs[node.ID]; ok {
				continue
			}
			selectedIDs[node.ID] = struct{}{}
			selected = append(selected, node)
		}
	}
	return buildLegacyNodeInfos(ctx, d, userSub, selected)
}

func legacyExpiredNodes(ctx context.Context, d *Data, userSub *ent.ProxyUserSubscribe) ([]*subscribeBiz.UserSubscribeNodeInfo, error) {
	expiredGroup, err := d.db.ProxyServerGroup.Query().
		Where(proxyservergroup.IsExpiredGroupEQ(true)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if !isNodeGroupTypeAccessible(expiredGroup.GroupType, nodeGroupAccessApp) {
		return nil, nil
	}
	if userSub.ExpireTime == nil {
		return nil, nil
	}
	expiredDays := int(time.Since(*userSub.ExpireTime).Hours() / 24)
	if expiredDays > expiredGroup.ExpiredDaysLimit {
		return nil, nil
	}
	if expiredGroup.MaxTrafficGBExpired != nil && *expiredGroup.MaxTrafficGBExpired > 0 {
		usedTrafficGB := float64(publicSubscribeInt64Value(userSub.ExpiredDownload)+publicSubscribeInt64Value(userSub.ExpiredUpload)) / (1024 * 1024 * 1024)
		if usedTrafficGB >= float64(*expiredGroup.MaxTrafficGBExpired) {
			return nil, nil
		}
	}
	nodes, err := d.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
		).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	selected := make([]*ent.ProxyNode, 0)
	for _, node := range nodes {
		if tool.Contains(node.NodeGroupIds, expiredGroup.ID) {
			selected = append(selected, node)
		}
	}
	return buildLegacyNodeInfos(ctx, d, userSub, selected)
}

func publicSubscribeResolveNodeGroupID(userSub *ent.ProxyUserSubscribe, subscribePlan *ent.ProxySubscribe) int64 {
	if userSub != nil && userSub.NodeGroupID != 0 {
		return userSub.NodeGroupID
	}
	if subscribePlan != nil && subscribePlan.NodeGroupID != nil && *subscribePlan.NodeGroupID != 0 {
		return *subscribePlan.NodeGroupID
	}
	if subscribePlan != nil && len(subscribePlan.NodeGroupIds) > 0 {
		return subscribePlan.NodeGroupIds[0]
	}
	return 0
}

const (
	nodeGroupTypeCommon = "common"
	nodeGroupTypeApp    = "app"
	nodeGroupAccessApp  = "app"
)

func normalizeNodeGroupType(groupType string) string {
	switch strings.ToLower(strings.TrimSpace(groupType)) {
	case "", nodeGroupTypeCommon:
		return nodeGroupTypeCommon
	case nodeGroupTypeApp:
		return nodeGroupTypeApp
	default:
		return nodeGroupTypeCommon
	}
}

func isNodeGroupTypeAccessible(groupType, accessType string) bool {
	switch accessType {
	case nodeGroupAccessApp:
		resolved := normalizeNodeGroupType(groupType)
		return resolved == nodeGroupTypeCommon || resolved == nodeGroupTypeApp
	default:
		return false
	}
}

func legacyAccessibleNodeGroup(ctx context.Context, d *Data, nodeGroupID int64) (*ent.ProxyServerGroup, error) {
	if d == nil || d.db == nil || nodeGroupID == 0 {
		return nil, nil
	}
	item, err := d.db.ProxyServerGroup.Query().
		Where(proxyservergroup.IDEQ(nodeGroupID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if !isNodeGroupTypeAccessible(item.GroupType, nodeGroupAccessApp) {
		return nil, nil
	}
	return item, nil
}

func nodeMatchesTags(nodeTags string, tags []string) bool {
	for _, tag := range tags {
		for _, item := range strings.Split(nodeTags, ",") {
			if strings.TrimSpace(item) == tag {
				return true
			}
		}
	}
	return false
}

func buildLegacyNodeInfos(ctx context.Context, d *Data, userSub *ent.ProxyUserSubscribe, nodes []*ent.ProxyNode) ([]*subscribeBiz.UserSubscribeNodeInfo, error) {
	if len(nodes) == 0 {
		return []*subscribeBiz.UserSubscribeNodeInfo{}, nil
	}
	serverIDs := make([]int64, 0, len(nodes))
	serverSeen := make(map[int64]struct{}, len(nodes))
	for _, node := range nodes {
		if _, ok := serverSeen[node.ServerID]; ok {
			continue
		}
		serverSeen[node.ServerID] = struct{}{}
		serverIDs = append(serverIDs, node.ServerID)
	}
	servers, err := d.db.ProxyServer.Query().
		Where(proxyserver.IDIn(serverIDs...)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	serverMap := make(map[int64]*ent.ProxyServer, len(servers))
	for _, server := range servers {
		serverMap[server.ID] = server
	}
	result := make([]*subscribeBiz.UserSubscribeNodeInfo, 0, len(nodes))
	for _, node := range nodes {
		server := serverMap[node.ServerID]
		if server == nil {
			continue
		}
		result = append(result, &subscribeBiz.UserSubscribeNodeInfo{
			ID:              node.ID,
			Name:            node.Name,
			Uuid:            stringValue(userSub.UUID),
			Protocol:        node.Protocol,
			Protocols:       server.Protocol,
			Port:            uint32(node.Port),
			Address:         node.Address,
			Tags:            strings.Split(node.Tags, ","),
			Country:         server.Country,
			City:            server.City,
			Longitude:       server.Longitude,
			Latitude:        server.Latitude,
			LatitudeCenter:  server.LatitudeCenter,
			LongitudeCenter: server.LongitudeCenter,
			CreatedAt:       node.CreatedAt.Unix(),
		})
	}
	return result, nil
}
