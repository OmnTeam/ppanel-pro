package subscribe

import (
	"context"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/subscribe/v1"
	"github.com/OmnTeam/npanel-pro/internal/biz/admin/subscribe"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
)

// SubscribeService subscribe service implementation
type SubscribeService struct {
	v1.UnimplementedSubscribeServer

	uc *subscribe.SubscribeUseCase
}

// NewSubscribeService create subscribe service
func NewSubscribeService(uc *subscribe.SubscribeUseCase) *SubscribeService {
	return &SubscribeService{
		uc: uc,
	}
}

// ==================== Subscribe Operations ====================

// CreateSubscribe create subscribe
func (s *SubscribeService) CreateSubscribe(ctx context.Context, req *v1.CreateSubscribeRequest) (*v1.CreateSubscribeReply, error) {
	if err := s.uc.CreateSubscribe(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CreateSubscribeReply{
		Code:    int32(responsecode.AdminCreateSubscribeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateSubscribeSuccess],
		Data: &v1.CreateSubscribeData{
			Success: true,
		},
	}, nil
}

// UpdateSubscribe update subscribe
func (s *SubscribeService) UpdateSubscribe(ctx context.Context, req *v1.UpdateSubscribeRequest) (*v1.UpdateSubscribeReply, error) {
	if err := s.uc.UpdateSubscribe(ctx, req); err != nil {
		return nil, err
	}
	return &v1.UpdateSubscribeReply{
		Code:    int32(responsecode.AdminUpdateSubscribeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateSubscribeSuccess],
		Data: &v1.UpdateSubscribeData{
			Success: true,
		},
	}, nil
}

// DeleteSubscribe delete subscribe
func (s *SubscribeService) DeleteSubscribe(ctx context.Context, req *v1.DeleteSubscribeRequest) (*v1.DeleteSubscribeReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	if err := s.uc.DeleteSubscribe(ctx, int(req.Id)); err != nil {
		return nil, err
	}
	return &v1.DeleteSubscribeReply{
		Code:    int32(responsecode.AdminDeleteSubscribeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteSubscribeSuccess],
		Data: &v1.DeleteSubscribeData{
			Success: true,
		},
	}, nil
}

// BatchDeleteSubscribe batch delete subscribes
func (s *SubscribeService) BatchDeleteSubscribe(ctx context.Context, req *v1.BatchDeleteSubscribeRequest) (*v1.BatchDeleteSubscribeReply, error) {
	idsInt := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		if id <= 0 {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		idsInt[i] = int(id)
	}
	if err := s.uc.BatchDeleteSubscribe(ctx, idsInt); err != nil {
		return nil, err
	}
	return &v1.BatchDeleteSubscribeReply{
		Code:    int32(responsecode.AdminBatchDeleteSubscribeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteSubscribeSuccess],
		Data: &v1.BatchDeleteSubscribeData{
			Success: true,
		},
	}, nil
}

// GetSubscribeDetails get subscribe details
func (s *SubscribeService) GetSubscribeDetails(ctx context.Context, req *v1.GetSubscribeDetailsRequest) (*v1.GetSubscribeDetailsReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	subscribe, err := s.uc.GetSubscribeDetails(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &v1.GetSubscribeDetailsReply{
		Code:    int32(responsecode.AdminGetSubscribeDetailsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetSubscribeDetailsSuccess],
		Data: &v1.GetSubscribeDetailsData{
			Subscribe: subscribe,
		},
	}, nil
}

// GetSubscribeList get subscribe list
func (s *SubscribeService) GetSubscribeList(ctx context.Context, req *v1.GetSubscribeListRequest) (*v1.GetSubscribeListReply, error) {
	data, err := s.uc.GetSubscribeList(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.GetSubscribeListReply{
		Code:    int32(responsecode.AdminGetSubscribeListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetSubscribeListSuccess],
		Data:    data,
	}, nil
}

// SubscribeSort subscribe sort
func (s *SubscribeService) SubscribeSort(ctx context.Context, req *v1.SubscribeSortRequest) (*v1.SubscribeSortReply, error) {
	if err := s.uc.SubscribeSort(ctx, req); err != nil {
		return nil, err
	}
	return &v1.SubscribeSortReply{
		Code:    int32(responsecode.AdminSubscribeSortSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminSubscribeSortSuccess],
		Data: &v1.SubscribeSortData{
			Success: true,
		},
	}, nil
}

// ==================== Subscribe Group Operations ====================

// CreateSubscribeGroup create subscribe group
func (s *SubscribeService) CreateSubscribeGroup(ctx context.Context, req *v1.CreateSubscribeGroupRequest) (*v1.CreateSubscribeGroupReply, error) {
	if err := s.uc.CreateSubscribeGroup(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CreateSubscribeGroupReply{
		Code:    int32(responsecode.AdminCreateSubscribeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateSubscribeGroupSuccess],
		Data: &v1.CreateSubscribeGroupData{
			Success: true,
		},
	}, nil
}

// UpdateSubscribeGroup update subscribe group
func (s *SubscribeService) UpdateSubscribeGroup(ctx context.Context, req *v1.UpdateSubscribeGroupRequest) (*v1.UpdateSubscribeGroupReply, error) {
	if err := s.uc.UpdateSubscribeGroup(ctx, req); err != nil {
		return nil, err
	}
	return &v1.UpdateSubscribeGroupReply{
		Code:    int32(responsecode.AdminUpdateSubscribeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateSubscribeGroupSuccess],
		Data: &v1.UpdateSubscribeGroupData{
			Success: true,
		},
	}, nil
}

// DeleteSubscribeGroup delete subscribe group
func (s *SubscribeService) DeleteSubscribeGroup(ctx context.Context, req *v1.DeleteSubscribeGroupRequest) (*v1.DeleteSubscribeGroupReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	if err := s.uc.DeleteSubscribeGroup(ctx, int(req.Id)); err != nil {
		return nil, err
	}
	return &v1.DeleteSubscribeGroupReply{
		Code:    int32(responsecode.AdminDeleteSubscribeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteSubscribeGroupSuccess],
		Data: &v1.DeleteSubscribeGroupData{
			Success: true,
		},
	}, nil
}

// BatchDeleteSubscribeGroup batch delete subscribe groups
func (s *SubscribeService) BatchDeleteSubscribeGroup(ctx context.Context, req *v1.BatchDeleteSubscribeGroupRequest) (*v1.BatchDeleteSubscribeGroupReply, error) {
	idsInt := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		if id <= 0 {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		idsInt[i] = int(id)
	}
	if err := s.uc.BatchDeleteSubscribeGroup(ctx, idsInt); err != nil {
		return nil, err
	}
	return &v1.BatchDeleteSubscribeGroupReply{
		Code:    int32(responsecode.AdminBatchDeleteSubscribeGroupSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteSubscribeGroupSuccess],
		Data: &v1.BatchDeleteSubscribeGroupData{
			Success: true,
		},
	}, nil
}

// GetSubscribeGroupList get subscribe group list
func (s *SubscribeService) GetSubscribeGroupList(ctx context.Context, req *v1.GetSubscribeGroupListRequest) (*v1.GetSubscribeGroupListReply, error) {
	data, err := s.uc.GetSubscribeGroupList(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSubscribeGroupListReply{
		Code:    int32(responsecode.AdminGetSubscribeGroupListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetSubscribeGroupListSuccess],
		Data:    data,
	}, nil
}

func (s *SubscribeService) ResetAllSubscribeToken(ctx context.Context, req *v1.ResetAllSubscribeTokenRequest) (*v1.ResetAllSubscribeTokenReply, error) {
	if err := s.uc.ResetAllSubscribeToken(ctx); err != nil {
		return nil, err
	}

	return &v1.ResetAllSubscribeTokenReply{
		Code:    200,
		Message: "success",
		Data: &v1.ResetAllSubscribeTokenData{
			Success: true,
		},
	}, nil
}
