package group

import (
	"context"
	"fmt"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/group/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/go-kratos/kratos/v2/log"
)

// GroupRepo group repository interface
type GroupRepo interface {
	// Node Group CRUD
	CreateNodeGroup(ctx context.Context, req *v1.CreateNodeGroupRequest) (int64, error)
	UpdateNodeGroup(ctx context.Context, req *v1.UpdateNodeGroupRequest) error
	DeleteNodeGroup(ctx context.Context, id int64) error
	GetNodeGroupList(ctx context.Context, req *v1.GetNodeGroupListRequest) ([]*ent.ProxyServerGroup, int64, error)

	// Group Config
	GetGroupConfig(ctx context.Context) (*v1.GroupConfig, error)
	UpdateGroupConfig(ctx context.Context, req *v1.UpdateGroupConfigRequest) error

	// Group Operations
	RecalculateGroup(ctx context.Context, mode string, triggerType string) (int64, error)
	GetRecalculationStatus(ctx context.Context) (*v1.RecalculationState, error)
	GetGroupHistory(ctx context.Context, req *v1.GetGroupHistoryRequest) ([]*ent.ProxyGroupHistory, int64, error)
	GetGroupHistoryDetail(ctx context.Context, historyID int64) (*ent.ProxyGroupHistory, error)
	PreviewUserNodes(ctx context.Context, userId int64) ([]*ent.ProxyNode, int64, error)
	ResetGroups(ctx context.Context) error
	GetSubscribeGroupMapping(ctx context.Context) ([]*v1.SubscribeGroupMappingItem, error)
	ExportGroupResult(ctx context.Context, historyID *int64) ([]byte, string, error)
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
	triggerType := req.TriggerType
	if triggerType == "" {
		triggerType = "manual" // 默认为手动触发
	}

	historyId, err := uc.repo.RecalculateGroup(ctx, req.Mode, triggerType)
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

	nodes, nodeGroupId, err := uc.repo.PreviewUserNodes(ctx, req.UserId)
	if err != nil {
		uc.log.Errorf("Failed to preview user nodes: %v", err)
		return nil, 0, err
	}

	return nodes, nodeGroupId, nil
}

// GetGroupHistoryDetail gets group history detail
func (uc *GroupUseCase) GetGroupHistoryDetail(ctx context.Context, historyID int64) (*ent.ProxyGroupHistory, error) {
	if historyID <= 0 {
		return nil, nil
	}

	history, err := uc.repo.GetGroupHistoryDetail(ctx, historyID)
	if err != nil {
		uc.log.Errorf("Failed to get group history detail: %v", err)
		return nil, err
	}

	return history, nil
}

// MigrateUsers migrates users from one group to another (已废弃 - UserGroup已移除)
// 此方法保留是为了API兼容，但总是返回错误
func (uc *GroupUseCase) MigrateUsers(ctx context.Context, fromGroupID, toGroupID int64, includeLocked bool) (successCount, failedCount int32, err error) {
	uc.log.Warnf("MigrateUsers is deprecated: UserGroup has been removed")
	return 0, 0, fmt.Errorf("UserGroup feature has been removed, please use RecalculateGroup instead")
}

// ResetGroups resets all user groups
func (uc *GroupUseCase) ResetGroups(ctx context.Context) error {
	err := uc.repo.ResetGroups(ctx)
	if err != nil {
		uc.log.Errorf("Failed to reset groups: %v", err)
		return err
	}

	return nil
}

// GetSubscribeGroupMapping gets subscribe group mapping
func (uc *GroupUseCase) GetSubscribeGroupMapping(ctx context.Context) ([]*v1.SubscribeGroupMappingItem, error) {
	list, err := uc.repo.GetSubscribeGroupMapping(ctx)
	if err != nil {
		uc.log.Errorf("Failed to get subscribe group mapping: %v", err)
		return nil, err
	}

	return list, nil
}

// ExportGroupResult exports group result as CSV
func (uc *GroupUseCase) ExportGroupResult(ctx context.Context, req *v1.ExportGroupResultRequest) ([]byte, string, error) {
	data, filename, err := uc.repo.ExportGroupResult(ctx, &req.HistoryId)
	if err != nil {
		uc.log.Errorf("Failed to export group result: %v", err)
		return nil, "", err
	}

	return data, filename, nil
}
