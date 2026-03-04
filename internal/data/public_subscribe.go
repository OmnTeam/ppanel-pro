package data

import (
	"context"
	"encoding/json"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	subscribeBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscribe"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type publicSubscribeRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicSubscribeRepo 创建Public Subscribe仓库
func NewPublicSubscribeRepo(data *Data, logger log.Logger) subscribeBiz.SubscribeRepo {
	return &publicSubscribeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// QuerySubscribeList 查询订阅列表
func (r *publicSubscribeRepo) QuerySubscribeList(ctx context.Context, language string) ([]*subscribeBiz.Subscribe, int64, error) {
	// 查询条件: sell=true
	query := r.data.db.ProxySubscribe.Query().
		Where(
			proxysubscribe.Sell(true),
		)

	// 语言过滤
	if language != "" {
		query = query.Where(
			proxysubscribe.Or(
				proxysubscribe.Language(language),
				proxysubscribe.Language(""),
			),
		)
	}

	// 查询
	subscribes, err := query.Order(ent.Asc(proxysubscribe.FieldSort)).All(ctx)
	if err != nil {
		r.log.Errorf("QuerySubscribeList query error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	total := int64(len(subscribes))
	result := make([]*subscribeBiz.Subscribe, 0, len(subscribes))

	for _, s := range subscribes {
		// 处理Description（指针类型）
		desc := ""
		if s.Description != nil {
			desc = *s.Description
		}

		// 处理ResetCycle和DeductionRatio（指针类型）
		resetCycle := int32(0)
		deductionRatio := int64(0)
		if s.ResetCycle != nil {
			resetCycle = int32(*s.ResetCycle)
		}
		if s.DeductionRatio != nil {
			deductionRatio = int64(*s.DeductionRatio)
		}

		// 处理Nodes（JSON数组）
		var nodes []int
		if s.Nodes != "" {
			_ = json.Unmarshal([]byte(s.Nodes), &nodes)
		}

		// 处理NodeTags（JSON数组）
		var nodeTags []string
		if s.NodeTags != "" {
			_ = json.Unmarshal([]byte(s.NodeTags), &nodeTags)
		}

		item := &subscribeBiz.Subscribe{
			ID:             int64(s.ID),
			Name:           s.Name,
			Language:       s.Language,
			Description:    desc,
			UnitPrice:      s.UnitPrice,
			UnitTime:       s.UnitTime,
			Replacement:    int64(s.Replacement),
			Inventory:      int64(s.Inventory),
			Traffic:        s.Traffic,
			SpeedLimit:     int64(s.SpeedLimit),
			DeviceLimit:    int64(s.DeviceLimit),
			Quota:          int64(s.Quota),
			Nodes:          nodes,
			NodeTags:       nodeTags,
			Show:           s.Show,
			Sell:           s.Sell,
			Sort:           int64(s.Sort),
			DeductionRatio: deductionRatio,
			AllowDeduction: s.AllowDeduction,
			ResetCycle:     resetCycle,
			CreatedAt:      s.CreatedAt.UnixMilli(),
			UpdatedAt:      s.UpdatedAt.UnixMilli(),
		}

		// 解析Discount字段（指针类型）
		if s.Discount != nil && *s.Discount != "" {
			var discounts []*subscribeBiz.SubscribeDiscount
			if err := json.Unmarshal([]byte(*s.Discount), &discounts); err == nil {
				item.Discount = discounts
			}
		}

		result = append(result, item)
	}

	return result, total, nil
}
