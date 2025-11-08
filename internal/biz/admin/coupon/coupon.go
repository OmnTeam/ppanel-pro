package coupon

import (
	"context"
	"math/rand"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/pkg/snowflake"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
)

// CouponRepo coupon repository interface
type CouponRepo interface {
	// CreateCoupon 创建优惠券
	CreateCoupon(ctx context.Context, tenantID int64, name, code string, count int64, typ int32, discount, startTime, expireTime, userLimit int64, subscribe string, enable bool) error

	// UpdateCoupon 更新优惠券
	UpdateCoupon(ctx context.Context, id, tenantID int64, name, code string, count int64, typ int32, discount, startTime, expireTime, userLimit int64, subscribe string, enable bool) error

	// DeleteCoupon 删除优惠券
	DeleteCoupon(ctx context.Context, id, tenantID int64) error

	// BatchDeleteCoupon 批量删除优惠券
	BatchDeleteCoupon(ctx context.Context, tenantID int64, ids []int64) error

	// GetCouponList 获取优惠券列表
	GetCouponList(ctx context.Context, tenantID, page, size, subscribe int64, search string) ([]*ent.ProxyCoupon, int64, error)

	// FindCouponByCode 根据代码查找优惠券（用于检查重复）
	FindCouponByCode(ctx context.Context, code string) (*ent.ProxyCoupon, error)
}

// CouponUseCase coupon use case
type CouponUseCase struct {
	repo CouponRepo
	log  *log.Helper
}

// NewCouponUseCase creates a new CouponUseCase
func NewCouponUseCase(repo CouponRepo, logger log.Logger) *CouponUseCase {
	return &CouponUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateCoupon 创建优惠券
func (uc *CouponUseCase) CreateCoupon(ctx context.Context, tenantID int64, name, code string, count int64, typ int32, discount, startTime, expireTime, userLimit int64, subscribe []int64, enable bool) error {
	// 如果code为空，自动生成
	if code == "" {
		code = generateCouponCode()
	}

	// 转换订阅ID列表为字符串
	subscribeStr := tool.Int64SliceToString(subscribe)

	return uc.repo.CreateCoupon(ctx, tenantID, name, code, count, typ, discount, startTime, expireTime, userLimit, subscribeStr, enable)
}

// UpdateCoupon 更新优惠券
func (uc *CouponUseCase) UpdateCoupon(ctx context.Context, id, tenantID int64, name, code string, count int64, typ int32, discount, startTime, expireTime, userLimit int64, subscribe []int64, enable bool) error {
	// 转换订阅ID列表为字符串
	subscribeStr := tool.Int64SliceToString(subscribe)

	return uc.repo.UpdateCoupon(ctx, id, tenantID, name, code, count, typ, discount, startTime, expireTime, userLimit, subscribeStr, enable)
}

// DeleteCoupon 删除优惠券
func (uc *CouponUseCase) DeleteCoupon(ctx context.Context, id, tenantID int64) error {
	return uc.repo.DeleteCoupon(ctx, id, tenantID)
}

// BatchDeleteCoupon 批量删除优惠券
func (uc *CouponUseCase) BatchDeleteCoupon(ctx context.Context, tenantID int64, ids []int64) error {
	return uc.repo.BatchDeleteCoupon(ctx, tenantID, ids)
}

// GetCouponList 获取优惠券列表
func (uc *CouponUseCase) GetCouponList(ctx context.Context, tenantID, page, size, subscribe int64, search string) ([]*ent.ProxyCoupon, int64, error) {
	return uc.repo.GetCouponList(ctx, tenantID, page, size, subscribe, search)
}

// generateCouponCode 生成优惠券代码
// 使用与原系统相同的算法：随机前缀 + 雪花ID的Base36编码
// 格式: XXXX-YYYY-... (每4个字符插入短横线)
func generateCouponCode() string {
	rand.Seed(time.Now().UnixNano())
	// 生成4位随机前缀（数字+大写字母）
	prefix := tool.KeyNew(4, 2)
	// 获取雪花ID并转为Base36编码
	sid := snowflake.GetID()
	encoded := tool.EncodeBase36(sid)
	// 每4个字符插入短横线
	return prefix + "-" + tool.StrToDashedString(encoded)
}
