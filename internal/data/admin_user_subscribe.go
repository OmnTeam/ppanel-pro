package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystemlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
)

type adminUserSubscribeRepo struct {
	data   *Data
	logger *log.Helper
}

// NewAdminUserSubscribeRepo creates a new admin user subscribe repository
func NewAdminUserSubscribeRepo(d *Data, logger log.Logger) userbiz.SubscribeRepo {
	return &adminUserSubscribeRepo{
		data:   d,
		logger: log.NewHelper(logger),
	}
}

// GetUserSubscribe 获取用户订阅列表
func (r *adminUserSubscribeRepo) GetUserSubscribe(ctx context.Context, req *v1.GetUserSubscribeRequest) ([]*ent.ProxyUserSubscribe, int64, error) {
	query := r.data.db.ProxyUserSubscribe.Query()

	// 用户ID过滤（可选）
	if req.UserId > 0 {
		query = query.Where(proxyusersubscribe.UserIDEQ(int(req.UserId)))
	}

	// 订阅套餐ID过滤（可选）
	if req.SubscribeId > 0 {
		query = query.Where(proxyusersubscribe.SubscribeIDEQ(int(req.SubscribeId)))
	}

	// 状态过滤（可选，0表示所有状态）
	if req.Status > 0 {
		query = query.Where(proxyusersubscribe.StatusEQ(int(req.Status)))
	}

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count user subscribes: %v", err)
		return nil, 0, err
	}

	// 分页查询
	list, err := query.
		Order(ent.Desc(proxyusersubscribe.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query user subscribes: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}

// CreateUserSubscribe 创建用户订阅
func (r *adminUserSubscribeRepo) CreateUserSubscribe(ctx context.Context, req *v1.CreateUserSubscribeRequest) (int64, error) {
	// 验证用户是否存在
	userExists, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.IDEQ(int(req.UserId)),
		).
		Exist(ctx)

	if err != nil {
		r.logger.Errorf("Failed to check user existence: %v", err)
		return 0, err
	}

	if !userExists {
		return 0, responsecode.NewKratosError(responsecode.ErrUserNotExist)
	}

	// TODO: 检查单订阅模式限制
	// 需要从proxy_system表读取single_subscribe_mode配置
	// 如果启用了单订阅模式且用户已有订阅，则返回错误

	// 验证订阅套餐是否存在
	subscribePlan, err := r.data.db.ProxySubscribe.Query().
		Where(
			proxysubscribe.IDEQ(int(req.SubscribeId)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return 0, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query subscribe plan: %v", err)
		return 0, err
	}

	// 如果未指定流量，使用套餐默认流量
	traffic := req.Traffic
	if traffic == 0 {
		traffic = subscribePlan.Traffic
	}

	// 生成Token和UUID
	token := uuidx.SubscribeToken(fmt.Sprintf("adminCreate:%d", time.Now().UnixMilli()))
	subscribeUUID := uuid.New().String()

	// 处理开始时间和过期时间
	startTime := time.Now()
	if req.StartTime > 0 {
		startTime = time.UnixMilli(req.StartTime)
	}
	expireTime := time.UnixMilli(req.ExpireTime)

	// 处理状态
	status := uint8(1) // 默认激活
	if req.Status > 0 {
		status = uint8(req.Status)
	}

	// 创建用户订阅
	created, err := r.data.db.ProxyUserSubscribe.Create().
		SetUserID(int(req.UserId)).
		SetOrderID(int(req.OrderId)).
		SetSubscribeID(int(req.SubscribeId)).
		SetStartTime(startTime).
		SetExpireTime(expireTime).
		SetTraffic(traffic).
		SetDownload(0).
		SetUpload(0).
		SetToken(token).
		SetUUID(subscribeUUID).
		SetStatus(int(status)).
		Save(ctx)

	if err != nil {
		r.logger.Errorf("Failed to create user subscribe: %v", err)
		return 0, err
	}

	// TODO: 清除缓存
	// 1. 清除用户缓存
	// 2. 清除订阅套餐缓存

	return int64(created.ID), nil
}

// UpdateUserSubscribe 更新用户订阅
func (r *adminUserSubscribeRepo) UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) error {
	// 查找用户订阅
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(int(req.Id)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query user subscribe: %v", err)
		return err
	}

	// 处理时间
	startTime := time.UnixMilli(req.StartTime)
	expireTime := time.UnixMilli(req.ExpireTime)

	// 处理状态（如果指定了状态则使用，否则根据过期时间计算）
	var status uint8
	if req.Status > 0 {
		status = uint8(req.Status)
	} else {
		// 自动计算状态
		if time.Now().After(expireTime) {
			status = 3 // 过期
		} else {
			status = 1 // 激活
		}
	}

	// 更新用户订阅
	err = userSub.Update().
		SetStartTime(startTime).
		SetExpireTime(expireTime).
		SetTraffic(req.Traffic).
		SetDownload(req.Download).
		SetUpload(req.Upload).
		SetStatus(int(status)).
		SetUpdatedAt(time.Now()).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to update user subscribe: %v", err)
		return err
	}

	// TODO: 清除缓存
	// 1. 清除用户订阅缓存
	// 2. 清除订阅套餐缓存（新旧两个套餐都需要清除）

	return nil
}

