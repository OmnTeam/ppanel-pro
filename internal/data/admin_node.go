package data

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	serverbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/server"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
)

type adminNodeRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminNodeRepo creates a new admin node repository
func NewAdminNodeRepo(data *Data, logger log.Logger) serverbiz.NodeRepo {
	return &adminNodeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateNode creates a new node
func (r *adminNodeRepo) CreateNode(ctx context.Context, node *serverbiz.Node) (*serverbiz.Node, error) {
	// Convert tags array to comma-separated string
	tagsStr := tool.StringSliceToString(node.Tags)

	builder := r.data.db.ProxyNode.Create().
		SetName(node.Name).
		SetTags(tagsStr).
		SetPort(node.Port).
		SetAddress(node.Address).
		SetServerID(node.ServerID).
		SetProtocol(node.Protocol).
		SetSort(int(node.Sort)) // uint32 to int

	if node.Enabled != nil {
		builder = builder.SetEnabled(*node.Enabled)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Convert tags string back to array and remove duplicates (match old project)
	tags := tool.RemoveDuplicateElements(tool.StringToStringSlice(created.Tags)...)

	enabled := created.Enabled

	return &serverbiz.Node{
		ID: int64(created.ID),

		Name:      created.Name,
		Tags:      tags,
		Port:      uint16(created.Port), // int to uint16
		Address:   created.Address,
		ServerID:  created.ServerID,
		Protocol:  created.Protocol,
		Enabled:   &enabled,
		Sort:      uint32(created.Sort),          // int to uint32
		CreatedAt: created.CreatedAt.UnixMilli(), // 返回毫秒时间戳
		UpdatedAt: created.UpdatedAt.UnixMilli(), // 返回毫秒时间戳
	}, nil
}

// UpdateNode updates an existing node
func (r *adminNodeRepo) UpdateNode(ctx context.Context, node *serverbiz.Node) (*serverbiz.Node, error) {
	// Convert tags array to comma-separated string
	tagsStr := tool.StringSliceToString(node.Tags)

	builder := r.data.db.ProxyNode.UpdateOneID(node.ID).
		SetName(node.Name).
		SetTags(tagsStr).
		SetPort(node.Port).
		SetAddress(node.Address).
		SetServerID(node.ServerID).
		SetProtocol(node.Protocol).
		SetSort(int(node.Sort))

	if node.Enabled != nil {
		builder = builder.SetEnabled(*node.Enabled)
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Convert tags string back to array and remove duplicates (match old project)
	tags := tool.RemoveDuplicateElements(tool.StringToStringSlice(updated.Tags)...)

	enabled := updated.Enabled

	return &serverbiz.Node{
		ID: int64(updated.ID),

		Name:      updated.Name,
		Tags:      tags,
		Port:      uint16(updated.Port), // int to uint16
		Address:   updated.Address,
		ServerID:  updated.ServerID,
		Protocol:  updated.Protocol,
		Enabled:   &enabled,
		Sort:      uint32(updated.Sort),          // int to uint32
		CreatedAt: updated.CreatedAt.UnixMilli(), // 返回毫秒时间戳
		UpdatedAt: updated.UpdatedAt.UnixMilli(), // 返回毫秒时间戳
	}, nil
}

// DeleteNode deletes a node
func (r *adminNodeRepo) DeleteNode(ctx context.Context, id int) error {
	// First query the node to get serverID for cache clearing
	node, err := r.data.db.ProxyNode.Query().
		Where(proxynode.ID(int64(id))).
		Only(ctx)
	if err != nil {
		return err
	}

	// Delete the node
	err = r.data.db.ProxyNode.DeleteOneID(int64(id)).Exec(ctx)
	if err != nil {
		return err
	}

	// Clear cache for the server
	if err := r.ClearNodeCache(ctx, []int{int(node.ServerID)}); err != nil {
		r.log.Warnf("Failed to clear node cache for server %d after deleting node %d: %v", node.ServerID, id, err)
		// Don't return error, just log warning
	}

	return nil
}

// FilterNodeList filters node list
func (r *adminNodeRepo) FilterNodeList(ctx context.Context, page, size int32, search string) (int64, []*serverbiz.Node, error) {
	query := r.data.db.ProxyNode.Query()
	if search != "" {
		searchPattern := "%" + search + "%"
		// 支持按名称、地址、标签、端口号搜索（与老项目保持一致）
		query = query.Where(func(s *sql.Selector) {
			s.Where(sql.Or(
				sql.Like(s.C(proxynode.FieldName), searchPattern),
				sql.Like(s.C(proxynode.FieldAddress), searchPattern),
				sql.Like(s.C(proxynode.FieldTags), searchPattern),
				sql.P(func(b *sql.Builder) {
					b.WriteString("CAST(")
					b.Ident(proxynode.FieldPort)
					b.WriteString(" AS CHAR) LIKE ")
					b.Arg(searchPattern)
				}), // 端口号转字符串搜索
			))
		})
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return 0, nil, err
	}

	// Get list
	list, err := query.
		Order(ent.Asc(proxynode.FieldSort)).
		Limit(int(size)).
		Offset(int((page - 1) * size)).
		All(ctx)
	if err != nil {
		return 0, nil, err
	}

	nodes := make([]*serverbiz.Node, 0, len(list))
	for _, item := range list {
		// Split tags and remove duplicates to match old project behavior
		tags := tool.RemoveDuplicateElements(tool.StringToStringSlice(item.Tags)...)
		enabled := item.Enabled

		nodes = append(nodes, &serverbiz.Node{
			ID: int64(item.ID),

			Name:      item.Name,
			Tags:      tags,
			Port:      uint16(item.Port), // int to uint16
			Address:   item.Address,
			ServerID:  item.ServerID,
			Protocol:  item.Protocol,
			Enabled:   &enabled,
			Sort:      uint32(item.Sort),          // int to uint32
			CreatedAt: item.CreatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
			UpdatedAt: item.UpdatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
		})
	}

	return int64(total), nodes, nil
}

// ToggleNodeStatus toggles node status
func (r *adminNodeRepo) ToggleNodeStatus(ctx context.Context, id int, enable *bool) (*serverbiz.Node, error) {
	if enable == nil {
		return nil, fmt.Errorf("enable parameter is required")
	}

	updated, err := r.data.db.ProxyNode.UpdateOneID(int64(id)).
		SetEnabled(*enable).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	tags := tool.RemoveDuplicateElements(tool.StringToStringSlice(updated.Tags)...)
	enabled := updated.Enabled

	return &serverbiz.Node{
		ID: int64(updated.ID),

		Name:      updated.Name,
		Tags:      tags,
		Port:      uint16(updated.Port), // int to uint16
		Address:   updated.Address,
		ServerID:  updated.ServerID,
		Protocol:  updated.Protocol,
		Enabled:   &enabled,
		Sort:      uint32(updated.Sort),          // int to uint32
		CreatedAt: updated.CreatedAt.UnixMilli(), // 返回毫秒时间戳
		UpdatedAt: updated.UpdatedAt.UnixMilli(), // 返回毫秒时间戳
	}, nil
}

// QueryNodeTags queries all unique node tags
func (r *adminNodeRepo) QueryNodeTags(ctx context.Context) ([]string, error) {
	nodes, err := r.data.db.ProxyNode.Query().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Collect all unique tags
	tagSet := make(map[string]bool)
	for _, node := range nodes {
		tags := tool.StringToStringSlice(node.Tags)
		for _, tag := range tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	// Convert to slice
	uniqueTags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		uniqueTags = append(uniqueTags, tag)
	}

	return uniqueTags, nil
}

// ResetNodeSort resets node sort order
func (r *adminNodeRepo) ResetNodeSort(ctx context.Context, sortItems []*serverbiz.SortItem) error {
	for _, item := range sortItems {
		// Update the node sort
		affected, err := r.data.db.ProxyNode.Update().
			Where(proxynode.ID(item.ID)).
			SetSort(int(item.Sort)).
			Save(ctx)
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf("node %d not found", item.ID)
		}
	}
	return nil
}

// ClearNodeCache clears node-related cache for specified servers
func (r *adminNodeRepo) ClearNodeCache(ctx context.Context, serverIDs []int) error {
	// Convert int serverIDs to []any for sql.In
	serverIDsAny := make([]any, len(serverIDs))
	for i, id := range serverIDs {
		serverIDsAny[i] = id
	}

	nodes, err := r.data.db.ProxyNode.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.In(s.C(proxynode.FieldServerID), serverIDsAny...))
		}).
		All(ctx)
	if err != nil {
		r.log.Errorf("Failed to query nodes for cache clearing: %v", err)
		return err
	}

	// Clear cache for each node
	for _, node := range nodes {
		// Clear status cache
		statusKey := fmt.Sprintf(StatusCacheKey, int64(node.ServerID))
		if err := r.data.rdb.Del(ctx, statusKey).Err(); err != nil {
			r.log.Warnf("Failed to delete status cache for server %d: %v", node.ServerID, err)
		}

		// Clear online user cache for this node's protocol
		if node.Protocol != "" {
			onlineKey := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, int64(node.ServerID), node.Protocol)
			if err := r.data.rdb.Del(ctx, onlineKey).Err(); err != nil {
				r.log.Warnf("Failed to delete online user cache for server %d protocol %s: %v", node.ServerID, node.Protocol, err)
			}
		}
	}

	r.log.Infof("Cleared cache for %d nodes across %d servers", len(nodes), len(serverIDs))
	return nil
}
