package data

import (
	"context"
	"strings"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxynode"
	"github.com/OmnTeam/npanel-pro/ent/proxyservergroup"
	subscriptionbiz "github.com/OmnTeam/npanel-pro/internal/biz/public/subscription"
	"github.com/OmnTeam/npanel-pro/pkg/tool"
)

// getNodesByGroup 在开启节点分组时取节点：
// 1. 订阅直接绑定的节点始终保留；
// 2. 再补充分组内可用于订阅输出的节点；
// 3. 两类节点都必须满足“已启用且未隐藏”。
func (r *publicSubscriptionRepo) getNodesByGroup(ctx context.Context, userSubscribe *subscriptionbiz.UserSubscribe, subscribePlan *ent.ProxySubscribe) ([]*ent.ProxyNode, error) {
	directNodeIDs := sanitizeSubscribeNodeIDs(subscribePlan)
	resultNodes, nodeIDMap, err := r.queryVisibleNodesByIDs(ctx, directNodeIDs)
	if err != nil {
		return nil, err
	}

	nodeGroupID, source := resolveSubscriptionNodeGroupID(userSubscribe, subscribePlan)
	r.log.Infof("Using %s: %d", source, nodeGroupID)

	var currentNodeGroup *ent.ProxyServerGroup
	if nodeGroupID > 0 {
		accessibleGroup, err := r.getAccessibleNodeGroupForSubscribe(ctx, nodeGroupID)
		if err != nil {
			return nil, err
		}
		if accessibleGroup == nil {
			r.log.Infof("Subscribe node group %d from %s is not accessible, keep direct subscribe nodes only", nodeGroupID, source)
			nodeGroupID = 0
		} else {
			currentNodeGroup = accessibleGroup
		}
	}

	if nodeGroupID > 0 {
		groupNodes, err := r.queryVisibleNodesByGroupID(ctx, nodeGroupID)
		if err != nil {
			return nil, err
		}
		for _, node := range groupNodes {
			if !nodeIDMap[node.ID] {
				resultNodes = append(resultNodes, node)
				nodeIDMap[node.ID] = true
			}
		}
	}

	r.log.Infof("Found %d nodes (group=%d direct=%d)", len(resultNodes), nodeGroupID, len(directNodeIDs))

	// 为分组节点补一个 tag，兼容旧订阅模板里对 tag 的使用。
	if currentNodeGroup != nil && currentNodeGroup.Name != "" {
		for _, node := range resultNodes {
			if node.Tags == "" && containsInt64(node.NodeGroupIds, currentNodeGroup.ID) {
				node.Tags = currentNodeGroup.Name
				r.log.Debugf("Set node_group name as tag for node %d: %s", node.ID, currentNodeGroup.Name)
			}
		}
	}

	return resultNodes, nil
}

func resolveSubscriptionNodeGroupID(userSubscribe *subscriptionbiz.UserSubscribe, subscribePlan *ent.ProxySubscribe) (int64, string) {
	if userSubscribe != nil && userSubscribe.NodeGroupID != 0 {
		return userSubscribe.NodeGroupID, "user_subscribe.node_group_id"
	}
	if subscribePlan != nil && subscribePlan.NodeGroupID != nil && *subscribePlan.NodeGroupID > 0 {
		return *subscribePlan.NodeGroupID, "subscribe.node_group_id"
	}
	if subscribePlan != nil && len(subscribePlan.NodeGroupIds) > 0 {
		return subscribePlan.NodeGroupIds[0], "subscribe.node_group_ids[0]"
	}
	return 0, ""
}

func resolveNodeGroupID(nodeGroupIDs []int64, preferred int64) int64 {
	if preferred > 0 {
		for _, gid := range nodeGroupIDs {
			if gid == preferred {
				return gid
			}
		}
	}
	if len(nodeGroupIDs) > 0 {
		return nodeGroupIDs[0]
	}
	return 0
}

// getNodesByTag 在未开启分组时取节点：
// 1. 直接绑定节点与 tag 匹配节点做并集；
// 2. 直接绑定节点不受 tag 结果影响；
// 3. 所有节点都必须满足“已启用且未隐藏”。
func (r *publicSubscriptionRepo) getNodesByTag(ctx context.Context, subscribePlan *ent.ProxySubscribe) ([]*ent.ProxyNode, error) {
	nodeIDs := sanitizeSubscribeNodeIDs(subscribePlan)
	tags := sanitizeSubscribeTags(subscribePlan)

	r.log.Infof("Subscribe nodes: raw=%v valid=%d, tags=%v", subscribePlan.Nodes, len(nodeIDs), tags)

	if len(nodeIDs) == 0 && len(tags) == 0 {
		return []*ent.ProxyNode{}, nil
	}

	resultNodes, nodeIDMap, err := r.queryVisibleNodesByIDs(ctx, nodeIDs)
	if err != nil {
		return nil, err
	}

	if len(tags) > 0 {
		tagNodes, err := r.queryVisibleNodesByTags(ctx, tags)
		if err != nil {
			return nil, err
		}
		for _, node := range tagNodes {
			if !nodeIDMap[node.ID] {
				resultNodes = append(resultNodes, node)
				nodeIDMap[node.ID] = true
			}
		}
	}

	return resultNodes, nil
}