// DeleteUserSubscribe 删除用户订阅
func (r *adminUserSubscribeRepo) DeleteUserSubscribe(ctx context.Context, id int64) error {
	// 查找用户订阅（用于后续清除缓存）
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(int(id)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query user subscribe: %v", err)
		return err
	}

	// 删除用户订阅
	err = r.data.db.ProxyUserSubscribe.DeleteOne(userSub).Exec(ctx)
	if err != nil {
		r.logger.Errorf("Failed to delete user subscribe: %v", err)
		return err
	}

	// TODO: 清除缓存
	// 1. 清除用户订阅缓存
	// 2. 清除订阅套餐缓存

	return nil
}

// GetUserSubscribeById 根据ID获取用户订阅详情（包含套餐信息）
func (r *adminUserSubscribeRepo) GetUserSubscribeById(ctx context.Context, id int64) (*v1.UserSubscribe, string, error) {
	// 查询用户订阅
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(int(id)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query user subscribe: %v", err)
		return nil, "", err
	}

	// 查询订阅套餐信息
	subscribePlan, err := r.data.db.ProxySubscribe.Query().
		Where(
			proxysubscribe.IDEQ(userSub.SubscribeID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query subscribe plan: %v", err)
		return nil, "", err
	}

	// 构建返回结果
	subscribe := &v1.UserSubscribe{
		Id:          int64(userSub.ID),
		TenantId:    0, // 租户已移除，设为0
		UserId:      int64(userSub.UserID),
		OrderId:     int64(userSub.OrderID),
		SubscribeId: int64(userSub.SubscribeID),
		StartTime:   userSub.StartTime.UnixMilli(),
		CreatedAt:   userSub.CreatedAt.UnixMilli(),
		UpdatedAt:   userSub.UpdatedAt.UnixMilli(),
		// 处理指针字段
		Traffic:  getInt64Value(userSub.Traffic),
		Download: getInt64Value(userSub.Download),
		Upload:   getInt64Value(userSub.Upload),
		Token:    getStringValue(userSub.Token),
		Uuid:     getStringValue(userSub.UUID),
		Status:   int32(getIntValue(userSub.Status)),
	}

	// 处理expire_time
	if userSub.ExpireTime != nil {
		subscribe.ExpireTime = userSub.ExpireTime.UnixMilli()
	}

	// 处理finished_at
	if userSub.FinishedAt != nil {
		subscribe.FinishedAt = userSub.FinishedAt.UnixMilli()
	}

	// 套餐名称
	subscribeName := subscribePlan.Name

	return subscribe, subscribeName, nil
}

// GetUserSubscribeDevices 获取用户订阅设备列表
func (r *adminUserSubscribeRepo) GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) ([]*ent.ProxyUserDevice, int64, error) {
	// 查询用户订阅信息以获取 user_id 和 subscribe_id
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(int(req.UserSubscribeId)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query user subscribe: %v", err)
		return nil, 0, err
	}

	// 查询设备列表
	query := r.data.db.ProxyUserDevice.Query().
		Where(
			proxyuserdevice.UserIDEQ(userSub.UserID),
		)

	// 添加subscribe_id过滤
	query = query.Where(proxyuserdevice.SubscribeIDEQ(userSub.SubscribeID))

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count user subscribe devices: %v", err)
		return nil, 0, err
	}

	// 查询所有设备（不分页）
	list, err := query.
		Order(ent.Desc(proxyuserdevice.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query user subscribe devices: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}

// GetUserSubscribeLogs 获取用户订阅日志
func (r *adminUserSubscribeRepo) GetUserSubscribeLogs(ctx context.Context, req *v1.GetUserSubscribeLogsRequest) ([]*ent.ProxySystemLog, int64, error) {
	query := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(logmodel.TypeSubscribe)), // Type = 20
		)

	// 可选：过滤特定用户订阅
	if req.UserSubscribeId > 0 {
		query = query.Where(proxysystemlog.ObjectIDEQ(req.UserSubscribeId))
	}

	// 可选：过滤日期
	if req.Date != "" {
		query = query.Where(proxysystemlog.DateEQ(req.Date))
	}

	// 注意：user_id 在系统日志中没有直接字段，需要通过 object_id 关联查询，这里暂不实现该过滤

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count user subscribe logs: %v", err)
		return nil, 0, err
	}

	// 分页查询
	list, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query user subscribe logs: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}

