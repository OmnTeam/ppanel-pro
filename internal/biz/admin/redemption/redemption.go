package redemption

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/redemption/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// RedemptionRepo redemption repository interface
type RedemptionRepo interface {
	// CreateRedemptionCode creates redemption codes in batch
	CreateRedemptionCode(ctx context.Context, req *v1.CreateRedemptionCodeRequest) (int64, error)
	// UpdateRedemptionCode updates redemption code
	UpdateRedemptionCode(ctx context.Context, req *v1.UpdateRedemptionCodeRequest) error
	// ToggleRedemptionCodeStatus toggles redemption code status
	ToggleRedemptionCodeStatus(ctx context.Context, req *v1.ToggleRedemptionCodeStatusRequest) error
	// DeleteRedemptionCode deletes redemption code
	DeleteRedemptionCode(ctx context.Context, id int64) error
	// BatchDeleteRedemptionCode batch deletes redemption codes
	BatchDeleteRedemptionCode(ctx context.Context, ids []int64) error
	// GetRedemptionCodeList gets redemption code list
	GetRedemptionCodeList(ctx context.Context, req *v1.GetRedemptionCodeListRequest) ([]*ent.ProxyRedemptionCode, int64, error)
	// GetRedemptionRecordList gets redemption record list
	GetRedemptionRecordList(ctx context.Context, req *v1.GetRedemptionRecordListRequest) ([]*ent.ProxyRedemptionRecord, int64, error)
}

// RedemptionUseCase redemption use case
type RedemptionUseCase struct {
	repo RedemptionRepo
	log  *log.Helper
}

// NewRedemptionUseCase creates a new redemption use case
func NewRedemptionUseCase(repo RedemptionRepo, logger log.Logger) *RedemptionUseCase {
	return &RedemptionUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/redemption")),
	}
}

// CreateRedemptionCode creates redemption codes in batch
func (uc *RedemptionUseCase) CreateRedemptionCode(ctx context.Context, req *v1.CreateRedemptionCodeRequest) (int64, error) {
	// Validate request parameters
	if req.TotalCount <= 0 {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}
	if req.SubscribePlan <= 0 {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}
	if req.UnitTime == "" {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}
	if req.Quantity <= 0 {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}
	if req.BatchCount < 1 {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	// Validate unit_time value
	validUnitTimes := map[string]bool{
		"day":       true,
		"month":     true,
		"quarter":   true,
		"half_year": true,
		"year":      true,
	}
	if !validUnitTimes[req.UnitTime] {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	// Call repository to create redemption codes
	createdCount, err := uc.repo.CreateRedemptionCode(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to create redemption codes: %v", err)
		return 0, err
	}

	return createdCount, nil
}

// UpdateRedemptionCode updates redemption code
func (uc *RedemptionUseCase) UpdateRedemptionCode(ctx context.Context, req *v1.UpdateRedemptionCodeRequest) error {
	// Validate request parameters
	if req.Id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	// Validate unit_time if provided
	if req.UnitTime != "" {
		validUnitTimes := map[string]bool{
			"day":       true,
			"month":     true,
			"quarter":   true,
			"half_year": true,
			"year":      true,
		}
		if !validUnitTimes[req.UnitTime] {
			return responsecode.NewKratosError(responsecode.ErrInvalidParams)
		}
	}

	// Validate status if provided
	if req.Status < 0 || req.Status > 1 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	// Call repository to update redemption code
	err := uc.repo.UpdateRedemptionCode(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to update redemption code: %v", err)
		return err
	}

	return nil
}

// ToggleRedemptionCodeStatus toggles redemption code status
func (uc *RedemptionUseCase) ToggleRedemptionCodeStatus(ctx context.Context, req *v1.ToggleRedemptionCodeStatusRequest) error {
	// Validate request parameters
	if req.Id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	if req.Status != 0 && req.Status != 1 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	// Call repository to toggle status
	err := uc.repo.ToggleRedemptionCodeStatus(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to toggle redemption code status: %v", err)
		return err
	}

	return nil
}

// DeleteRedemptionCode deletes redemption code
func (uc *RedemptionUseCase) DeleteRedemptionCode(ctx context.Context, id int64) error {
	if id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	err := uc.repo.DeleteRedemptionCode(ctx, id)
	if err != nil {
		uc.log.Errorf("Failed to delete redemption code: %v", err)
		return err
	}

	return nil
}

// BatchDeleteRedemptionCode batch deletes redemption codes
func (uc *RedemptionUseCase) BatchDeleteRedemptionCode(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	err := uc.repo.BatchDeleteRedemptionCode(ctx, ids)
	if err != nil {
		uc.log.Errorf("Failed to batch delete redemption codes: %v", err)
		return err
	}

	return nil
}

// GetRedemptionCodeList gets redemption code list
func (uc *RedemptionUseCase) GetRedemptionCodeList(ctx context.Context, req *v1.GetRedemptionCodeListRequest) ([]*ent.ProxyRedemptionCode, int64, error) {
	if req.Page <= 0 || req.Size <= 0 {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	list, total, err := uc.repo.GetRedemptionCodeList(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to get redemption code list: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

// GetRedemptionRecordList gets redemption record list
func (uc *RedemptionUseCase) GetRedemptionRecordList(ctx context.Context, req *v1.GetRedemptionRecordListRequest) ([]*ent.ProxyRedemptionRecord, int64, error) {
	if req.Page <= 0 || req.Size <= 0 {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	list, total, err := uc.repo.GetRedemptionRecordList(ctx, req)
	if err != nil {
		uc.log.Errorf("Failed to get redemption record list: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}