const (
	subscriptionNodeGroupTypeCommon      = "common"
	subscriptionNodeGroupTypeSubscribe   = "subscribe"
	subscriptionNodeGroupAccessSubscribe = "subscribe"
)

func normalizeSubscriptionNodeGroupType(groupType string) string {
	switch strings.ToLower(strings.TrimSpace(groupType)) {
	case "", subscriptionNodeGroupTypeCommon:
		return subscriptionNodeGroupTypeCommon
	case subscriptionNodeGroupTypeSubscribe:
		return subscriptionNodeGroupTypeSubscribe
	default:
		return strings.ToLower(strings.TrimSpace(groupType))
	}
}

func isSubscriptionNodeGroupTypeAccessible(groupType, accessType string) bool {
	switch accessType {
	case subscriptionNodeGroupAccessSubscribe:
		resolved := normalizeSubscriptionNodeGroupType(groupType)
		return resolved == subscriptionNodeGroupTypeCommon || resolved == subscriptionNodeGroupTypeSubscribe
	default:
		return false
	}
}

func (r *publicSubscriptionRepo) getAccessibleNodeGroupForSubscribe(ctx context.Context, nodeGroupID int64) (*ent.ProxyServerGroup, error) {
	if nodeGroupID == 0 {
		return nil, nil
	}

	nodeGroup, err := r.data.db.ProxyServerGroup.Query().
		Where(
			proxyservergroup.IDEQ(nodeGroupID),
			proxyservergroup.IsExpiredGroupEQ(false),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			r.log.Debugf("Subscribe node group %d not found", nodeGroupID)
			return nil, nil
		}
		r.log.Errorf("Failed to query subscribe node group %d: %v", nodeGroupID, err)
		return nil, err
	}

	if !isSubscriptionNodeGroupTypeAccessible(nodeGroup.GroupType, subscriptionNodeGroupAccessSubscribe) {
		r.log.Infof("Subscribe node group %d is not accessible for subscribe output, type=%s", nodeGroupID, nodeGroup.GroupType)
		return nil, nil
	}

	return nodeGroup, nil
}

func sanitizeSubscribeNodeIDs(subscribePlan *ent.ProxySubscribe) []int64 {
	if subscribePlan == nil {
		return nil
	}
	rawNodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)
	nodeIDs := make([]int64, 0, len(rawNodeIDs))
	seen := make(map[int64]struct{}, len(rawNodeIDs))
	for _, id := range rawNodeIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		nodeIDs = append(nodeIDs, id)
	}
	return nodeIDs
}

func sanitizeSubscribeTags(subscribePlan *ent.ProxySubscribe) []string {
	if subscribePlan == nil {
		return nil
	}
	tags := tool.StringToStringSlice(subscribePlan.NodeTags)
	cleaned := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		cleaned = append(cleaned, tag)
	}
	return cleaned
}

func (r *publicSubscriptionRepo) queryVisibleNodesByIDs(ctx context.Context, nodeIDs []int64) ([]*ent.ProxyNode, map[int64]bool, error) {
	result := make([]*ent.ProxyNode, 0)
	seen := make(map[int64]bool)
	if len(nodeIDs) == 0 {
		return result, seen, nil
	}

	nodes, err := r.data.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
			proxynode.IDIn(nodeIDs...),
		).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		r.log.Errorf("Failed to query visible nodes by ids: %v", err)
		return nil, nil, err
	}

	for _, node := range nodes {
		if !seen[node.ID] {
			result = append(result, node)
			seen[node.ID] = true
		}
	}

	return result, seen, nil
}

func (r *publicSubscriptionRepo) queryVisibleNodesByGroupID(ctx context.Context, nodeGroupID int64) ([]*ent.ProxyNode, error) {
	if nodeGroupID <= 0 {
		return []*ent.ProxyNode{}, nil
	}

	// Ent 当前对 JSON 数组包含查询不够方便，这里先查出可见节点再做一次内存过滤。
	nodes, err := r.data.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
		).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		r.log.Errorf("Failed to query visible nodes for group %d: %v", nodeGroupID, err)
		return nil, err
	}

	result := make([]*ent.ProxyNode, 0, len(nodes))
	for _, node := range nodes {
		if containsInt64(node.NodeGroupIds, nodeGroupID) {
			result = append(result, node)
		}
	}
	return result, nil
}

func (r *publicSubscriptionRepo) queryVisibleNodesByTags(ctx context.Context, tags []string) ([]*ent.ProxyNode, error) {
	if len(tags) == 0 {
		return []*ent.ProxyNode{}, nil
	}

	nodes, err := r.data.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
		).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		r.log.Errorf("Failed to query visible nodes by tags: %v", err)
		return nil, err
	}

	result := make([]*ent.ProxyNode, 0, len(nodes))
	for _, node := range nodes {
		if nodeMatchesTags(node.Tags, tags) {
			result = append(result, node)
		}
	}
	return result, nil
}

func containsInt64(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
