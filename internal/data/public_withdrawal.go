package data

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserwithdrawal"
	"github.com/OmnTeam/ppanel-pro/internal/biz/public/withdrawal"
	"github.com/go-kratos/kratos/v2/log"
)

// withdrawalRepo 提现数据仓储
type withdrawalRepo struct {
	data *Data
	log  *log.Helper
}

// NewWithdrawalRepo 创建提现数据仓储
func NewWithdrawalRepo(data *Data, logger log.Logger) withdrawal.WithdrawalRepo {
	return &withdrawalRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateWithdrawal 创建提现记录
func (r *withdrawalRepo) CreateWithdrawal(ctx context.Context, userID int64, amount int64, content string) error {
	_, err := r.data.db.ProxyUserWithdrawal.Create().
		SetUserID(userID).
		SetAmount(amount).
		SetContent(content).
		SetStatus(0).
		SetReason("").
		Save(ctx)
	return err
}

// GetWithdrawalByID 根据ID获取提现记录
func (r *withdrawalRepo) GetWithdrawalByID(ctx context.Context, id int64) (*withdrawal.Withdrawal, error) {
	entity, err := r.data.db.ProxyUserWithdrawal.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.convertToModel(entity), nil
}

// GetUserWithdrawals 获取用户提现记录列表
func (r *withdrawalRepo) GetUserWithdrawals(ctx context.Context, userID int64, page, pageSize int32) ([]*withdrawal.Withdrawal, int, error) {
	query := r.data.db.ProxyUserWithdrawal.Query().
		Where(proxyuserwithdrawal.UserID(userID)).
		Order(ent.Desc(proxyuserwithdrawal.FieldCreatedAt))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	entities, err := query.
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	withdrawals := make([]*withdrawal.Withdrawal, len(entities))
	for i, entity := range entities {
		withdrawals[i] = r.convertToModel(entity)
	}

	return withdrawals, total, nil
}

// UpdateUserCommission 更新用户佣金
func (r *withdrawalRepo) UpdateUserCommission(ctx context.Context, userID int64, commission int64) error {
	return r.data.db.ProxyUser.UpdateOneID(userID).
		SetCommission(commission).
		Exec(ctx)
}

// GetUserByID 根据ID获取用户
func (r *withdrawalRepo) GetUserByID(ctx context.Context, userID int64) (*withdrawal.User, error) {
	entity, err := r.data.db.ProxyUser.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	commission := int64(0)
	if entity.Commission != nil {
		commission = *entity.Commission
	}
	return &withdrawal.User{
		ID:         entity.ID,
		Commission: commission,
	}, nil
}

// convertToModel 转换为业务模型
func (r *withdrawalRepo) convertToModel(entity *ent.ProxyUserWithdrawal) *withdrawal.Withdrawal {
	content := ""
	if entity.Content != nil {
		content = *entity.Content
	}
	reason := ""
	if entity.Reason != nil {
		reason = *entity.Reason
	}
	return &withdrawal.Withdrawal{
		ID:        entity.ID,
		UserID:    entity.UserID,
		Amount:    entity.Amount,
		Content:   content,
		Status:    int8(entity.Status),
		Reason:    reason,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
