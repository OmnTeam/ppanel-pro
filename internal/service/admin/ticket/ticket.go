package ticket

import (
	"context"

	pb "github.com/OmnTeam/ppanel-pro/api/admin/ticket/v1"
	ticketbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/ticket"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

type TicketService struct {
	pb.UnimplementedTicketServer
	uc *ticketbiz.TicketUseCase
}

func NewTicketService(uc *ticketbiz.TicketUseCase) *TicketService {
	return &TicketService{uc: uc}
}

// UpdateTicketStatus 更新工单状态
func (s *TicketService) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.UpdateTicketStatusReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	err := s.uc.UpdateTicketStatus(ctx, int(req.Id), int8(req.Status))
	if err != nil {
		return nil, err
	}

	return &pb.UpdateTicketStatusReply{
		Code:    int32(responsecode.AdminUpdateTicketStatusSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateTicketStatusSuccess],
	}, nil
}

// GetTicket 获取工单详情
func (s *TicketService) GetTicket(ctx context.Context, req *pb.GetTicketRequest) (*pb.GetTicketReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	ticket, err := s.uc.GetTicket(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &pb.GetTicketReply{
		Code:    int32(responsecode.AdminGetTicketSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetTicketSuccess],
		Data:    s.convertTicketToProto(ticket),
	}, nil
}

// CreateTicketFollow 创建工单跟进
func (s *TicketService) CreateTicketFollow(ctx context.Context, req *pb.CreateTicketFollowRequest) (*pb.CreateTicketFollowReply, error) {
	if req.TicketId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	follow := &ticketbiz.Follow{
		TicketId: req.TicketId,
		From:     req.From,
		Type:     int8(req.Type),
		Content:  req.Content,
	}

	err := s.uc.CreateTicketFollow(ctx, follow)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTicketFollowReply{
		Code:    int32(responsecode.AdminCreateTicketFollowSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateTicketFollowSuccess],
	}, nil
}

// GetTicketList 获取工单列表
func (s *TicketService) GetTicketList(ctx context.Context, req *pb.GetTicketListRequest) (*pb.GetTicketListReply, error) {
	var status *int8
	if req.Status != 0 {
		s := int8(req.Status)
		status = &s
	}

	page := int(req.Page)
	if page == 0 {
		page = 1
	}
	size := int(req.Size)
	if size == 0 {
		size = 10
	}

	userID := req.UserId

	total, list, err := s.uc.GetTicketList(ctx, page, size, userID, status, req.Search)
	if err != nil {
		return nil, err
	}

	tickets := make([]*pb.TicketInfo, 0, len(list))
	for _, ticket := range list {
		tickets = append(tickets, s.convertTicketToProto(ticket))
	}

	return &pb.GetTicketListReply{
		Code:    int32(responsecode.AdminGetTicketListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetTicketListSuccess],
		Data: &pb.GetTicketListData{
			Total: total,
			List:  tickets,
		},
	}, nil
}

// convertTicketToProto 将业务层的Ticket转换为proto格式
func (s *TicketService) convertTicketToProto(ticket *ticketbiz.Ticket) *pb.TicketInfo {
	follows := make([]*pb.TicketFollow, 0, len(ticket.Follow))
	for _, f := range ticket.Follow {
		follows = append(follows, &pb.TicketFollow{
			Id:        f.Id,
			TicketId:  f.TicketId,
			From:      f.From,
			Type:      int32(f.Type),
			Content:   f.Content,
			CreatedAt: int64(f.CreatedAt),
		})
	}

	return &pb.TicketInfo{
		Id:          ticket.Id,
		Title:       ticket.Title,
		Description: ticket.Description,
		UserId:      ticket.UserId,
		Follow:      follows,
		Status:      int32(ticket.Status),
		CreatedAt:   int64(ticket.CreatedAt),
		UpdatedAt:   int64(ticket.UpdatedAt),
	}
}
