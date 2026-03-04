package group

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/group/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/go-kratos/kratos/v2/log"
)

// GroupRepo group repository interface
type GroupRepo interface {
	// User Group CRUD
	CreateUserGroup(ctx context.Context, req *v1.CreateUserGroupRequest) (int64, error)
	UpdateUserGroup(ctx context.Context, req *v1.UpdateUserGroupRequest) error
	DeleteUserGroup(ctx context.Context, id int64) error
	GetUserGroupList(ctx context.Context, req *v1.GetUserGroupListRequest) ([]*ent.ProxyUserGroup, int64, error)
	UpdateUserUserGroup(ctx context.Context, req *v1.UpdateUserUserGroupRequest) error

	// Node Group CRUD
	CreateNodeGroup(ctx context.Context, req *v1.CreateNodeGroupRequest) (int64, error)
	UpdateNodeGroup(ctx context.Context, req *v1.UpdateNodeGroupRequest) error
	DeleteNodeGroup(ctx context.Context, id int64) error
	GetNodeGroupList(ctx context.Context, req *v1.GetNodeGroupListRequest) ([]*ent.ProxyServerGroup, int64, error)

	// Group Config
	GetGroupConfig(ctx context.Context) (*v1.GroupConfig, error)
	UpdateGroupConfig(ctx context.Context, req *v1.UpdateGroupConfigRequest) error

	// Group Operations (simplified)
	RecalculateGroup(ctx context.Context, mode string) (int64, error)
	GetRecalculationStatus(ctx context.Context) (*v1.RecalculationState, error)
	GetGroupHistory(ctx context.Context, req *v1.GetGroupHistoryRequest) ([]*ent.ProxyGroupHistory, int64, error)
	PreviewUserNodes(ctx context.Context, userId int64) ([]*ent.ProxyNode, int64, error)
}

// GroupUseCase group use case
type GroupUseCase struct {
	repo GroupRepo
	log  *log.Helper
}

// NewGroupUseCase creates a new group use case
func NewGroupUseCase(repo GroupRepo, logger log.Logger) *GroupUseCase {
	return &GroupUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/group")),
	}
}

// CreateUserGroup creates user group
func (uc *GroupUseCase) CreateUserGroup(ctx context.Context, req *v1.CreateUserGroupRequest) (int64, error) {
	if req.Name == "" {
		return 0, nil
	}

	id, err := uc.repo.CreateUserGroup(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to create user group: %v", err)
		return 0, err
	}

	return id, nil
}

// UpdateUserGroup updates user group
func (uc *GroupUseCase) UpdateUserGroup(ctx context.Context, req *v1.UpdateUserGroupRequest) error {
	if req.Id <= 0 {
		return nil
	}

	err := uc.repo.UpdateUserGroup(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to update user group: %v", err)
		return err
	}

	return nil
}

// DeleteUserGroup deletes user group
func (uc *GroupUseCase) DeleteUserGroup(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}

	err := uc.repo.DeleteUserGroup(ctx, id)
	if err != nil {
		uc.log.Errorf("Failed to delete user group: %v", err)
		return err
	}

	return nil
}

