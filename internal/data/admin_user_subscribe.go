package data

import (
	"context"
	"fmt"
	"strconv"
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
	if req.UserId != "" {
		userID, _ := strconv.ParseInt(req.UserId, 10, 64)
		query = query.Where(proxyusersubscribe.UserIDEQ(userID))
	}

	// 订阅套餐ID过滤（可选）
	if req.SubscribeId != "" {
		subscribeID, _ := strconv.ParseInt(req.SubscribeId, 10, 64)
		query = query.Where(proxyusersubscribe.SubscribeIDEQ(subscribeID))
	}

	// 状态过滤（可选，0表示所有状态）
	if req.Status > 0 {
		query = query.Where(proxyusersubscribe.StatusEQ(int8(req.Status)))
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
	// Parse user ID
	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 验证用户是否存在
	userExists, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.IDEQ(userID),
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

	// Parse subscribe ID
	subscribeID, err := strconv.ParseInt(req.SubscribeId, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 验证订阅套餐是否存在
	subscribePlan, err := r.data.db.ProxySubscribe.Query().
		Where(
			proxysubscribe.IDEQ(subscribeID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return 0, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query subscribe plan: %v", err)
		return 0, err
	}

	// Parse traffic if provided
	var traffic *int64
	if req.Traffic != "" {
		trafficVal, err := strconv.ParseInt(req.Traffic, 10, 64)
		if err != nil {
			return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		traffic = &trafficVal
	} else if subscribePlan.Traffic > 0 {
		traffic = &subscribePlan.Traffic
	}

	// 生成Token和UUID
	tokenStr := uuidx.SubscribeToken(fmt.Sprintf("adminCreate:%d", time.Now().UnixMilli()))
	subscribeUUID := uuid.New().String()

	// 处理开始时间和过期时间
	startTime := time.Now()
	if req.StartTime != "" {
		startTimeVal, _ := strconv.ParseInt(req.StartTime, 10, 64)
		startTime = time.UnixMilli(startTimeVal)
	}
	expireTimeVal, _ := strconv.ParseInt(req.ExpireTime, 10, 64)
	expireTime := time.UnixMilli(expireTimeVal)

	// 处理状态
	var status *int8
	if req.Status > 0 {
		s := int8(req.Status)
		status = &s
	} else {
		s := int8(1) // 默认激活
		status = &s
	}

	// Parse order ID
	orderID, _ := strconv.ParseInt(req.OrderId, 10, 64)

	// 创建用户订阅
	created, err := r.data.db.ProxyUserSubscribe.Create().
		SetUserID(userID).
		SetOrderID(orderID).
		SetSubscribeID(subscribeID).
		SetStartTime(startTime).
		SetExpireTime(expireTime).
		SetNillableTraffic(traffic).
		SetNillableDownload(nil).
		SetNillableUpload(nil).
		SetNillableToken(&tokenStr).
		SetNillableUUID(&subscribeUUID).
		SetNillableStatus(status).
		Save(ctx)

	if err != nil {
		r.logger.Errorf("Failed to create user subscribe: %v", err)
		return 0, err
	}

	// TODO: 清除缓存
	// 1. 清除用户缓存
	// 2. 清除订阅套餐缓存

	return created.ID, nil
}

// UpdateUserSubscribe 更新用户订阅
func (r *adminUserSubscribeRepo) UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) error {
	// Parse ID
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 查找用户订阅
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(id),
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
	startTimeVal, _ := strconv.ParseInt(req.StartTime, 10, 64)
	startTime := time.UnixMilli(startTimeVal)
	expireTimeVal, _ := strconv.ParseInt(req.ExpireTime, 10, 64)
	expireTime := time.UnixMilli(expireTimeVal)

	// 处理状态（如果指定了状态则使用，否则根据过期时间计算）
	var status *int8
	if req.Status > 0 {
		s := int8(req.Status)
		status = &s
	} else {
		// 自动计算状态
		s := int8(1) // 默认激活
		if time.Now().After(expireTime) {
			s = int8(3) // 过期
		}
		status = &s
	}

	// 更新用户订阅
	update := userSub.Update().
		SetStartTime(startTime).
		SetExpireTime(expireTime).
		SetNillableStatus(status).
		SetUpdatedAt(time.Now())

	// 只设置非零值
	if req.Traffic != "" {
		trafficVal, _ := strconv.ParseInt(req.Traffic, 10, 64)
		update.SetNillableTraffic(&trafficVal)
	}
	if req.Download != "" {
		downloadVal, _ := strconv.ParseInt(req.Download, 10, 64)
		update.SetNillableDownload(&downloadVal)
	}
	if req.Upload != "" {
		uploadVal, _ := strconv.ParseInt(req.Upload, 10, 64)
		update.SetNillableUpload(&uploadVal)
	}

	err = update.Exec(ctx)

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
			proxyusersubscribe.IDEQ(id),
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
			proxyusersubscribe.IDEQ(id),
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
		Id:          strconv.FormatInt(int64(userSub.ID), 10),
		UserId:      strconv.FormatInt(int64(userSub.UserID), 10),
		OrderId:     strconv.FormatInt(int64(userSub.OrderID), 10),
		SubscribeId: strconv.FormatInt(int64(userSub.SubscribeID), 10),
		StartTime:   strconv.FormatInt(userSub.StartTime.UnixMilli(), 10),
		CreatedAt:   strconv.FormatInt(userSub.CreatedAt.UnixMilli(), 10),
		UpdatedAt:   strconv.FormatInt(userSub.UpdatedAt.UnixMilli(), 10),
		// 处理指针字段
		Traffic:  formatInt64Value(userSub.Traffic),
		Download: formatInt64Value(userSub.Download),
		Upload:   formatInt64Value(userSub.Upload),
		Token:    getStringValue(userSub.Token),
		Uuid:     getStringValue(userSub.UUID),
		Status:   int32(getInt8Value(userSub.Status)),
	}

	// 处理expire_time
	if userSub.ExpireTime != nil {
		subscribe.ExpireTime = strconv.FormatInt(userSub.ExpireTime.UnixMilli(), 10)
	}

	// 处理finished_at
	if userSub.FinishedAt != nil {
		subscribe.FinishedAt = strconv.FormatInt(userSub.FinishedAt.UnixMilli(), 10)
	}

	// 套餐名称
	subscribeName := subscribePlan.Name

	return subscribe, subscribeName, nil
}

// GetUserSubscribeDevices 获取用户订阅设备列表
func (r *adminUserSubscribeRepo) GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) ([]*ent.ProxyUserDevice, int64, error) {
	// Parse user subscribe ID
	userSubID, err := strconv.ParseInt(req.UserSubscribeId, 10, 64)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 查询用户订阅信息以获取 user_id 和 subscribe_id
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(userSubID),
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
	if req.UserSubscribeId != "" {
		userSubID, _ := strconv.ParseInt(req.UserSubscribeId, 10, 64)
		query = query.Where(proxysystemlog.ObjectIDEQ(userSubID))
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
	if req.UserSubscribeId != "" {
		userSubID, _ := strconv.ParseInt(req.UserSubscribeId, 10, 64)
		query = query.Where(proxysystemlog.ObjectIDEQ(userSubID))
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
	if req.UserId != "" {
		userID, _ := strconv.ParseInt(req.UserId, 10, 64)
		query = query.Where(proxytrafficlog.UserIDEQ(userID))
	}

	// 可选：过滤特定用户订阅（需要通过user_subscribe_id查询subscribe_id）
	if req.UserSubscribeId != "" {
		userSubID, _ := strconv.ParseInt(req.UserSubscribeId, 10, 64)
		userSub, err := r.data.db.ProxyUserSubscribe.Query().
			Where(
				proxyusersubscribe.IDEQ(userSubID),
			).
			Only(ctx)
		if err == nil {
			subscribeId := userSub.SubscribeID
			query = query.Where(proxytrafficlog.SubscribeIDEQ(subscribeId))
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

// 辅助函数：获取int64指针的值并格式化为字符串，nil返回"0"
func formatInt64Value(p *int64) string {
	if p == nil {
		return "0"
	}
	return strconv.FormatInt(*p, 10)
}

// 辅助函数：获取string指针的值，nil返回空字符串
func getStringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// 辅助函数：获取int8指针的值，nil返回0
func getInt8Value(p *int8) int {
	if p == nil {
		return 0
	}
	return int(*p)
}

// 辅助函数：获取uint8指针的值，nil返回0
func getUint8Value(p *uint8) uint8 {
	if p == nil {
		return 0
	}
	return *p
}
