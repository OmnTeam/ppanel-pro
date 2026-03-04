package data

import (
	"context"
	"time"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/group/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxygrouphistory"
	"github.com/OmnTeam/ppanel-pro/ent/proxyservergroup"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusergroup"
	groupbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/group"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type adminGroupRepo struct {
	data   *Data
	logger *log.Helper
}

// NewAdminGroupRepo creates a new admin group repository
func NewAdminGroupRepo(d *Data, logger log.Logger) groupbiz.GroupRepo {
	return &adminGroupRepo{
		data:   d,
		logger: log.NewHelper(logger),
	}
}

// CreateUserGroup creates user group
func (r *adminGroupRepo) CreateUserGroup(ctx context.Context, req *v1.CreateUserGroupRequest) (int64, error) {
	group, err := r.data.db.ProxyUserGroup.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetSort(int(req.Sort)).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		r.logger.Errorf("Failed to create user group: %v", err)
		return 0, err
	}

	return group.ID, nil
}

// UpdateUserGroup updates user group
func (r *adminGroupRepo) UpdateUserGroup(ctx context.Context, req *v1.UpdateUserGroupRequest) error {
	group, err := r.data.db.ProxyUserGroup.Query().
		Where(proxyusergroup.IDEQ(req.Id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrUserGroupNotFound)
		}
		return err
	}

	update := group.Update()

	if req.Name != "" {
		update.SetName(req.Name)
	}
	if req.Description != "" {
		update.SetDescription(req.Description)
	}
	update.SetSort(int(req.Sort)).
		SetUpdatedAt(time.Now())

	return update.Exec(ctx)
}

// DeleteUserGroup deletes user group
func (r *adminGroupRepo) DeleteUserGroup(ctx context.Context, id int64) error {
	deletedCount, err := r.data.db.ProxyUserGroup.Delete().
		Where(proxyusergroup.IDEQ(id)).
		Exec(ctx)

	if err != nil {
		return err
	}

	if deletedCount == 0 {
		return responsecode.NewKratosError(responsecode.ErrUserGroupNotFound)
	}

	return nil
}

// GetUserGroupList gets user group list
func (r *adminGroupRepo) GetUserGroupList(ctx context.Context, req *v1.GetUserGroupListRequest) ([]*ent.ProxyUserGroup, int64, error) {
	query := r.data.db.ProxyUserGroup.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	list, err := query.
		Order(ent.Asc(proxyusergroup.FieldSort)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// UpdateUserUserGroup updates user's user group
func (r *adminGroupRepo) UpdateUserUserGroup(ctx context.Context, req *v1.UpdateUserUserGroupRequest) error {
	// 验证用户存在
	user, err := r.data.db.ProxyUser.Query().
		Where(proxyuser.IDEQ(req.UserId)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Errorf("User not found: user_id=%d", req.UserId)
			return responsecode.NewKratosError(responsecode.ErrUserNotFound)
		}
		return err
	}

	// 更新用户分组
	update := user.Update()

	// 设置group_id（只支持单个用户组）
	if len(req.GroupIds) > 0 {
		groupID := req.GroupIds[0] // 取第一个分组ID
		update.SetNillableGroupID(&groupID)
	} else {
		// 清空分组
		update.ClearGroupID()
	}

	// 设置group_locked
	update.SetGroupLocked(req.GroupLocked)

	err = update.Exec(ctx)
	if err != nil {
		r.logger.Errorf("Failed to update user group: user_id=%d, error=%v", req.UserId, err)
		return err
	}

	r.logger.Infof("Updated user user group: user_id=%d, group_ids=%v, group_locked=%v",
		req.UserId, req.GroupIds, req.GroupLocked)

	return nil
}

// CreateNodeGroup creates node group
func (r *adminGroupRepo) CreateNodeGroup(ctx context.Context, req *v1.CreateNodeGroupRequest) (int64, error) {
	group, err := r.data.db.ProxyServerGroup.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetSort(int(req.Sort)).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		r.logger.Errorf("Failed to create node group: %v", err)
		return 0, err
	}

	return group.ID, nil
}

// UpdateNodeGroup updates node group
func (r *adminGroupRepo) UpdateNodeGroup(ctx context.Context, req *v1.UpdateNodeGroupRequest) error {
	group, err := r.data.db.ProxyServerGroup.Query().
		Where(proxyservergroup.IDEQ(req.Id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrServerGroupNotFound)
		}
		return err
	}

	update := group.Update()

	if req.Name != "" {
		update.SetName(req.Name)
	}
	if req.Description != "" {
		update.SetDescription(req.Description)
	}
	update.SetSort(int(req.Sort)).
		SetUpdatedAt(time.Now())

	return update.Exec(ctx)
}

