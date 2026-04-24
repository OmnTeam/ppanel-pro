package ticket

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Ticket 工单信息
type Ticket struct {
	Id          int64
	Title       string
	Description string
	UserId      int64
	Status      int8
	Follow      []*Follow
	CreatedAt   int // Unix timestamp in milliseconds
	UpdatedAt   int // Unix timestamp in milliseconds
}

// Follow 工单跟进信息
type Follow struct {
	Id        int64
	TicketId  int64
	From      string
	Type      int8
	Content   string
	CreatedAt int // Unix timestamp in milliseconds
}

// TicketRepo 工单仓库接口
type TicketRepo interface {
	// GetTicket 获取工单详情（包含跟进列表）
	GetTicket(ctx context.Context, id int) (*Ticket, error)
	// GetTicketList 获取工单列表
	GetTicketList(ctx context.Context, page, size int, userId int64, status *int8, search string) (int32, []*Ticket, error)
	// UpdateTicketStatus 更新工单状态
	UpdateTicketStatus(ctx context.Context, id int, status int8) error
	// CreateTicketFollow 创建工单跟进
	CreateTicketFollow(ctx context.Context, follow *Follow) error
	// GetTicketById 获取工单基本信息（不含跟进列表）
	GetTicketById(ctx context.Context, id int) (*Ticket, error)
}

// TicketUseCase 工单用例
type TicketUseCase struct {
	repo TicketRepo
	log  *log.Helper
}

// NewTicketUseCase 创建工单用例
func NewTicketUseCase(repo TicketRepo, logger log.Logger) *TicketUseCase {
	return &TicketUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/ticket")),
	}
}

// GetTicket 获取工单详情
func (uc *TicketUseCase) GetTicket(ctx context.Context, id int) (*Ticket, error) {
	ticket, err := uc.repo.GetTicket(ctx, id)
	if err != nil {
		uc.log.Errorf("GetTicket error: %v, id=%d", err, id)
		return nil, err
	}
	return ticket, nil
}

// GetTicketList 获取工单列表
func (uc *TicketUseCase) GetTicketList(ctx context.Context, page, size int, userId int64, status *int8, search string) (int32, []*Ticket, error) {
	total, list, err := uc.repo.GetTicketList(ctx, page, size, userId, status, search)
	if err != nil {
		uc.log.Errorf("GetTicketList error: %v, page=%d, size=%d, userId=%d", err, page, size, userId)
		return 0, nil, err
	}
	return total, list, nil
}

// UpdateTicketStatus 更新工单状态
func (uc *TicketUseCase) UpdateTicketStatus(ctx context.Context, id int, status int8) error {
	err := uc.repo.UpdateTicketStatus(ctx, id, status)
	if err != nil {
		uc.log.Errorf("UpdateTicketStatus error: %v, id=%d, status=%d", err, id, status)
		return err
	}
	return nil
}

// CreateTicketFollow 创建工单跟进
func (uc *TicketUseCase) CreateTicketFollow(ctx context.Context, follow *Follow) error {
	// 检查工单是否存在
	_, err := uc.repo.GetTicketById(ctx, int(follow.TicketId))
	if err != nil {
		uc.log.Errorf("CreateTicketFollow GetTicketById error: %v, ticketId=%d", err, int(follow.TicketId))
		return err
	}

	// 创建跟进
	err = uc.repo.CreateTicketFollow(ctx, follow)
	if err != nil {
		uc.log.Errorf("CreateTicketFollow error: %v, follow=%+v", err, follow)
		return err
	}

	// 更新工单状态为 Waiting (2)
	err = uc.repo.UpdateTicketStatus(ctx, int(follow.TicketId), 2)
	if err != nil {
		uc.log.Errorf("CreateTicketFollow UpdateTicketStatus error: %v, ticketId=%d", err, int(follow.TicketId))
		return err
	}

	return nil
}
