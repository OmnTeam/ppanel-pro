package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// NodeUsecase is the node use case
type NodeUsecase struct {
	repo NodeRepo
	log  *log.Helper
}

// NewNodeUsecase creates a new node use case
func NewNodeUsecase(repo NodeRepo, logger log.Logger) *NodeUsecase {
	return &NodeUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateNode creates a new node
func (uc *NodeUsecase) CreateNode(ctx context.Context, name string, tags []string, port uint16, address string, serverID int64, protocol string, enabled *bool) (*Node, error) {
	node := &Node{
		Name:     name,
		Tags:     tags,
		Port:     port,
		Address:  address,
		ServerID: serverID,
		Protocol: protocol,
		Enabled:  enabled,
	}
	return uc.repo.CreateNode(ctx, node)
}

// UpdateNode updates an existing node
func (uc *NodeUsecase) UpdateNode(ctx context.Context, id int64, name string, tags []string, port uint16, address string, serverID int64, protocol string, enabled *bool) (*Node, error) {
	node := &Node{
		ID:       id,
		Name:     name,
		Tags:     tags,
		Port:     port,
		Address:  address,
		ServerID: serverID,
		Protocol: protocol,
		Enabled:  enabled,
	}

	updatedNode, err := uc.repo.UpdateNode(ctx, node)
	if err != nil {
		return nil, err
	}

	// Clear node cache for the server
	if err := uc.repo.ClearNodeCache(ctx, []int64{serverID}); err != nil {
		uc.log.Warnf("Failed to clear node cache for server %d: %v", serverID, err)
		// Don't return error, just log warning
	}

	return updatedNode, nil
}

// DeleteNode deletes a node
func (uc *NodeUsecase) DeleteNode(ctx context.Context, id int64) error {
	return uc.repo.DeleteNode(ctx, id)
}

// FilterNodeList filters node list
func (uc *NodeUsecase) FilterNodeList(ctx context.Context, page, size int32, search string) (int64, []*Node, error) {
	return uc.repo.FilterNodeList(ctx, page, size, search)
}

// ToggleNodeStatus toggles node status
func (uc *NodeUsecase) ToggleNodeStatus(ctx context.Context, id int64, enable *bool) (*Node, error) {
	node, err := uc.repo.ToggleNodeStatus(ctx, id, enable)
	if err != nil {
		return nil, err
	}

	// Clear node cache for the server
	if err := uc.repo.ClearNodeCache(ctx, []int64{node.ServerID}); err != nil {
		uc.log.Warnf("Failed to clear node cache for server %d after toggling node %d: %v", node.ServerID, id, err)
		// Don't return error, just log warning
	}

	return node, nil
}

// QueryNodeTags queries all node tags
func (uc *NodeUsecase) QueryNodeTags(ctx context.Context) ([]string, error) {
	return uc.repo.QueryNodeTags(ctx)
}

// ResetNodeSort resets node sort
func (uc *NodeUsecase) ResetNodeSort(ctx context.Context, sortItems []*SortItem) error {
	return uc.repo.ResetNodeSort(ctx, sortItems)
}