// DeleteNodeGroup deletes node group
func (r *adminGroupRepo) DeleteNodeGroup(ctx context.Context, id int64) error {
	deletedCount, err := r.data.db.ProxyServerGroup.Delete().
		Where(proxyservergroup.IDEQ(id)).
		Exec(ctx)

	if err != nil {
		return err
	}

	if deletedCount == 0 {
		return responsecode.NewKratosError(responsecode.ErrServerGroupNotFound)
	}

	return nil
}

// GetNodeGroupList gets node group list
func (r *adminGroupRepo) GetNodeGroupList(ctx context.Context, req *v1.GetNodeGroupListRequest) ([]*ent.ProxyServerGroup, int64, error) {
	query := r.data.db.ProxyServerGroup.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	list, err := query.
		Order(ent.Asc(proxyservergroup.FieldSort)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// GetGroupConfig gets group config
func (r *adminGroupRepo) GetGroupConfig(ctx context.Context) (*v1.GroupConfig, error) {
	// TODO: Implement actual config storage and retrieval
	// For now, return default config
	config := &v1.GroupConfig{
		Enabled: false,
		Mode:    "user",
		Config:  "{}", // Empty JSON object
	}

	return config, nil
}

// UpdateGroupConfig updates group config
func (r *adminGroupRepo) UpdateGroupConfig(ctx context.Context, req *v1.UpdateGroupConfigRequest) error {
	// TODO: Implement actual config storage
	r.logger.Infof("Update group config: enabled=%v, mode=%s", req.Enabled, req.Mode)
	return nil
}

// RecalculateGroup recalculates groups
func (r *adminGroupRepo) RecalculateGroup(ctx context.Context, mode string) (int64, error) {
	// Create a history record
	history, err := r.data.db.ProxyGroupHistory.Create().
		SetGroupMode(mode).
		SetTriggerType("manual").
		SetStatus("running").
		SetProgress(0).
		SetTotal(0).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return 0, err
	}

	// TODO: Implement actual recalculation logic
	// This is a complex operation that would need to:
	// 1. Calculate which users/nodes belong to which groups
	// 2. Update the group_id fields in the database
	// 3. Update the history record with results

	r.logger.Infof("Group recalculation started: history_id=%d, mode=%s", history.ID, mode)

	return history.ID, nil
}

// GetRecalculationStatus gets recalculation status
func (r *adminGroupRepo) GetRecalculationStatus(ctx context.Context) (*v1.RecalculationState, error) {
	// Get the latest running or pending history record
	history, err := r.data.db.ProxyGroupHistory.Query().
		Where(proxygrouphistory.StatusIn("running", "pending")).
		Order(ent.Desc(proxygrouphistory.FieldCreatedAt)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// No active recalculation
			return &v1.RecalculationState{
				State:    "idle",
				Progress: 0,
				Total:    0,
			}, nil
		}
		return nil, err
	}

	return &v1.RecalculationState{
		State:    history.Status,
		Progress: int32(history.Progress),
		Total:    int32(history.Total),
	}, nil
}

// GetGroupHistory gets group history
func (r *adminGroupRepo) GetGroupHistory(ctx context.Context, req *v1.GetGroupHistoryRequest) ([]*ent.ProxyGroupHistory, int64, error) {
	query := r.data.db.ProxyGroupHistory.Query()

	// Optional filters
	if req.GroupMode != "" {
		query = query.Where(proxygrouphistory.GroupModeEQ(req.GroupMode))
	}
	if req.TriggerType != "" {
		query = query.Where(proxygrouphistory.TriggerTypeEQ(req.TriggerType))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	list, err := query.
		Order(ent.Desc(proxygrouphistory.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// PreviewUserNodes previews user nodes
func (r *adminGroupRepo) PreviewUserNodes(ctx context.Context, userId int64) ([]*ent.ProxyNode, int64, error) {
	// Get user's group
	user, err := r.data.db.ProxyUser.Query().
		Where(proxyuser.IDEQ(userId)).
		Only(ctx)

	if err != nil {
		return nil, 0, err
	}

	var userGroupId int64
	if user.GroupID != nil {
		userGroupId = *user.GroupID
	}

	// TODO: Implement actual node filtering based on user group
	// For now, return all enabled nodes
	nodes, err := r.data.db.ProxyNode.Query().
		All(ctx)

	if err != nil {
		return nil, userGroupId, err
	}

	return nodes, userGroupId, nil
}
