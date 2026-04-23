package data

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/OmnTeam/ppanel-pro/ent"
	portalBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/portal"
)

// SubscribeDiscount 订阅折扣配置
//type SubscribeDiscount struct {
//	Quantity int`json:"quantity"` // 购买数量
//	Discount int`json:"discount"` // 折扣值（百分比 0-100）
//}

// parseSubscribeDiscounts 解析订阅折扣配置
func parseSubscribeDiscounts(discountJSON string) []SubscribeDiscount {
	if discountJSON == "" {
		return nil
	}
	var discounts []SubscribeDiscount
	_ = json.Unmarshal([]byte(discountJSON), &discounts)
	return discounts
}

func parsePortalTrafficLimits(trafficLimitJSON string) []portalBiz.TrafficLimit {
	if trafficLimitJSON == "" {
		return []portalBiz.TrafficLimit{}
	}
	var limits []portalBiz.TrafficLimit
	_ = json.Unmarshal([]byte(trafficLimitJSON), &limits)
	return limits
}

// stringToInt64Slice 将逗号分隔的字符串转换为int64切片
// 例如："1,2,3" -> []int64{1, 2, 3}
func stringToInt64Slice(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if val, err := strconv.ParseInt(part, 10, 64); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// getDiscount 计算订阅折扣率
// ⚠️ 复刻原项目逻辑（portal/tool.go line 9-18）
// 返回折扣率（如0.95表示95折，即5%折扣）
func getDiscount(discounts []SubscribeDiscount, quantity int64) float64 {
	var finalDiscount int64 = 100

	// 找到适用的最小折扣百分比（最大优惠）
	for _, discount := range discounts {
		if quantity >= discount.Quantity && discount.Discount < finalDiscount {
			finalDiscount = discount.Discount
		}
	}

	return float64(finalDiscount) / float64(100)
}

// calculateCoupon 计算优惠券折扣金额
// ⚠️ 完全复刻原项目逻辑（portal/tool.go line 20-26）
// type: 1=百分比, 2=固定金额
func calculateCoupon(amount int64, coupon *ent.ProxyCoupon) int {
	if coupon.Type == 1 {
		// 百分比折扣 - 使用float64计算避免精度问题
		return int(float64(amount) * (float64(coupon.Discount) / float64(100)))
	} else {
		// 固定金额 - 取折扣值和订单金额的最小值
		if coupon.Discount < amount {
			return int(coupon.Discount)
		}
		return int(amount)
	}
}

// calculateFee 计算支付手续费
// feeMode: 0=无费用, 1=百分比, 2=固定金额, 3=百分比+固定金额
func calculateFee(amount int64, payment *ent.ProxyPayment) int {
	var fee float64

	switch payment.FeeMode {
	case 0:
		// 无手续费
		return 0
	case 1:
		// 百分比手续费
		fee = float64(amount) * (float64(payment.FeePercent) / float64(100))
	case 2:
		// 固定金额手续费
		if amount > 0 {
			fee = float64(payment.FeeAmount)
		}
	case 3:
		// 百分比 + 固定金额
		fee = float64(amount)*(float64(payment.FeePercent)/float64(100)) + float64(payment.FeeAmount)
	}

	return int(fee)
}
