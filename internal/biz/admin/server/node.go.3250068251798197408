package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type NodeUsecase struct {
	repo NodeRepo
	log  *log.Helper
}

func NewNodeUsecase(repo NodeRepo, logger log.Logger) *NodeUsecase {
	return &NodeUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *NodeUsecase) CreateNode(ctx context.Context, name string, tags []string, port uint16, address string, serverID int64, protocol string, enabled *bool, nodeType string, isHidden *bool, nodeGroupIDs []int64) (*Node, error) {
	return uc.repo.CreateNode(ctx, &Node{
		Name:         name,
		Tags:         tags,
		Port:         port,
		Address:      address,
		ServerID:     serverID,
		Protocol:     protocol,
		Enabled:      enabled,
		NodeType:     nodeType,
		IsHidden:     isHidden,
		NodeGroupIDs: nodeGroupIDs,
	})
}

func (uc *NodeUsecase) UpdateNode(ctx context.Context, id int, name string, tags []string, port uint16, address string, serverID int64, protocol string, enabled *bool, nodeType string, isHidden *bool, nodeGroupIDs []int64) (*Node, error) {
	node := &Node{
		ID:           int64(id),
		Name:         name,
		Tags:         tags,
		Port:         port,
		Address:      address,
		ServerID:     serverID,
		Protocol:     protocol,
		Enabled:      enabled,
		NodeType:     nodeType,
		IsHidden:     isHidden,
		NodeGroupIDs: nodeGroupIDs,
	}
	updatedNode, err := uc.repo.UpdateNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.ClearNodeCache(ctx, []int{int(serverID)}); err != nil {
		uc.log.Warnf("Failed to clear node cache for server %d: %v", serverID, err)
	}
	return updatedNode, nil
}

func (uc *NodeUsecase) DeleteNode(ctx context.Context, id int) error {
	return uc.repo.DeleteNode(ctx, id)
}

func (uc *NodeUsecase) FilterNodeList(ctx context.Context, page, size int32, search string, nodeGroupID *int64) (int32, []*Node, error) {
	return uc.repo.FilterNodeList(ctx, page, size, search, nodeGroupID)
}

func (uc *NodeUsecase) ToggleNodeStatus(ctx context.Context, id int, enable *bool) (*Node, error) {
	node, err := uc.repo.ToggleNodeStatus(ctx, id, enable)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.ClearNodeCache(ctx, []int{int(node.ServerID)}); err != nil {
		uc.log.Warnf("Failed to clear node cache for server %d after toggling node %d: %v", node.ServerID, id, err)
	}
	return node, nil
}

func (uc *NodeUsecase) QueryNodeTags(ctx context.Context) ([]string, error) {
	return uc.repo.QueryNodeTags(ctx)
}
func (uc *NodeUsecase) ResetNodeSort(ctx context.Context, sortItems []*SortItem) error {
	return uc.repo.ResetNodeSort(ctx, sortItems)
}
