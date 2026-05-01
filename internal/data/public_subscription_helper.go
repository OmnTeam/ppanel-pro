package data

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	subscriptionbiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscription"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

// getNodesByGroup 按照原项目逻辑获取分组节点
func (r *publicSubscriptionRepo) getNodesByGroup(ctx context.Context, userSubscribe *subscriptionbiz.UserSubscribe, subscribePlan *ent.ProxySubscribe) ([]*ent.ProxyNode, error) {
	nodeGroupId, source := resolveSubscriptionNodeGroupID(userSubscribe, subscribePlan)
	r.log.Infof("Using %s: %d", source, nodeGroupId)

	var currentNodeGroup *ent.ProxyServerGroup
	if nodeGroupId > 0 {
		accessibleGroup, err := r.getAccessibleNodeGroupForSubscribe(ctx, nodeGroupId)
		if err != nil {
			return nil, err
		}
		if accessibleGroup == nil {
			r.log.Infof("Subscribe node group %d from %s is not accessible, fallback to public nodes", nodeGroupId, source)
			nodeGroupId = 0
		} else {
			currentNodeGroup = accessibleGroup
		}
	}

	// 查询所有启用且非隐藏的节点
	allNodes, err := r.data.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
		).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)

	if err != nil {
		r.log.Errorf("Failed to query nodes: %v", err)
		return nil, err
	}

	// 过滤节点
	var resultNodes []*ent.ProxyNode
	nodeIdMap := make(map[int64]bool)

	for _, n := range allNodes {
		// 1. 公共节点（node_group_ids 为空），所有人可见
		if n.NodeGroupIds == nil || len(n.NodeGroupIds) == 0 {
			if !nodeIdMap[n.ID] {
				resultNodes = append(resultNodes, n)
				nodeIdMap[n.ID] = true
			}
			continue
		}

		// 2. 如果有节点组，检查节点是否属于该节点组
		if nodeGroupId != 0 {
			for _, gid := range n.NodeGroupIds {
				if gid == nodeGroupId {
					if !nodeIdMap[n.ID] {
						resultNodes = append(resultNodes, n)
						nodeIdMap[n.ID] = true
					}
					break
				}
			}
		}
	}

	// 旧项目在分组模式下仍然会补上 subscribe.nodes 里直接绑定的节点。
	directNodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)
	if len(directNodeIDs) > 0 {
		for _, n := range allNodes {
			if tool.Contains(directNodeIDs, n.ID) && !nodeIdMap[n.ID] {
				resultNodes = append(resultNodes, n)
				nodeIdMap[n.ID] = true
			}
		}
	}

	r.log.Infof("Found %d nodes (group=%d)", len(resultNodes), nodeGroupId)

	// 4. 为分组节点设置节点组名称为 tag（如果节点没有 tag）
	if currentNodeGroup != nil && currentNodeGroup.Name != "" {
		for _, n := range resultNodes {
			// 只为分组节点设置 tag，公共节点不设置
			// 并且只在节点没有 tag 时设置
			if n.Tags == "" && n.NodeGroupIds != nil && len(n.NodeGroupIds) > 0 {
				n.Tags = currentNodeGroup.Name
				r.log.Debugf("Set node_group name as tag for node %d: %s", n.ID, currentNodeGroup.Name)
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

// getNodesByTag 按照标签和节点ID获取节点
func (r *publicSubscriptionRepo) getNodesByTag(ctx context.Context, subscribePlan *ent.ProxySubscribe) ([]*ent.ProxyNode, error) {
	// 解析节点ID和标签
	rawNodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)
	nodeIDs := make([]int64, 0, len(rawNodeIDs))
	for _, id := range rawNodeIDs {
		if id > 0 {
			nodeIDs = append(nodeIDs, id)
		}
	}
	tags := tool.StringToStringSlice(subscribePlan.NodeTags)
	tags = tool.RemoveDuplicateElements(tags...)

	r.log.Infof("Subscribe nodes: raw=%v valid=%d, tags: %v", subscribePlan.Nodes, len(nodeIDs), len(tags))

	// 如果没有指定有效节点和标签，返回空列表
	if len(nodeIDs) == 0 && len(tags) == 0 {
		return []*ent.ProxyNode{}, nil
	}

	// 构建节点查询
	query := r.data.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
			proxynode.IsHiddenEQ(false),
		)

	// 解析节点ID并添加到查询条件
	if len(nodeIDs) > 0 {
		anyNodeIds := make([]any, len(nodeIDs))
		for i, id := range nodeIDs {
			anyNodeIds[i] = id
		}
		query = query.Where(func(s *sql.Selector) {
			s.Where(sql.In(proxynode.FieldID, anyNodeIds...))
		})
	}

	// 根据标签过滤
	if len(tags) > 0 {
		query = query.Where(func(s *sql.Selector) {
			conditions := make([]*sql.Predicate, 0, len(tags))
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					continue
				}
				conditions = append(conditions, sql.ExprP("FIND_IN_SET(?, "+proxynode.FieldTags+")", tag))
			}
			if len(conditions) > 0 {
				s.Where(sql.Or(conditions...))
			}
		})
	}

	// 查询节点
	nodes, err := query.
		Order(ent.Asc(proxynode.FieldSort)).
		Limit(1000).
		All(ctx)

	if err != nil {
		r.log.Errorf("Failed to query nodes: %v", err)
		return nil, err
	}

	return nodes, nil
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
		return subscriptionNodeGroupTypeCommon
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

	nodeGroup, err := r.data.db.ProxyServerGroup.Get(ctx, nodeGroupID)
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
