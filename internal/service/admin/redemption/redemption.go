package redemption

import (
	"context"
	"strconv"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/redemption/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/redemption"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// RedemptionService redemption service implementation
type RedemptionService struct {
	v1.UnimplementedRedemptionServer

	uc *redemption.RedemptionUseCase
}

// NewRedemptionService create redemption service
func NewRedemptionService(uc *redemption.RedemptionUseCase) *RedemptionService {
	return &RedemptionService{
		uc: uc,
	}
}

// CreateRedemptionCode 创建兑换码
func (s *RedemptionService) CreateRedemptionCode(ctx context.Context, req *v1.CreateRedemptionCodeRequest) (*v1.CreateRedemptionCodeReply, error) {
	createdCount, err := s.uc.CreateRedemptionCode(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.CreateRedemptionCodeReply{
		Code:    int32(responsecode.AdminCreateRedemptionCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateRedemptionCodeSuccess],
		Data: &v1.CreateRedemptionCodeData{
			CreatedCount: createdCount,
		},
	}, nil
}

// UpdateRedemptionCode 更新兑换码
func (s *RedemptionService) UpdateRedemptionCode(ctx context.Context, req *v1.UpdateRedemptionCodeRequest) (*v1.UpdateRedemptionCodeReply, error) {
	if err := s.uc.UpdateRedemptionCode(ctx, req); err != nil {
		return nil, err
	}

	return &v1.UpdateRedemptionCodeReply{
		Code:    int32(responsecode.AdminUpdateRedemptionCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateRedemptionCodeSuccess],
		Data: &v1.UpdateRedemptionCodeData{
			Success: true,
		},
	}, nil
}

// ToggleRedemptionCodeStatus 切换兑换码状态
func (s *RedemptionService) ToggleRedemptionCodeStatus(ctx context.Context, req *v1.ToggleRedemptionCodeStatusRequest) (*v1.ToggleRedemptionCodeStatusReply, error) {
	if err := s.uc.ToggleRedemptionCodeStatus(ctx, req); err != nil {
		return nil, err
	}

	return &v1.ToggleRedemptionCodeStatusReply{
		Code:    int32(responsecode.AdminToggleRedemptionCodeStatusSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminToggleRedemptionCodeStatusSuccess],
		Data: &v1.ToggleRedemptionCodeStatusData{
			Success: true,
		},
	}, nil
}

// DeleteRedemptionCode 删除兑换码
func (s *RedemptionService) DeleteRedemptionCode(ctx context.Context, req *v1.DeleteRedemptionCodeRequest) (*v1.DeleteRedemptionCodeReply, error) {
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil || id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if err := s.uc.DeleteRedemptionCode(ctx, id); err != nil {
		return nil, err
	}

	return &v1.DeleteRedemptionCodeReply{
		Code:    int32(responsecode.AdminDeleteRedemptionCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteRedemptionCodeSuccess],
		Data: &v1.DeleteRedemptionCodeData{
			Success: true,
		},
	}, nil
}

// BatchDeleteRedemptionCode 批量删除兑换码
func (s *RedemptionService) BatchDeleteRedemptionCode(ctx context.Context, req *v1.BatchDeleteRedemptionCodeRequest) (*v1.BatchDeleteRedemptionCodeReply, error) {
	ids := make([]int64, 0, len(req.Ids))
	for _, item := range req.Ids {
		id, err := strconv.ParseInt(item, 10, 64)
		if err != nil || id <= 0 {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		ids = append(ids, id)
	}

	if err := s.uc.BatchDeleteRedemptionCode(ctx, ids); err != nil {
		return nil, err
	}

	return &v1.BatchDeleteRedemptionCodeReply{
		Code:    int32(responsecode.AdminBatchDeleteRedemptionCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteRedemptionCodeSuccess],
		Data: &v1.BatchDeleteRedemptionCodeData{
			Success: true,
		},
	}, nil
}

// GetRedemptionCodeList 获取兑换码列表
func (s *RedemptionService) GetRedemptionCodeList(ctx context.Context, req *v1.GetRedemptionCodeListRequest) (*v1.GetRedemptionCodeListReply, error) {
	list, total, err := s.uc.GetRedemptionCodeList(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert ent entities to proto messages
	redemptionCodes := make([]*v1.RedemptionCode, 0, len(list))
	for _, item := range list {
		redemptionCodes = append(redemptionCodes, &v1.RedemptionCode{
			Id:            strconv.FormatInt(item.ID, 10),
			Code:          item.Code,
			TotalCount:    item.TotalCount,
			UsedCount:     item.UsedCount,
			SubscribePlan: item.SubscribePlan,
			UnitTime:      item.UnitTime,
			Quantity:      item.Quantity,
			Status:        int64(item.Status),
			CreatedAt:     item.CreatedAt.Unix(),
			UpdatedAt:     item.UpdatedAt.Unix(),
		})
	}

	return &v1.GetRedemptionCodeListReply{
		Code:    int32(responsecode.AdminGetRedemptionCodeListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetRedemptionCodeListSuccess],
		Data: &v1.GetRedemptionCodeListData{
			Total: total,
			List:  redemptionCodes,
		},
	}, nil
}

// GetRedemptionRecordList 获取兑换记录列表
func (s *RedemptionService) GetRedemptionRecordList(ctx context.Context, req *v1.GetRedemptionRecordListRequest) (*v1.GetRedemptionRecordListReply, error) {
	list, total, err := s.uc.GetRedemptionRecordList(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert ent entities to proto messages
	redemptionRecords := make([]*v1.RedemptionRecord, 0, len(list))
	for _, item := range list {
		redemptionRecords = append(redemptionRecords, &v1.RedemptionRecord{
			Id:               strconv.FormatInt(item.ID, 10),
			RedemptionCodeId: item.RedemptionCodeID,
			UserId:           strconv.FormatInt(item.UserID, 10),
			SubscribeId:      strconv.FormatInt(item.SubscribeID, 10),
			UnitTime:         item.UnitTime,
			Quantity:         item.Quantity,
			RedeemedAt:       item.RedeemedAt.Unix(),
			CreatedAt:        item.CreatedAt.Unix(),
		})
	}

	return &v1.GetRedemptionRecordListReply{
		Code:    int32(responsecode.AdminGetRedemptionRecordListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetRedemptionRecordListSuccess],
		Data: &v1.GetRedemptionRecordListData{
			Total: total,
			List:  redemptionRecords,
		},
	}, nil
}
