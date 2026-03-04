package ticket

import (
	"context"
	"strconv"

	pb "github.com/OmnTeam/ppanel-pro/api/public/ticket/v1"
	ticketBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/ticket"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

// TicketService implements the ticket service
type TicketService struct {
	pb.UnimplementedTicketServer

	uc     ticketBiz.TicketUseCase
	logger *log.Helper
}

// NewTicketService creates a new ticket service
func NewTicketService(uc ticketBiz.TicketUseCase, logger log.Logger) *TicketService {
	return &TicketService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// CreateUserTicket creates a new ticket
func (s *TicketService) CreateUserTicket(ctx context.Context, req *pb.CreateUserTicketRequest) (*pb.TicketCreateReply, error) {
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

	return &pb.TicketCreateReply{
		Code:    int32(responsecode.UserTicketCreateSuccess),
		Message: responsecode.CodeMessages[responsecode.UserTicketCreateSuccess],
	}, nil
}

// GetUserTicketList gets user's ticket list with pagination
func (s *TicketService) GetUserTicketList(ctx context.Context, req *pb.GetUserTicketListRequest) (*pb.TicketListReply, error) {
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
		Search: req.Search,
	})

	if err != nil {
		s.logger.Errorf("[GetUserTicketList] failed: %v", err)
		return nil, err
	}

	// 转换为Proto响应
	list := make([]*pb.TicketInfo, 0, len(result.List))
	for _, t := range result.List {
		list = append(list, &pb.TicketInfo{
			Id:          strconv.FormatInt(t.ID, 10),
			Title:       t.Title,
			Description: t.Description,
			UserId:      strconv.FormatInt(t.UserID, 10),
			Status:      t.Status,
			CreatedAt:   strconv.FormatInt(int64(t.CreatedAt), 10), // Already Unix milliseconds
			UpdatedAt:   strconv.FormatInt(int64(t.UpdatedAt), 10), // Already Unix milliseconds
		})
	}

	return &pb.TicketListReply{
		Code:    int32(responsecode.UserTicketListQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserTicketListQuerySuccess],
		Data: &pb.TicketListData{
			Total: int32(result.Total),
			List:  list,
		},
	}, nil
}

// GetUserTicketDetails gets ticket details
func (s *TicketService) GetUserTicketDetails(ctx context.Context, req *pb.GetUserTicketDetailsRequest) (*pb.TicketDetailReply, error) {
	s.logger.Infof("[GetUserTicketDetails] id: %d", req.Id)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[GetUserTicketDetails] userID: %d", userID)

	// 调用UseCase
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	ticket, err := s.uc.GetTicketDetails(ctx, &ticketBiz.GetTicketDetailsParams{
		UserID: userID,
		ID:     id,
	})

	if err != nil {
		s.logger.Errorf("[GetUserTicketDetails] failed: %v", err)
		return nil, err
	}

	// 转换为Proto响应
	follows := make([]*pb.TicketFollow, 0, len(ticket.Follows))
	for _, f := range ticket.Follows {
		follows = append(follows, &pb.TicketFollow{
			Id:        strconv.FormatInt(f.ID, 10),
			TicketId:  strconv.FormatInt(f.TicketID, 10),
			From:      f.From,
			Type:      f.Type,
			Content:   f.Content,
			CreatedAt: strconv.FormatInt(int64(f.CreatedAt), 10), // Already Unix milliseconds
		})
	}

	return &pb.TicketDetailReply{
		Code:    int32(responsecode.UserTicketDetailQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserTicketDetailQuerySuccess],
		Data: &pb.TicketInfo{
			Id:          strconv.FormatInt(ticket.ID, 10),
			Title:       ticket.Title,
			Description: ticket.Description,
			UserId:      strconv.FormatInt(ticket.UserID, 10),
			Follow:      follows,
			Status:      ticket.Status,
			CreatedAt:   strconv.FormatInt(int64(ticket.CreatedAt), 10), // Already Unix milliseconds
			UpdatedAt:   strconv.FormatInt(int64(ticket.UpdatedAt), 10), // Already Unix milliseconds
		},
	}, nil
}

// UpdateUserTicketStatus updates ticket status
func (s *TicketService) UpdateUserTicketStatus(ctx context.Context, req *pb.UpdateUserTicketStatusRequest) (*pb.TicketStatusReply, error) {
	s.logger.Infof("[UpdateUserTicketStatus] id: %d, status: %d", req.Id, req.Status)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[UpdateUserTicketStatus] userID: %d", userID)

	// 调用UseCase
	err := s.uc.UpdateTicketStatus(ctx, &ticketBiz.UpdateTicketStatusParams{
		UserID: userID,
		ID:     parseInt64(req.Id),
		Status: req.Status,
	})

	if err != nil {
		s.logger.Errorf("[UpdateUserTicketStatus] failed: %v", err)
		return nil, err
	}

	return &pb.TicketStatusReply{
		Code:    int32(responsecode.UserTicketStatusUpdateSuccess),
		Message: responsecode.CodeMessages[responsecode.UserTicketStatusUpdateSuccess],
	}, nil
}

// CreateUserTicketFollow creates a follow-up for ticket
func (s *TicketService) CreateUserTicketFollow(ctx context.Context, req *pb.CreateUserTicketFollowRequest) (*pb.TicketFollowReply, error) {
	s.logger.Infof("[CreateUserTicketFollow] ticketID: %d", req.TicketId)

	// 从context获取user_id
	userID := middleware.GetUserID(ctx)
	s.logger.Infof("[CreateUserTicketFollow] userID: %d", userID)

	// 调用UseCase
	err := s.uc.CreateTicketFollow(ctx, &ticketBiz.CreateTicketFollowParams{
		UserID:   userID,
		TicketID: parseInt64(req.TicketId),
		From:     req.From,
		Type:     req.Type,
		Content:  req.Content,
	})

	if err != nil {
		s.logger.Errorf("[CreateUserTicketFollow] failed: %v", err)
		return nil, err
	}

	return &pb.TicketFollowReply{
		Code:    int32(responsecode.UserTicketFollowCreateSuccess),
		Message: responsecode.CodeMessages[responsecode.UserTicketFollowCreateSuccess],
	}, nil
}
