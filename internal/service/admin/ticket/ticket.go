package ticket

import (
	"context"
	"strconv"

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

func parseStringID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return id, nil
}

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

// UpdateTicketStatus 更新工单状态
func (s *TicketService) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.UpdateTicketStatusReply, error) {
	id, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.uc.UpdateTicketStatus(ctx, int(id), int8(req.Status))
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
	id, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	ticket, err := s.uc.GetTicket(ctx, int(id))
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
	ticketID, err := parseStringID(req.TicketId)
	if err != nil {
		return nil, err
	}

	follow := &ticketbiz.Follow{
		TicketId: ticketID,
		From:     req.From,
		Type:     int8(req.Type),
		Content:  req.Content,
	}

	err = s.uc.CreateTicketFollow(ctx, follow)
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

	var userID int64
	var err error
	if req.UserId != "" {
		userID, err = parseStringID(req.UserId)
		if err != nil {
			return nil, err
		}
	}

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
			Id:        formatInt64(f.Id),
			TicketId:  formatInt64(f.TicketId),
			From:      f.From,
			Type:      int32(f.Type),
			Content:   f.Content,
			CreatedAt: int64(f.CreatedAt),
		})
	}

	return &pb.TicketInfo{
		Id:          formatInt64(ticket.Id),
		Title:       ticket.Title,
		Description: ticket.Description,
		UserId:      formatInt64(ticket.UserId),
		Follow:      follows,
		Status:      int32(ticket.Status),
		CreatedAt:   int64(ticket.CreatedAt),
		UpdatedAt:   int64(ticket.UpdatedAt),
	}
}
