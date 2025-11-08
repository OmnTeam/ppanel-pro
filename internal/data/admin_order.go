package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyorder"
	"github.com/OmnTeam/ppanel-pro/ent/proxypayment"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/order"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

const orderModule = "data/admin_order"

type orderRepo struct {
	data *Data
	log  *log.Helper
}

// NewOrderRepo create order repository
func NewOrderRepo(data *Data, logger log.Logger) order.OrderRepo {
	return &orderRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", orderModule)),
	}
}

// CreateOrder 创建订单
func (r *orderRepo) CreateOrder(ctx context.Context, tenantID, userID int64, orderType int32, quantity, price, amount, discount int64,
	coupon string, couponDiscount, commission, feeAmount, paymentID int64, tradeNo string,
	status int32, subscribeID int64) error {

	// 如果paymentID > 0，验证支付方式是否存在并获取token作为method
	var method string
	if paymentID > 0 {
		payment, err := r.data.db.ProxyPayment.Query().
			Where(
				proxypayment.ID(int(paymentID)),
			).
			Only(ctx)
		if err != nil {
			r.log.Errorw("msg", "payment method not found", "error", err, "paymentID", paymentID)
			return err
		}
		method = payment.Token
	}

	// 生成订单号
	orderNo := tool.GenerateTradeNo()

	// 创建订单
	_, err := r.data.db.ProxyOrder.Create().
		SetUserID(userID).
		SetOrderNo(orderNo).
		SetType(int8(orderType)).
		SetQuantity(quantity).
		SetPrice(price).
		SetAmount(amount).
		SetDiscount(discount).
		SetCoupon(coupon).
		SetCouponDiscount(couponDiscount).
		SetCommission(commission).
		SetFeeAmount(feeAmount).
		SetPaymentID(paymentID).
		SetMethod(method).
		SetTradeNo(tradeNo).
		SetStatus(int8(status)).
		SetSubscribeID(subscribeID).
		Save(ctx)

	return err
}

// UpdateOrderStatus 更新订单状态
func (r *orderRepo) UpdateOrderStatus(ctx context.Context, id, tenantID int64, status int32, paymentID int64, tradeNo string) error {
	// 查询订单
	orderInfo, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.ID(id),
		).
		Only(ctx)
	if err != nil {
		r.log.Errorw("msg", "order not found", "error", err, "orderID", id)
		return err
	}

	// 使用事务执行更新
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	updater := tx.ProxyOrder.UpdateOneID(id)

	// 更新状态
	updater = updater.SetStatus(int8(status))

	// 如果paymentID > 0，更新支付方式和method
	if paymentID > 0 {
		payment, err := tx.ProxyPayment.Query().
			Where(
				proxypayment.ID(int(paymentID)),
			).
			Only(ctx)
		if err != nil {
			r.log.Errorw("msg", "payment method not found", "error", err, "paymentID", paymentID)
			return rollback(tx, err)
		}
		updater = updater.SetPaymentID(paymentID).SetMethod(payment.Token)
	}

	// 如果提供了tradeNo，更新交易号
	if tradeNo != "" {
		updater = updater.SetTradeNo(tradeNo)
	}

	// 执行更新
	if err := updater.Exec(ctx); err != nil {
		r.log.Errorw("msg", "update order failed", "error", err, "orderID", id)
		return rollback(tx, err)
	}

	// TODO: 如果订单状态变为2(已付款)，需要入队ForthwithActivateOrder任务
	// 任务负载: {"order_no": orderInfo.OrderNo}
	// 任务类型: forthwith:activate:order
	if status == 2 {
		r.log.Infow("msg", "order status changed to paid, should enqueue activate task", "orderNo", orderInfo.OrderNo)
		// payload := map[string]string{"order_no": orderInfo.OrderNo}
		// task := asynq.NewTask(types.ForthwithActivateOrder, jsonPayload)
		// r.data.queue.EnqueueContext(ctx, task)
	}

	return tx.Commit()
}

// GetOrderList 获取订单列表
func (r *orderRepo) GetOrderList(ctx context.Context, tenantID, page, size, userID int64, status int32, subscribeID int64, search string) ([]*ent.ProxyOrder, int64, error) {
	query := r.data.db.ProxyOrder.Query()

	// 用户ID筛选
	if userID != 0 {
		query = query.Where(proxyorder.UserID(userID))
	}

	// 订单状态筛选
	if status != 0 {
		query = query.Where(proxyorder.Status(int8(status)))
	}

	// 订阅ID筛选
	if subscribeID != 0 {
		query = query.Where(proxyorder.SubscribeID(subscribeID))
	}

	// 搜索关键字（订单号、交易号或优惠券）
	if search != "" {
		query = query.Where(
			proxyorder.Or(
				proxyorder.OrderNoContains(search),
				proxyorder.TradeNoContains(search),
				proxyorder.CouponContains(search),
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
		Order(ent.Desc(proxyorder.FieldID)).
		All(ctx)

	return list, int64(total), err
}
