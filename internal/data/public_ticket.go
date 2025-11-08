package data

import (
	"context"
	"fmt"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyticket"
	"github.com/OmnTeam/ppanel-pro/ent/proxyticketfollow"
	ticketBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/ticket"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type publicTicketRepo struct {
	data   *Data
	logger *log.Helper
}

// NewPublicTicketRepo creates a new ticket repository
func NewPublicTicketRepo(d *Data, logger log.Logger) ticketBiz.TicketRepo {
	return &publicTicketRepo{
		data:   d,
		logger: log.NewHelper(logger),
	}
}

// CreateTicket creates a new ticket
// 完整复刻原项目 createUserTicketLogic.go
// ⚠️ 注意：当前schema暂不支持tenant_id字段
func (r *publicTicketRepo) CreateTicket(ctx context.Context, tenantID, userID int64, title, description string) error {
	r.logger.Infof("[CreateTicket] tenantID: %d, userID: %d, title: %s", tenantID, userID, title)

	// 创建工单
	err := r.data.db.ProxyTicket.Create().
		SetUserID(userID).
		SetTitle(title).
		SetDescription(description).
		SetStatus(int8(ticketBiz.StatusPending)).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("[CreateTicket] 创建工单失败: %v", err)
		return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("创建工单失败: %v", err))
	}

	r.logger.Infof("[CreateTicket] 工单创建成功")
	return nil
}

// GetTicketList gets user's ticket list
// 完整复刻原项目 getUserTicketListLogic.go
// ⚠️ 注意：当前schema暂不支持tenant_id字段
// 支持分页、状态过滤、搜索功能
func (r *publicTicketRepo) GetTicketList(ctx context.Context, tenantID, userID int64, page, size int64, status *int32, search *string) (int64, []*ticketBiz.TicketInfo, error) {
	r.logger.Infof("[GetTicketList] tenantID: %d, userID: %d, page: %d, size: %d", tenantID, userID, page, size)

	// 构建基础查询 - 包含user_id过滤
	query := r.data.db.ProxyTicket.Query().
		Where(proxyticket.UserIDEQ(userID))

	// 状态过滤
	// ⚠️ 重要：完整复刻原项目逻辑（model.go:68-69）
	// 如果指定了status，按指定值过滤；否则默认排除已关闭的工单（status != 4）
	if status != nil {
		r.logger.Infof("[GetTicketList] 状态过滤: %d", *status)
		query = query.Where(proxyticket.StatusEQ(int8(*status)))
	} else {
		// 默认排除已关闭的工单（Closed = 4）
		r.logger.Infof("[GetTicketList] 默认排除已关闭工单")
		query = query.Where(proxyticket.StatusNEQ(int8(ticketBiz.StatusClosed)))
	}

	// 搜索功能（标题或描述包含关键字）
	if search != nil && *search != "" {
		r.logger.Infof("[GetTicketList] 搜索关键字: %s", *search)
		query = query.Where(proxyticket.Or(
			proxyticket.TitleContains(*search),
			proxyticket.DescriptionContains(*search),
		))
	}

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("[GetTicketList] 查询总数失败: %v", err)
		return 0, nil, errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查询总数失败: %v", err))
	}

	// 分页查询
	offset := (page - 1) * size
	tickets, err := query.
		Order(ent.Desc(proxyticket.FieldID)). // ⚠️ 重要：原项目按ID降序，不是created_at
		Offset(int(offset)).
		Limit(int(size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("[GetTicketList] 查询工单列表失败: %v", err)
		return 0, nil, errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查询工单列表失败: %v", err))
	}

	// 转换为业务对象
	list := make([]*ticketBiz.TicketInfo, 0, len(tickets))
	for _, t := range tickets {
		list = append(list, &ticketBiz.TicketInfo{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			UserID:      t.UserID,
			Status:      int32(t.Status),
			CreatedAt:   t.CreatedAt.UnixMilli(), // Convert to Unix milliseconds
			UpdatedAt:   t.UpdatedAt.UnixMilli(), // Convert to Unix milliseconds
		})
	}

	r.logger.Infof("[GetTicketList] 查询成功, total: %d, count: %d", total, len(list))
	return int64(total), list, nil
}