// GetUserGroupList gets user group list
func (uc *GroupUseCase) GetUserGroupList(ctx context.Context, req *v1.GetUserGroupListRequest) ([]*ent.ProxyUserGroup, int64, error) {
	if req.Page <= 0 || req.Size <= 0 {
		return nil, 0, nil
	}

	list, total, err := uc.repo.GetUserGroupList(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to get user group list: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateUserUserGroup updates user's user group
func (uc *GroupUseCase) UpdateUserUserGroup(ctx context.Context, req *v1.UpdateUserUserGroupRequest) error {
	err := uc.repo.UpdateUserUserGroup(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to update user user group: %v", err)
		return err
	}

	return nil
}

// CreateNodeGroup creates node group
func (uc *GroupUseCase) CreateNodeGroup(ctx context.Context, req *v1.CreateNodeGroupRequest) (int64, error) {
	if req.Name == "" {
		return 0, nil
	}

	id, err := uc.repo.CreateNodeGroup(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to create node group: %v", err)
		return 0, err
	}

	return id, nil
}

// UpdateNodeGroup updates node group
func (uc *GroupUseCase) UpdateNodeGroup(ctx context.Context, req *v1.UpdateNodeGroupRequest) error {
	if req.Id <= 0 {
		return nil
	}

	err := uc.repo.UpdateNodeGroup(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to update node group: %v", err)
		return err
	}

	return nil
}

// DeleteNodeGroup deletes node group
func (uc *GroupUseCase) DeleteNodeGroup(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}

	err := uc.repo.DeleteNodeGroup(ctx, id)
	if err != nil {
		uc.log.Errorf("Failed to delete node group: %v", err)
		return err
	}

	return nil
}

// GetNodeGroupList gets node group list
func (uc *GroupUseCase) GetNodeGroupList(ctx context.Context, req *v1.GetNodeGroupListRequest) ([]*ent.ProxyServerGroup, int64, error) {
	if req.Page <= 0 || req.Size <= 0 {
		return nil, 0, nil
	}

	list, total, err := uc.repo.GetNodeGroupList(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to get node group list: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

// GetGroupConfig gets group config
func (uc *GroupUseCase) GetGroupConfig(ctx context.Context) (*v1.GroupConfig, *v1.RecalculationState, error) {
	config, err := uc.repo.GetGroupConfig(ctx)
	if err != nil {
		uc.log.Errorf("Failed to get group config: %v", err)
		return nil, nil, err
	}

	// Get recalculation status
	state, err := uc.repo.GetRecalculationStatus(ctx)
	if err != nil {
		uc.log.Warnf("Failed to get recalculation status: %v", err)
		state = &v1.RecalculationState{
			State:    "idle",
			Progress: 0,
			Total:    0,
		}
	}

	return config, state, nil
}

// UpdateGroupConfig updates group config
func (uc *GroupUseCase) UpdateGroupConfig(ctx context.Context, req *v1.UpdateGroupConfigRequest) error {
	err := uc.repo.UpdateGroupConfig(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to update group config: %v", err)
		return err
	}

	return nil
}

// RecalculateGroup recalculates groups
func (uc *GroupUseCase) RecalculateGroup(ctx context.Context, req *v1.RecalculateGroupRequest) (int64, error) {
	historyId, err := uc.repo.RecalculateGroup(ctx, req.Mode)
	if err != nil {
		uc.log.Errorf("Failed to recalculate group: %v", err)
		return 0, err
	}

	return historyId, nil
}

// GetRecalculationStatus gets recalculation status
func (uc *GroupUseCase) GetRecalculationStatus(ctx context.Context) (*v1.RecalculationState, error) {
	state, err := uc.repo.GetRecalculationStatus(ctx)
	if err != nil {
		uc.log.Errorf("Failed to get recalculation status: %v", err)
		return nil, err
	}

	return state, nil
}

// GetGroupHistory gets group history
func (uc *GroupUseCase) GetGroupHistory(ctx context.Context, req *v1.GetGroupHistoryRequest) ([]*ent.ProxyGroupHistory, int64, error) {
	if req.Page <= 0 || req.Size <= 0 {
		return nil, 0, nil
	}

	list, total, err := uc.repo.GetGroupHistory(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to get group history: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

// PreviewUserNodes previews user nodes
func (uc *GroupUseCase) PreviewUserNodes(ctx context.Context, req *v1.PreviewUserNodesRequest) ([]*ent.ProxyNode, int64, error) {
	if req.UserId <= 0 {
		return nil, 0, nil
	}

	nodes, userGroupId, err := uc.repo.PreviewUserNodes(ctx, req.UserId)
	if err != nil {
		uc.log.Errorf("Failed to preview user nodes: %v", err)
		return nil, 0, err
	}

	return nodes, userGroupId, nil
}
