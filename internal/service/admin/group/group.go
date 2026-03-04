package group

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/group/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/group"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// GroupService group service implementation
type GroupService struct {
	v1.UnimplementedGroupServer

	uc *group.GroupUseCase
}

// NewGroupService create group service
func NewGroupService(uc *group.GroupUseCase) *GroupService {
	return &GroupService{
		uc: uc,
	}
}

// ===== 用户组管理 =====

// GetUserGroupList 获取用户组列表
func (s *GroupService) GetUserGroupList(ctx context.Context, req *v1.GetUserGroupListRequest) (*v1.GetUserGroupListReply, error) {
	list, total, err := s.uc.GetUserGroupList(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert ent entities to proto messages
	userGroups := make([]*v1.UserGroup, 0, len(list))
	for _, item := range list {
		userGroups = append(userGroups, &v1.UserGroup{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Sort:        int32(item.Sort),
			CreatedAt:   item.CreatedAt.Unix(),
			UpdatedAt:   item.UpdatedAt.Unix(),
		})
	}

	return &v1.GetUserGroupListReply{
		Code:    int32(responsecode.AdminGetUserGroupListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetUserGroupListSuccess],
		Data: &v1.GetUserGroupListData{
			Total: total,
			List:  userGroups,
		},
	}, nil
}

// CreateUserGroup 创建用户组
func (s *GroupService) CreateUserGroup(ctx context.Context, req *v1.CreateUserGroupRequest) (*v1.CreateUserGroupReply, error) {
	id, err := s.uc.CreateUserGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.CreateUserGroupReply{
		Code:    int32(responsecode.AdminCreateUserGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateUserGroupSuccess],
		Data: &v1.CreateUserGroupData{
			Id: id,
		},
	}, nil
}

// UpdateUserGroup 更新用户组
func (s *GroupService) UpdateUserGroup(ctx context.Context, req *v1.UpdateUserGroupRequest) (*v1.UpdateUserGroupReply, error) {
	if err := s.uc.UpdateUserGroup(ctx, req); err != nil {
		return nil, err
	}

	return &v1.UpdateUserGroupReply{
		Code:    int32(responsecode.AdminUpdateUserGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateUserGroupSuccess],
		Data: &v1.UpdateUserGroupData{
			Success: true,
		},
	}, nil
}

// DeleteUserGroup 删除用户组
func (s *GroupService) DeleteUserGroup(ctx context.Context, req *v1.DeleteUserGroupRequest) (*v1.DeleteUserGroupReply, error) {
	if err := s.uc.DeleteUserGroup(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteUserGroupReply{
		Code:    int32(responsecode.AdminDeleteUserGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteUserGroupSuccess],
		Data: &v1.DeleteUserGroupData{
			Success: true,
		},
	}, nil
}

// UpdateUserUserGroup 更新用户的用户组
func (s *GroupService) UpdateUserUserGroup(ctx context.Context, req *v1.UpdateUserUserGroupRequest) (*v1.UpdateUserUserGroupReply, error) {
	if err := s.uc.UpdateUserUserGroup(ctx, req); err != nil {
		return nil, err
	}

	return &v1.UpdateUserUserGroupReply{
		Code:    int32(responsecode.AdminUpdateUserUserGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateUserUserGroupSuccess],
		Data: &v1.UpdateUserUserGroupData{
			Success: true,
		},
	}, nil
}

// ===== 节点组管理 =====

// GetNodeGroupList 获取节点组列表
func (s *GroupService) GetNodeGroupList(ctx context.Context, req *v1.GetNodeGroupListRequest) (*v1.GetNodeGroupListReply, error) {
	list, total, err := s.uc.GetNodeGroupList(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert ent entities to proto messages
	nodeGroups := make([]*v1.NodeGroup, 0, len(list))
	for _, item := range list {
		nodeGroups = append(nodeGroups, &v1.NodeGroup{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Sort:        int32(item.Sort),
			CreatedAt:   item.CreatedAt.Unix(),
			UpdatedAt:   item.UpdatedAt.Unix(),
		})
	}

	return &v1.GetNodeGroupListReply{
		Code:    int32(responsecode.AdminGetNodeGroupListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetNodeGroupListSuccess],
		Data: &v1.GetNodeGroupListData{
			Total: total,
			List:  nodeGroups,
		},
	}, nil
}

// CreateNodeGroup 创建节点组
func (s *GroupService) CreateNodeGroup(ctx context.Context, req *v1.CreateNodeGroupRequest) (*v1.CreateNodeGroupReply, error) {
	id, err := s.uc.CreateNodeGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.CreateNodeGroupReply{
		Code:    int32(responsecode.AdminCreateNodeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateNodeGroupSuccess],
		Data: &v1.CreateNodeGroupData{
			Id: id,
		},
	}, nil
}

// UpdateNodeGroup 更新节点组
func (s *GroupService) UpdateNodeGroup(ctx context.Context, req *v1.UpdateNodeGroupRequest) (*v1.UpdateNodeGroupReply, error) {
	if err := s.uc.UpdateNodeGroup(ctx, req); err != nil {
		return nil, err
	}

	return &v1.UpdateNodeGroupReply{
		Code:    int32(responsecode.AdminUpdateNodeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateNodeGroupSuccess],
		Data: &v1.UpdateNodeGroupData{
			Success: true,
		},
	}, nil
}

// DeleteNodeGroup 删除节点组
func (s *GroupService) DeleteNodeGroup(ctx context.Context, req *v1.DeleteNodeGroupRequest) (*v1.DeleteNodeGroupReply, error) {
	if err := s.uc.DeleteNodeGroup(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteNodeGroupReply{
		Code:    int32(responsecode.AdminDeleteNodeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteNodeGroupSuccess],
		Data: &v1.DeleteNodeGroupData{
			Success: true,
		},
	}, nil
}

// ===== 分组配置管理 =====

// GetGroupConfig 获取分组配置
func (s *GroupService) GetGroupConfig(ctx context.Context, req *v1.GetGroupConfigRequest) (*v1.GetGroupConfigReply, error) {
	config, state, err := s.uc.GetGroupConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.GetGroupConfigReply{
		Code:    int32(responsecode.AdminGetGroupConfigSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetGroupConfigSuccess],
		Data: &v1.GetGroupConfigData{
			Config: config,
			State:  state,
		},
	}, nil
}

// UpdateGroupConfig 更新分组配置
func (s *GroupService) UpdateGroupConfig(ctx context.Context, req *v1.UpdateGroupConfigRequest) (*v1.UpdateGroupConfigReply, error) {
	if err := s.uc.UpdateGroupConfig(ctx, req); err != nil {
		return nil, err
	}

	return &v1.UpdateGroupConfigReply{
		Code:    int32(responsecode.AdminUpdateGroupConfigSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateGroupConfigSuccess],
		Data: &v1.UpdateGroupConfigData{
			Success: true,
		},
	}, nil
}

// ===== 分组操作 =====

// RecalculateGroup 重新计算分组
func (s *GroupService) RecalculateGroup(ctx context.Context, req *v1.RecalculateGroupRequest) (*v1.RecalculateGroupReply, error) {
	historyId, err := s.uc.RecalculateGroup(ctx, &v1.RecalculateGroupRequest{Mode: req.Mode})
	if err != nil {
		return nil, err
	}

	return &v1.RecalculateGroupReply{
		Code:    int32(responsecode.AdminRecalculateGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminRecalculateGroupSuccess],
		Data: &v1.RecalculateGroupData{
			HistoryId: historyId,
		},
	}, nil
}

// GetRecalculationStatus 获取重新计算状态
func (s *GroupService) GetRecalculationStatus(ctx context.Context, req *v1.GetRecalculationStatusRequest) (*v1.GetRecalculationStatusReply, error) {
	state, err := s.uc.GetRecalculationStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.GetRecalculationStatusReply{
		Code:    int32(responsecode.AdminGetRecalculationStatusSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetRecalculationStatusSuccess],
		Data: &v1.GetRecalculationStatusData{
			State: state,
		},
	}, nil
}

// GetGroupHistory 获取分组历史
func (s *GroupService) GetGroupHistory(ctx context.Context, req *v1.GetGroupHistoryRequest) (*v1.GetGroupHistoryReply, error) {
	list, total, err := s.uc.GetGroupHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert ent entities to proto messages
	histories := make([]*v1.GroupHistory, 0, len(list))
	for _, item := range list {
		histories = append(histories, &v1.GroupHistory{
			Id:          item.ID,
			GroupMode:   item.GroupMode,
			TriggerType: item.TriggerType,
			Status:      item.Status,
			Progress:    int32(item.Progress),
			Total:       int32(item.Total),
			CreatedAt:   item.CreatedAt.Unix(),
		})
	}

	return &v1.GetGroupHistoryReply{
		Code:    int32(responsecode.AdminGetGroupHistorySuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetGroupHistorySuccess],
		Data: &v1.GetGroupHistoryData{
			Total: total,
			List:  histories,
		},
	}, nil
}

// GetGroupHistoryDetail 获取分组历史详情
func (s *GroupService) GetGroupHistoryDetail(ctx context.Context, req *v1.GetGroupHistoryDetailRequest) (*v1.GetGroupHistoryDetailReply, error) {
	// TODO: Implement get group history detail
	return &v1.GetGroupHistoryDetailReply{
		Code:    int32(responsecode.AdminGetGroupHistoryDetailSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetGroupHistoryDetailSuccess],
		Data: &v1.GetGroupHistoryDetailData{
			Detail: &v1.GroupHistoryDetail{
				History: &v1.GroupHistory{},
				Result:  "",
				Error:   "",
			},
		},
	}, nil
}

// ExportGroupResult 导出分组结果
func (s *GroupService) ExportGroupResult(ctx context.Context, req *v1.ExportGroupResultRequest) (*v1.ExportGroupResultReply, error) {
	// TODO: Implement export group result
	return &v1.ExportGroupResultReply{
		Code:    int32(responsecode.AdminExportGroupResultSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminExportGroupResultSuccess],
		Data: &v1.ExportGroupResultData{
			FileUrl: "",
		},
	}, nil
}

// MigrateUsers 迁移用户
func (s *GroupService) MigrateUsers(ctx context.Context, req *v1.MigrateUsersRequest) (*v1.MigrateUsersReply, error) {
	// TODO: Implement migrate users
	return &v1.MigrateUsersReply{
		Code:    int32(responsecode.AdminMigrateUsersSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminMigrateUsersSuccess],
		Data: &v1.MigrateUsersData{
			SuccessCount: 0,
			FailedCount:  0,
		},
	}, nil
}

// PreviewUserNodes 预览用户节点
func (s *GroupService) PreviewUserNodes(ctx context.Context, req *v1.PreviewUserNodesRequest) (*v1.PreviewUserNodesReply, error) {
	nodes, userGroupId, err := s.uc.PreviewUserNodes(ctx, &v1.PreviewUserNodesRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}

	// Convert ent entities to proto messages
	nodeList := make([]*v1.Node, 0, len(nodes))
	for _, item := range nodes {
		nodeList = append(nodeList, &v1.Node{
			Id:   item.ID,
			Name: item.Name,
			Tags: item.Tags,
			Sort: int32(item.Sort),
		})
	}

	return &v1.PreviewUserNodesReply{
		Code:    int32(responsecode.AdminPreviewUserNodesSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminPreviewUserNodesSuccess],
		Data: &v1.PreviewUserNodesData{
			UserId:      req.UserId,
			UserGroupId: userGroupId,
			Nodes:       nodeList,
		},
	}, nil
}

// ResetGroups 重置所有分组
func (s *GroupService) ResetGroups(ctx context.Context, req *v1.ResetGroupsRequest) (*v1.ResetGroupsReply, error) {
	// TODO: Implement reset groups
	return &v1.ResetGroupsReply{
		Code:    int32(responsecode.AdminResetGroupsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminResetGroupsSuccess],
		Data: &v1.ResetGroupsData{
			Success: true,
		},
	}, nil
}