// GetUserSubscribeResetTrafficLogs 获取用户订阅重置流量日志
func (r *adminUserSubscribeRepo) GetUserSubscribeResetTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeResetTrafficLogsRequest) ([]*ent.ProxySystemLog, int64, error) {
	query := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(logmodel.TypeResetSubscribe)), // Type = 23
		)

	// 可选：过滤特定用户订阅
	if req.UserSubscribeId > 0 {
		query = query.Where(proxysystemlog.ObjectIDEQ(req.UserSubscribeId))
	}

	// 可选：过滤日期
	if req.Date != "" {
		query = query.Where(proxysystemlog.DateEQ(req.Date))
	}

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count reset traffic logs: %v", err)
		return nil, 0, err
	}

	// 分页查询
	list, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query reset traffic logs: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}

// GetUserSubscribeTrafficLogs 获取用户订阅流量日志
func (r *adminUserSubscribeRepo) GetUserSubscribeTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeTrafficLogsRequest) ([]*ent.ProxyTrafficLog, int64, error) {
	query := r.data.db.ProxyTrafficLog.Query()

	// 可选：过滤特定用户
	if req.UserId > 0 {
		query = query.Where(proxytrafficlog.UserIDEQ(req.UserId))
	}

	// 可选：过滤特定用户订阅（需要通过user_subscribe_id查询subscribe_id）
	if req.UserSubscribeId > 0 {
		userSub, err := r.data.db.ProxyUserSubscribe.Query().
			Where(
				proxyusersubscribe.IDEQ(int(req.UserSubscribeId)),
			).
			Only(ctx)
		if err == nil {
			query = query.Where(proxytrafficlog.SubscribeIDEQ(int64(userSub.SubscribeID)))
		}
	}

	// TODO: 可选过滤日期 - traffic_log表没有date字段，需要通过timestamp范围查询实现
	// if req.Date != "" { ... }

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count traffic logs: %v", err)
		return nil, 0, err
	}

	// 分页查询
	list, err := query.
		Order(ent.Desc(proxytrafficlog.FieldTimestamp)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
		All(ctx)

	if err != nil {
		r.logger.Errorf("Failed to query traffic logs: %v", err)
		return nil, 0, err
	}

	return list, int64(total), nil
}

// 辅助函数：获取int64指针的值，nil返回0
func getInt64Value(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// 辅助函数：获取string指针的值，nil返回空字符串
func getStringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// 辅助函数：获取int指针的值，nil返回0
func getIntValue(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// 辅助函数：获取uint8指针的值，nil返回0
func getUint8Value(p *uint8) uint8 {
	if p == nil {
		return 0
	}
	return *p
}
