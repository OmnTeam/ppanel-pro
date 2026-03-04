package ticket

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type ticketUseCase struct {
	repo   TicketRepo
	logger *log.Helper
}

// NewTicketUseCase creates a new ticket use case
func NewTicketUseCase(repo TicketRepo, logger log.Logger) TicketUseCase {
	return &ticketUseCase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// CreateTicket creates a new ticket
// 完整复刻原项目 createUserTicketLogic.go
// 包含用户权限验证
func (uc *ticketUseCase) CreateTicket(ctx context.Context, params *CreateTicketParams) error {
	uc.logger.Infof("[CreateTicket] userID: %d, title: %s", params.UserID, params.Title)

	// 调用repo创建工单
	err := uc.repo.CreateTicket(ctx, int(params.UserID), params.Title, params.Description)
	if err != nil {
		uc.logger.Errorf("[CreateTicket] 创建工单失败: %v", err)
		return err
	}

	uc.logger.Infof("[CreateTicket] 工单创建成功")
	return nil
}

// GetTicketList gets user's ticket list with pagination
// 完整复刻原项目 getUserTicketListLogic.go
// 包含分页、状态过滤、搜索功能
func (uc *ticketUseCase) GetTicketList(ctx context.Context, params *GetTicketListParams) (*GetTicketListResult, error) {
	uc.logger.Infof("[GetTicketList] userID: %d, page: %d, size: %d",
		params.UserID, params.Page, int(params.Size))

	total, list, err := uc.repo.GetTicketList(ctx, int(params.UserID),
		int(params.Page), int(params.Size), params.Status, params.Search)
	if err != nil {
		uc.logger.Errorf("[GetTicketList] 查询工单列表失败: %v", err)
		return nil, err
	}

	uc.logger.Infof("[GetTicketList] 查询成功, total: %d", total)
	return &GetTicketListResult{
		Total: total,
		List:  list,
	}, nil
}

// GetTicketDetails gets ticket details with follow-ups
// 完整复刻原项目 getUserTicketDetailsLogic.go
// 包含权限验证：只能查看自己的工单
func (uc *ticketUseCase) GetTicketDetails(ctx context.Context, params *GetTicketDetailsParams) (*TicketInfo, error) {
	uc.logger.Infof("[GetTicketDetails] userID: %d, ticketID: %d",
		params.UserID, int(params.ID))

	// 查询工单基本信息
	ticket, err := uc.repo.GetTicketByID(ctx, int(params.ID))
	if err != nil {
		uc.logger.Errorf("[GetTicketDetails] 查询工单失败: %v", err)
		return nil, err
	}

	// 权限验证：检查工单是否属于当前用户
	if ticket.UserID != params.UserID {
		uc.logger.Errorf("[GetTicketDetails] 权限验证失败, ticket.UserID: %d, current userID: %d",
			ticket.UserID, int(params.UserID))
		return nil, errors.Forbidden("INVALID_ACCESS", "无权访问该工单")
	}

	// 查询工单跟进记录
	follows, err := uc.repo.GetTicketFollows(ctx, int(params.ID))
	if err != nil {
		uc.logger.Errorf("[GetTicketDetails] 查询跟进记录失败: %v", err)
		return nil, err
	}

	ticket.Follows = follows

	uc.logger.Infof("[GetTicketDetails] 查询成功, follows: %d", len(follows))
	return ticket, nil
}

// UpdateTicketStatus updates ticket status
// 完整复刻原项目 updateUserTicketStatusLogic.go
// 包含权限验证：只能更新自己的工单
func (uc *ticketUseCase) UpdateTicketStatus(ctx context.Context, params *UpdateTicketStatusParams) error {
	uc.logger.Infof("[UpdateTicketStatus] userID: %d, ticketID: %d, status: %d",
		params.UserID, params.ID, params.Status)

	// 调用repo更新状态（repo层会验证用户权限）
	err := uc.repo.UpdateTicketStatus(ctx, params.UserID, params.ID, params.Status)
	if err != nil {
		uc.logger.Errorf("[UpdateTicketStatus] 更新工单状态失败: %v", err)
		return err
	}

	uc.logger.Infof("[UpdateTicketStatus] 更新成功")
	return nil
}

// CreateTicketFollow creates a follow-up and updates ticket status to Pending
// 完整复刻原项目 createUserTicketFollowLogic.go
// 步骤：
// 1. 验证工单存在且属于当前用户
// 2. 创建跟进记录
// 3. 将工单状态更新为Pending（待处理）
func (uc *ticketUseCase) CreateTicketFollow(ctx context.Context, params *CreateTicketFollowParams) error {
	uc.logger.Infof("[CreateTicketFollow] userID: %d, ticketID: %d",
		params.UserID, params.TicketID)

	// 1. 查询工单，验证权限
	ticket, err := uc.repo.GetTicketByID(ctx, int(params.TicketID))
	if err != nil {
		uc.logger.Errorf("[CreateTicketFollow] 查询工单失败: %v", err)
		return err
	}

	// 权限验证：检查工单是否属于当前用户
	if ticket.UserID != params.UserID {
		uc.logger.Errorf("[CreateTicketFollow] 权限验证失败, ticket.UserID: %d, current userID: %d",
			ticket.UserID, int(params.UserID))
		return errors.Forbidden("INVALID_ACCESS", "无权操作该工单")
	}

	// 2. 创建跟进记录
	err = uc.repo.CreateTicketFollow(ctx, params.TicketID,
		params.From, params.Type, params.Content)
	if err != nil {
		uc.logger.Errorf("[CreateTicketFollow] 创建跟进记录失败: %v", err)
		return err
	}

	// 3. 将工单状态更新为Pending（待处理）
	err = uc.repo.UpdateTicketStatus(ctx, params.UserID, params.TicketID, StatusPending)
	if err != nil {
		uc.logger.Errorf("[CreateTicketFollow] 更新工单状态失败: %v", err)
		return err
	}

	uc.logger.Infof("[CreateTicketFollow] 跟进记录创建成功")
	return nil
}
