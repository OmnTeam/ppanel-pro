package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	logbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/log"
	"github.com/go-kratos/kratos/v2/log"
)

type adminTrafficLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminTrafficLogRepo 创建流量日志仓库
func NewAdminTrafficLogRepo(data *Data, logger log.Logger) logbiz.TrafficLogRepo {
	return &adminTrafficLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// FilterTrafficLogDetails 过滤流量日志详情
func (r *adminTrafficLogRepo) FilterTrafficLogDetails(ctx context.Context, tenantID int64, page, size int32, date string, serverID, userID, subscribeID *int64) ([]*ent.ProxyTrafficLog, int64, error) {
	// 设置默认值
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 20
	}

	// 解析日期范围
	var start, end time.Time
	if date != "" {
		day, err := time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			return nil, 0, err
		}
		start = day
		end = day.Add(24*time.Hour - time.Nanosecond)
	} else {
		// 查询今天
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24*time.Hour - time.Nanosecond)
	}

	// 构建查询
	query := r.data.db.ProxyTrafficLog.
		Query().
		Where(func(s *sql.Selector) {
			// 时间范围过滤
			s.Where(sql.And(
				sql.GTE(s.C(proxytrafficlog.FieldTimestamp), start),
				sql.LTE(s.C(proxytrafficlog.FieldTimestamp), end),
			))

			// 时间范围过滤
			s.Where(sql.And(
				sql.GTE(s.C(proxytrafficlog.FieldTimestamp), start),
				sql.LTE(s.C(proxytrafficlog.FieldTimestamp), end),
			))

			// 服务器ID过滤
			if serverID != nil && *serverID > 0 {
				s.Where(sql.EQ(s.C(proxytrafficlog.FieldServerID), *serverID))
			}

			// 用户ID过滤
			if userID != nil && *userID > 0 {
				s.Where(sql.EQ(s.C(proxytrafficlog.FieldUserID), *userID))
			}

			// 订阅ID过滤
			if subscribeID != nil && *subscribeID > 0 {
				s.Where(sql.EQ(s.C(proxytrafficlog.FieldSubscribeID), *subscribeID))
			}
		})

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	list, err := query.
		Order(ent.Desc(proxytrafficlog.FieldTimestamp)).
		Limit(int(size)).
		Offset(int((page - 1) * size)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}