// GetTicketByID gets ticket by ID
// ⚠️ 注意：当前schema暂不支持tenant_id字段
func (r *publicTicketRepo) GetTicketByID(ctx context.Context, tenantID, ticketID int64) (*ticketBiz.TicketInfo, error) {
	r.logger.Infof("[GetTicketByID] tenantID: %d, ticketID: %d", tenantID, ticketID)

	// 查询工单
	ticket, err := r.data.db.ProxyTicket.Query().
		Where(proxyticket.IDEQ(ticketID)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Errorf("[GetTicketByID] 工单不存在: ticketID=%d", ticketID)
			return nil, errors.NotFound("TICKET_NOT_FOUND", "工单不存在")
		}
		r.logger.Errorf("[GetTicketByID] 查询工单失败: %v", err)
		return nil, errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查询工单失败: %v", err))
	}

	r.logger.Infof("[GetTicketByID] 查询成功")
	return &ticketBiz.TicketInfo{
		ID:          ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		UserID:      ticket.UserID,
		Status:      int32(ticket.Status),
		CreatedAt:   ticket.CreatedAt.UnixMilli(), // Convert to Unix milliseconds
		UpdatedAt:   ticket.UpdatedAt.UnixMilli(), // Convert to Unix milliseconds
	}, nil
}

// UpdateTicketStatus updates ticket status
// 完整复刻原项目 updateUserTicketStatusLogic.go
// ⚠️ 注意：当前schema暂不支持tenant_id字段，使用user_id过滤确保权限安全
func (r *publicTicketRepo) UpdateTicketStatus(ctx context.Context, tenantID, userID, ticketID int64, status int32) error {
	r.logger.Infof("[UpdateTicketStatus] tenantID: %d, userID: %d, ticketID: %d, status: %d",
		tenantID, userID, ticketID, status)

	// 更新状态 - 使用user_id过滤确保安全
	affected, err := r.data.db.ProxyTicket.Update().
		Where(proxyticket.IDEQ(ticketID),
			proxyticket.UserIDEQ(userID)).
		SetStatus(int8(status)).
		Save(ctx)

	if err != nil {
		r.logger.Errorf("[UpdateTicketStatus] 更新状态失败: %v", err)
		return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("更新状态失败: %v", err))
	}

	if affected == 0 {
		r.logger.Errorf("[UpdateTicketStatus] 工单不存在或无权限, affected: %d", affected)
		return errors.NotFound("TICKET_NOT_FOUND", "工单不存在或无权限")
	}

	r.logger.Infof("[UpdateTicketStatus] 更新成功, affected: %d", affected)
	return nil
}

// CreateTicketFollow creates a follow-up record
// 完整复刻原项目 createUserTicketFollowLogic.go
// ⚠️ 注意：当前schema暂不支持tenant_id字段
func (r *publicTicketRepo) CreateTicketFollow(ctx context.Context, tenantID int64, ticketID int64, from string, followType int32, content string) error {
	r.logger.Infof("[CreateTicketFollow] tenantID: %d, ticketID: %d, from: %s, type: %d",
		tenantID, ticketID, from, followType)

	// 创建跟进记录
	err := r.data.db.ProxyTicketFollow.Create().
		SetTicketID(ticketID).
		SetFrom(from).
		SetType(int8(followType)).
		SetContent(content).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("[CreateTicketFollow] 创建跟进记录失败: %v", err)
		return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("创建跟进记录失败: %v", err))
	}

	r.logger.Infof("[CreateTicketFollow] 创建成功")
	return nil
}

// GetTicketFollows gets all follow-ups for a ticket
// ⚠️ 注意：当前schema暂不支持tenant_id字段
func (r *publicTicketRepo) GetTicketFollows(ctx context.Context, tenantID, ticketID int64) ([]*ticketBiz.TicketFollow, error) {
	r.logger.Infof("[GetTicketFollows] tenantID: %d, ticketID: %d", tenantID, ticketID)

	// 查询跟进记录
	follows, err := r.data.db.ProxyTicketFollow.Query().
		Where(proxyticketfollow.TicketIDEQ(ticketID)).
		Order(ent.Asc(proxyticketfollow.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("[GetTicketFollows] 查询跟进记录失败: %v", err)
		return nil, errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查询跟进记录失败: %v", err))
	}

	// 转换为业务对象
	list := make([]*ticketBiz.TicketFollow, 0, len(follows))
	for _, f := range follows {
		list = append(list, &ticketBiz.TicketFollow{
			ID:        f.ID,
			TicketID:  f.TicketID,
			From:      f.From,
			Type:      int32(f.Type),
			Content:   f.Content,
			CreatedAt: f.CreatedAt.UnixMilli(), // Convert to Unix milliseconds
		})
	}

	r.logger.Infof("[GetTicketFollows] 查询成功, count: %d", len(list))
	return list, nil
}
