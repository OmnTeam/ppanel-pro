package ticket

import (
	"context"

	pb "github.com/OmnTeam/npanel-pro/api/public/ticket/v1"
	ticketBiz "github.com/OmnTeam/npanel-pro/internal/biz/public/ticket"
	"github.com/OmnTeam/npanel-pro/internal/pkg/middleware"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TicketService implements the ticket service
type TicketService struct {
	pb.UnimplementedTicketServer

	uc     ticketBiz.TicketUseCase
	logger *log.Helper
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// NewTicketService creates a new ticket service
func NewTicketService(uc ticketBiz.TicketUseCase, logger log.Logger) *TicketService {
	return &TicketService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// CreateUserTicket creates a new ticket
func (s *TicketService) CreateUserTicket(ctx context.Context, req *pb.CreateUserTicketRequest) (*emptypb.Empty, error) {
	s.logger.Infof("[CreateUserTicket] title: %s", req.Title)

	// 从context获取user_id（通过认证middleware注入）
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[CreateUserTicket] userID: %d", userID)

	// 调用UseCase
	err := s.uc.CreateTicket(ctx, &ticketBiz.CreateTicketParams{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	})

	if err != nil {
		s.logger.Errorf("[CreateUserTicket] failed: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetUserTicketList gets user's ticket list with pagination
func (s *TicketService) GetUserTicketList(ctx context.Context, req *pb.GetUserTicketListRequest) (*pb.GetUserTicketListReply, error) {
	s.logger.Infof("[GetUserTicketList] page: %d, size: %d", req.Page, req.Size)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[GetUserTicketList] userID: %d", userID)

	// 调用UseCase
	result, err := s.uc.GetTicketList(ctx, &ticketBiz.GetTicketListParams{
		UserID: userID,
		Page:   int64(req.Page),
		Size:   int64(req.Size),
		Status: req.Status,
		Search: optionalString(req.Search),
	})

	if err != nil {
		s.logger.Errorf("[GetUserTicketList] failed: %v", err)
		return nil, err
	}

	// 转换为Proto响应
	list := make([]*pb.TicketInfo, 0, len(result.List))
	for _, t := range result.List {
		list = append(list, &pb.TicketInfo{
			Id:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			UserId:      t.UserID,
			Status:      t.Status,
			CreatedAt:   int64(t.CreatedAt),
			UpdatedAt:   int64(t.UpdatedAt),
		})
	}

	return &pb.GetUserTicketListReply{
		Total: int32(result.Total),
		List:  list,
	}, nil
}

// GetUserTicketDetails gets ticket details
func (s *TicketService) GetUserTicketDetails(ctx context.Context, req *pb.GetUserTicketDetailsRequest) (*pb.TicketInfo, error) {
	s.logger.Infof("[GetUserTicketDetails] id: %d", req.Id)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[GetUserTicketDetails] userID: %d", userID)

	// 调用UseCase
	ticket, err := s.uc.GetTicketDetails(ctx, &ticketBiz.GetTicketDetailsParams{
		UserID: userID,
		ID:     req.Id,
	})

	if err != nil {
		s.logger.Errorf("[GetUserTicketDetails] failed: %v", err)
		return nil, err
	}

	// 转换为Proto响应
	follows := make([]*pb.TicketFollow, 0, len(ticket.Follows))
	for _, f := range ticket.Follows {
		follows = append(follows, &pb.TicketFollow{
			Id:        f.ID,
			TicketId:  f.TicketID,
			From:      f.From,
			Type:      f.Type,
			Content:   f.Content,
			CreatedAt: int64(f.CreatedAt),
		})
	}

	return &pb.TicketInfo{
		Id:          ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		UserId:      ticket.UserID,
		Follow:      follows,
		Status:      ticket.Status,
		CreatedAt:   int64(ticket.CreatedAt),
		UpdatedAt:   int64(ticket.UpdatedAt),
	}, nil
}

// UpdateUserTicketStatus updates ticket status
func (s *TicketService) UpdateUserTicketStatus(ctx context.Context, req *pb.UpdateUserTicketStatusRequest) (*emptypb.Empty, error) {
	s.logger.Infof("[UpdateUserTicketStatus] id: %d, status: %d", req.Id, req.Status)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[UpdateUserTicketStatus] userID: %d", userID)

	// 调用UseCase
	err := s.uc.UpdateTicketStatus(ctx, &ticketBiz.UpdateTicketStatusParams{
		UserID: userID,
		ID:     req.Id,
		Status: req.GetStatus(),
	})

	if err != nil {
		s.logger.Errorf("[UpdateUserTicketStatus] failed: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateUserTicketFollow creates a follow-up for ticket
func (s *TicketService) CreateUserTicketFollow(ctx context.Context, req *pb.CreateUserTicketFollowRequest) (*emptypb.Empty, error) {
	s.logger.Infof("[CreateUserTicketFollow] ticketID: %d", req.TicketId)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[CreateUserTicketFollow] userID: %d", userID)

	// 调用UseCase
	err := s.uc.CreateTicketFollow(ctx, &ticketBiz.CreateTicketFollowParams{
		UserID:   userID,
		TicketID: req.TicketId,
		From:     req.From,
		Type:     req.Type,
		Content:  req.Content,
	})

	if err != nil {
		s.logger.Errorf("[CreateUserTicketFollow] failed: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
