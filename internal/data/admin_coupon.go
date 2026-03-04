package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxycoupon"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/coupon"
)

const couponModule = "data/admin_coupon"

type couponRepo struct {
	data *Data
	log  *log.Helper
}

// NewCouponRepo create coupon repository
func NewCouponRepo(data *Data, logger log.Logger) coupon.CouponRepo {
	return &couponRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", couponModule)),
	}
}

// CreateCoupon 创建优惠券
func (r *couponRepo) CreateCoupon(ctx context.Context, name, code string, count int, typ int32, discount, startTime, expireTime, userLimit int64, subscribe string, enable bool) error {
	_, err := r.data.db.ProxyCoupon.Create().
		SetName(name).
		SetCode(code).
		SetCount(int64(count)).
		SetType(int8(typ)).
		SetDiscount(discount).
		SetStartTime(time.Unix(startTime, 0)).
		SetEndTime(time.Unix(expireTime, 0)).
		SetUserLimit(userLimit).
		SetSubscribe(subscribe).
		SetStatus(func() int8 {
			if enable {
				return 1
			} else {
				return 0
			}
		}()).
		Save(ctx)

	return err
}

// UpdateCoupon 更新优惠券
func (r *couponRepo) UpdateCoupon(ctx context.Context, id int, name, code string, count int, typ int32, discount, startTime, expireTime, userLimit int64, subscribe string, enable bool) error {
	return r.data.db.ProxyCoupon.Update().
		Where(
			proxycoupon.ID(int64(id)),
		).
		SetName(name).
		SetCode(code).
		SetCount(int64(count)).
		SetType(int8(typ)).
		SetDiscount(discount).
		SetStartTime(time.Unix(startTime, 0)).
		SetEndTime(time.Unix(expireTime, 0)).
		SetUserLimit(userLimit).
		SetSubscribe(subscribe).
		SetStatus(func() int8 {
			if enable {
				return 1
			} else {
				return 0
			}
		}()).
		Exec(ctx)
}

// DeleteCoupon 删除优惠券
func (r *couponRepo) DeleteCoupon(ctx context.Context, id int) error {
	_, err := r.data.db.ProxyCoupon.Delete().
		Where(
			proxycoupon.ID(int64(id)),
		).
		Exec(ctx)
	return err
}

// BatchDeleteCoupon 批量删除优惠券
func (r *couponRepo) BatchDeleteCoupon(ctx context.Context, ids []int) error {
	// Convert int IDs to int64
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	_, err := r.data.db.ProxyCoupon.Delete().
		Where(
			proxycoupon.IDIn(int64IDs...),
		).
		Exec(ctx)
	return err
}

// GetCouponList 获取优惠券列表
func (r *couponRepo) GetCouponList(ctx context.Context, page, size, subscribe int64, search string) ([]*ent.ProxyCoupon, int64, error) {
	query := r.data.db.ProxyCoupon.Query()

	// 如果有搜索关键字
	if search != "" {
		query = query.Where(
			proxycoupon.Or(
				proxycoupon.NameContains(search),
				proxycoupon.CodeContains(search),
			),
		)
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	list, err := query.
		Limit(int(size)).
		Offset(int((page - 1) * size)).
		Order(ent.Desc(proxycoupon.FieldID)).
		All(ctx)

	return list, int64(total), err
}

// FindCouponByCode 根据代码查找优惠券
func (r *couponRepo) FindCouponByCode(ctx context.Context, code string) (*ent.ProxyCoupon, error) {
	return r.data.db.ProxyCoupon.Query().
		Where(proxycoupon.Code(code)).
		Only(ctx)
}
