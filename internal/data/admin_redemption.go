package data

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/redemption/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyredemptioncode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyredemptionrecord"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	redemptionbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/redemption"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type adminRedemptionRepo struct {
	data   *Data
	logger *log.Helper
}

// NewAdminRedemptionRepo creates a new admin redemption repository
func NewAdminRedemptionRepo(d *Data, logger log.Logger) redemptionbiz.RedemptionRepo {
	return &adminRedemptionRepo{
		data:   d,
		logger: log.NewHelper(logger),
	}
}

// generateUniqueCode generates a unique 16-character redemption code
func (r *adminRedemptionRepo) generateUniqueCode(ctx context.Context) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed confusing characters
	const codeLength = 16

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		code := make([]byte, codeLength)
		for j := 0; j < codeLength; j++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", fmt.Errorf("failed to generate random number: %w", err)
			}
			code[j] = charset[num.Int64()]
		}

		codeStr := string(code)

		// Check if code already exists
		_, err := r.data.db.ProxyRedemptionCode.Query().
			Where(proxyredemptioncode.Code(codeStr)).
			Only(ctx)

		if ent.IsNotFound(err) {
			return codeStr, nil
		} else if err != nil {
			return "", fmt.Errorf("failed to check code existence: %w", err)
		}
		// Code exists, try again
	}

	return "", fmt.Errorf("failed to generate unique code after %d retries", maxRetries)
}

// CreateRedemptionCode creates redemption codes in batch
func (r *adminRedemptionRepo) CreateRedemptionCode(ctx context.Context, req *v1.CreateRedemptionCodeRequest) (int64, error) {
	// Verify subscribe plan exists
	_, err := r.data.db.ProxySubscribe.Query().
		Where(proxysubscribe.IDEQ(req.SubscribePlan)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Errorf("Subscribe plan not found: %d", req.SubscribePlan)
			return 0, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query subscribe plan: %v", err)
		return 0, err
	}

	// Validate batch count
	if req.BatchCount < 1 {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParams)
	}

	// Generate redemption codes in batch
	var createdCodes int64
	now := time.Now()

	for i := int64(0); i < req.BatchCount; i++ {
		code, err := r.generateUniqueCode(ctx)
		if err != nil {
			r.logger.Errorf("Failed to generate unique code: %v", err)
			return 0, responsecode.NewKratosError(responsecode.ErrInternalError)
		}

		_, err = r.data.db.ProxyRedemptionCode.Create().
			SetCode(code).
			SetTotalCount(req.TotalCount).
			SetUsedCount(0).
			SetSubscribePlan(req.SubscribePlan).
			SetUnitTime(req.UnitTime).
			SetQuantity(req.Quantity).
			SetStatus(1).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)

		if err != nil {
			r.logger.Errorf("Failed to create redemption code: %v", err)
			return 0, err
		}

		createdCodes++
	}

	return createdCodes, nil
}

// UpdateRedemptionCode updates redemption code
func (r *adminRedemptionRepo) UpdateRedemptionCode(ctx context.Context, req *v1.UpdateRedemptionCodeRequest) error {
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil || id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// Find redemption code
	code, err := r.data.db.ProxyRedemptionCode.Query().
		Where(proxyredemptioncode.IDEQ(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrRedemptionCodeNotFound)
		}
		r.logger.Errorf("Failed to query redemption code: %v", err)
		return err
	}

	// Build update
	update := code.Update()

	if req.TotalCount > 0 {
		if req.TotalCount < code.UsedCount {
			return responsecode.NewKratosError(responsecode.ErrInvalidParams)
		}
		update.SetTotalCount(req.TotalCount)
	}
	if req.SubscribePlan > 0 {
		// Verify subscribe plan exists
		_, err := r.data.db.ProxySubscribe.Query().
			Where(proxysubscribe.IDEQ(req.SubscribePlan)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
			}
			return err
		}
		update.SetSubscribePlan(req.SubscribePlan)
	}
	if req.UnitTime != "" {
		update.SetUnitTime(req.UnitTime)
	}
	if req.Quantity > 0 {
		update.SetQuantity(req.Quantity)
	}

	update.SetUpdatedAt(time.Now())

	err = update.Exec(ctx)
	if err != nil {
		r.logger.Errorf("Failed to update redemption code: %v", err)
		return err
	}

	return nil
}

// ToggleRedemptionCodeStatus toggles redemption code status
func (r *adminRedemptionRepo) ToggleRedemptionCodeStatus(ctx context.Context, req *v1.ToggleRedemptionCodeStatusRequest) error {
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil || id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// Find redemption code
	code, err := r.data.db.ProxyRedemptionCode.Query().
		Where(proxyredemptioncode.IDEQ(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrRedemptionCodeNotFound)
		}
		r.logger.Errorf("Failed to query redemption code: %v", err)
		return err
	}

	// Update status
	err = code.Update().
		SetStatus(int8(req.Status)).
		SetUpdatedAt(time.Now()).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to toggle redemption code status: %v", err)
		return err
	}

	return nil
}

// DeleteRedemptionCode deletes redemption code
func (r *adminRedemptionRepo) DeleteRedemptionCode(ctx context.Context, id int64) error {
	_, err := r.data.db.ProxyRedemptionCode.Delete().
		Where(proxyredemptioncode.IDEQ(id)).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to delete redemption code: %v", err)
		return err
	}

	return nil
}

// BatchDeleteRedemptionCode batch deletes redemption codes
func (r *adminRedemptionRepo) BatchDeleteRedemptionCode(ctx context.Context, ids []int64) error {
	_, err := r.data.db.ProxyRedemptionCode.Delete().
		Where(proxyredemptioncode.IDIn(ids...)).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to batch delete redemption codes: %v", err)
		return err
	}

	return nil
}

// GetRedemptionCodeList gets redemption code list with pagination
func (r *adminRedemptionRepo) GetRedemptionCodeList(ctx context.Context, req *v1.GetRedemptionCodeListRequest) ([]*ent.ProxyRedemptionCode, int64, error) {
	query := r.data.db.ProxyRedemptionCode.Query()

	// Optional filters
	if req.SubscribePlan > 0 {
		query = query.Where(proxyredemptioncode.SubscribePlanEQ(req.SubscribePlan))
	}
	if req.UnitTime != "" {
		query = query.Where(proxyredemptioncode.UnitTimeEQ(req.UnitTime))
	}
	if req.Code != "" {
		// Partial match search
		query = query.Where(proxyredemptioncode.CodeContains(req.Code))
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count redemption codes: %v", err)
		return nil, 0, err
	}

	// Get paginated list
	list, err := query.
		Order(ent.Desc(proxyredemptioncode.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query redemption codes: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}

// GetRedemptionRecordList gets redemption record list with pagination
func (r *adminRedemptionRepo) GetRedemptionRecordList(ctx context.Context, req *v1.GetRedemptionRecordListRequest) ([]*ent.ProxyRedemptionRecord, int64, error) {
	query := r.data.db.ProxyRedemptionRecord.Query()

	// Optional filters
	if req.UserId != "" {
		userID, err := strconv.ParseInt(req.UserId, 10, 64)
		if err != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		query = query.Where(proxyredemptionrecord.UserIDEQ(userID))
	}
	if req.CodeId > 0 {
		query = query.Where(proxyredemptionrecord.RedemptionCodeIDEQ(req.CodeId))
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count redemption records: %v", err)
		return nil, 0, err
	}

	// Get paginated list
	list, err := query.
		Order(ent.Desc(proxyredemptionrecord.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query redemption records: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}
