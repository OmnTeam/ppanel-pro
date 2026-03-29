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

	// 查询所有启用的节点
	allNodes, err := r.data.db.ProxyNode.Query().
		Where(proxynode.EnabledEQ(true)).
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
	if nodeGroupId > 0 {
		// 查询节点组信息
		nodeGroup, err := r.data.db.ProxyServerGroup.Get(ctx, nodeGroupId)
		if err == nil && nodeGroup != nil && nodeGroup.Name != "" {
			for _, n := range resultNodes {
				// 只为分组节点设置 tag，公共节点不设置
				// 并且只在节点没有 tag 时设置
				if n.Tags == "" && n.NodeGroupIds != nil && len(n.NodeGroupIds) > 0 {
					n.Tags = nodeGroup.Name
					r.log.Debugf("Set node_group name as tag for node %d: %s", n.ID, nodeGroup.Name)
				}
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
	nodeIdsStr := subscribePlan.Nodes
	tags := tool.StringToStringSlice(subscribePlan.NodeTags)
	tags = tool.RemoveDuplicateElements(tags...)

	r.log.Infof("Subscribe nodes: %v, tags: %v", nodeIdsStr, len(tags))

	// 如果没有指定节点和标签，返回空列表
	if nodeIdsStr == "" && len(tags) == 0 {
		return []*ent.ProxyNode{}, nil
	}

	// 构建节点查询
	query := r.data.db.ProxyNode.Query().
		Where(proxynode.EnabledEQ(true))

	// 解析节点ID并添加到查询条件
	if nodeIdsStr != "" {
		nodeIds := tool.StringToInt64Slice(nodeIdsStr)
		if len(nodeIds) > 0 {
			anyNodeIds := make([]any, len(nodeIds))
			for i, id := range nodeIds {
				anyNodeIds[i] = id
			}
			query = query.Where(func(s *sql.Selector) {
				s.Where(sql.In(proxynode.FieldID, anyNodeIds...))
			})
		}
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
