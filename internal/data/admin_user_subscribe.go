package data

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystemlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
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

func (r *adminUserSubscribeRepo) isSingleSubscribeModeEnabled(ctx context.Context) bool {
	enabled := r.data.conf != nil && r.data.conf.Subscribe != nil && r.data.conf.Subscribe.SingleModel

	values, err := loadSystemConfigMap(ctx, r.data.db, "subscribe")
	if err == nil {
		enabled = systemConfigBool(values, enabled, "SingleModel", "single_model")
	}

	return enabled
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

	if r.isSingleSubscribeModeEnabled(ctx) {
		count, err := r.data.db.ProxyUserSubscribe.Query().
			Where(proxyusersubscribe.UserIDEQ(userID)).
			Count(ctx)
		if err != nil {
			r.logger.Errorf("Failed to count user subscribes: %v", err)
			return 0, err
		}
		if count >= 1 {
			return 0, responsecode.NewKratosError(responsecode.ErrSingleSubscribeModeExceedsLimit)
		}
	}

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

	trafficVal := req.Traffic
	if trafficVal == 0 {
		trafficVal = subscribePlan.Traffic
	}

	// 生成Token和UUID
	tokenStr := uuidx.SubscribeToken(fmt.Sprintf("adminCreate:%d", time.Now().UnixMilli()))
	subscribeUUID := uuid.New().String()

	startTime := time.Now()
	var expireTime *time.Time
	if req.ExpiredAt > 0 {
		t := time.UnixMilli(req.ExpiredAt)
		expireTime = &t
	}

	// 创建用户订阅
	create := r.data.db.ProxyUserSubscribe.Create().
		SetUserID(userID).
		SetOrderID(0).
		SetSubscribeID(subscribeID).
		SetStartTime(startTime).
		SetTraffic(trafficVal).
		SetDownload(0).
		SetUpload(0).
		SetNodeGroupID(getInt64ValueFromPointer(subscribePlan.NodeGroupID)).
		SetGroupLocked(false).
		SetToken(tokenStr).
		SetUUID(subscribeUUID).
		SetStatus(1)
	if expireTime != nil {
		create = create.SetExpireTime(*expireTime)
	}
	created, err := create.Save(ctx)

	if err != nil {
		r.logger.Errorf("Failed to create user subscribe: %v", err)
		return 0, err
	}

	return created.ID, nil
}

// UpdateUserSubscribe 更新用户订阅
func (r *adminUserSubscribeRepo) UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) error {
	id, err := strconv.ParseInt(req.UserSubscribeId, 10, 64)
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

	status := int8(1)
	var expireTime *time.Time
	if req.ExpiredAt > 0 {
		t := time.UnixMilli(req.ExpiredAt)
		expireTime = &t
		if time.Since(t).Minutes() > 0 {
			status = 3
		}
	}

	update := userSub.Update().
		SetTraffic(req.Traffic).
		SetDownload(req.Download).
		SetUpload(req.Upload).
		SetStatus(status).
		SetUpdatedAt(time.Now())
	if expireTime != nil {
		update.SetExpireTime(*expireTime)
	} else {
		update.ClearExpireTime()
	}
	if req.SubscribeId != "" {
		subscribeID, err := strconv.ParseInt(req.SubscribeId, 10, 64)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		update.SetSubscribeID(subscribeID)
	}

	err = update.Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to update user subscribe: %v", err)
		return err
	}

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

	return nil
}

// GetUserSubscribeById 根据ID获取用户订阅详情（包含套餐信息）
func (r *adminUserSubscribeRepo) GetUserSubscribeById(ctx context.Context, id int64) (*v1.UserSubscribeDetail, error) {
	// 查询用户订阅
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.IDEQ(id),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query user subscribe: %v", err)
		return nil, err
	}

	userInfo, err := r.data.db.ProxyUser.Query().
		Where(proxyuser.IDEQ(userSub.UserID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotExist)
		}
		r.logger.Errorf("Failed to query user info: %v", err)
		return nil, err
	}

	subscribePlan, err := r.data.db.ProxySubscribe.Query().
		Where(proxysubscribe.IDEQ(userSub.SubscribeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		r.logger.Errorf("Failed to query subscribe plan: %v", err)
		return nil, err
	}

	authMethods, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserIDEQ(userSub.UserID)).
		All(ctx)
	if err != nil {
		r.logger.Errorf("Failed to query user auth methods: %v", err)
		return nil, err
	}

	email := ""
	telephone := ""
	telephoneAreaCode := ""
	for _, method := range authMethods {
		switch strings.ToLower(strings.TrimSpace(method.AuthType)) {
		case "email":
			if email == "" {
				email = method.AuthIdentifier
			}
		case "mobile":
			if telephone == "" {
				parts := strings.SplitN(method.AuthIdentifier, "-", 2)
				if len(parts) == 2 {
					telephoneAreaCode = parts[0]
					telephone = parts[1]
				} else {
					telephone = method.AuthIdentifier
				}
			}
		}
	}

	detail := &v1.UserSubscribeDetail{
		Id:          strconv.FormatInt(int64(userSub.ID), 10),
		UserId:      strconv.FormatInt(int64(userSub.UserID), 10),
		OrderId:     strconv.FormatInt(int64(userSub.OrderID), 10),
		SubscribeId: strconv.FormatInt(int64(userSub.SubscribeID), 10),
		NodeGroupId: strconv.FormatInt(int64(userSub.NodeGroupID), 10),
		GroupLocked: userSub.GroupLocked,
		StartTime:   userSub.StartTime.UnixMilli(),
		Traffic:     getInt64ValueFromPointer(userSub.Traffic),
		Download:    getInt64ValueFromPointer(userSub.Download),
		Upload:      getInt64ValueFromPointer(userSub.Upload),
		Token:       getStringValue(userSub.Token),
		Status:      int32(getInt8Value(userSub.Status)),
		CreatedAt:   userSub.CreatedAt.UnixMilli(),
		UpdatedAt:   userSub.UpdatedAt.UnixMilli(),
		FinishedAt:  timePointerToUnixMilli(userSub.FinishedAt),
		Uuid:        getStringValue(userSub.UUID),
		User: &v1.User{
			Id:                    strconv.FormatInt(userInfo.ID, 10),
			Email:                 email,
			Telephone:             telephone,
			TelephoneAreaCode:     telephoneAreaCode,
			Avatar:                getStringValue(userInfo.Avatar),
			Balance:               getInt64ValueFromPointer(userInfo.Balance),
			ReferCode:             getStringValue(userInfo.ReferCode),
			RefererId:             getInt64ValueFromPointer(userInfo.RefererID),
			Commission:            getInt64ValueFromPointer(userInfo.Commission),
			ReferralPercentage:    int32(userInfo.ReferralPercentage),
			OnlyFirstPurchase:     userInfo.OnlyFirstPurchase,
			GiftAmount:            getInt64ValueFromPointer(userInfo.GiftAmount),
			Enable:                userInfo.Enable,
			IsAdmin:               userInfo.IsAdmin,
			EnableBalanceNotify:   userInfo.EnableBalanceNotify,
			EnableLoginNotify:     userInfo.EnableLoginNotify,
			EnableSubscribeNotify: userInfo.EnableSubscribeNotify,
			EnableTradeNotify:     userInfo.EnableTradeNotify,
			CreatedAt:             userInfo.CreatedAt.UnixMilli(),
			UpdatedAt:             userInfo.UpdatedAt.UnixMilli(),
			Telegram:              getInt64ValueFromPointer(userInfo.Telegram),
			DeletedAt:             timePointerToUnixMilli(userInfo.DeletedAt),
			IsDel:                 userInfo.IsDel != nil && *userInfo.IsDel == 1,
		},
		Subscribe: &v1.Subscribe{
			Id:          strconv.FormatInt(subscribePlan.ID, 10),
			Name:        subscribePlan.Name,
			Description: getStringValue(subscribePlan.Description),
			UnitPrice:   subscribePlan.UnitPrice,
			UnitTime:    subscribePlan.UnitTime,
			Traffic:     subscribePlan.Traffic,
			SpeedLimit:  subscribePlan.SpeedLimit,
			DeviceLimit: subscribePlan.DeviceLimit,
			ResetCycle:  getInt64ValueFromPointer(subscribePlan.ResetCycle),
			CreatedAt:   subscribePlan.CreatedAt.UnixMilli(),
			UpdatedAt:   subscribePlan.UpdatedAt.UnixMilli(),
		},
	}
	if userSub.ExpireTime != nil {
		detail.ExpireTime = userSub.ExpireTime.UnixMilli()
	}
	return detail, nil
}

// GetUserSubscribeDevices 获取用户订阅设备列表
func (r *adminUserSubscribeRepo) GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) ([]*ent.ProxyUserDevice, int64, error) {
	query := r.data.db.ProxyUserDevice.Query()
	if req.UserId != "" {
		userID, err := strconv.ParseInt(req.UserId, 10, 64)
		if err != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		query = query.Where(proxyuserdevice.UserIDEQ(userID))
	}
	if req.SubscribeId != "" {
		subscribeID, err := strconv.ParseInt(req.SubscribeId, 10, 64)
		if err != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		query = query.Where(proxyuserdevice.SubscribeIDEQ(subscribeID))
	}

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Errorf("Failed to count user subscribe devices: %v", err)
		return nil, 0, err
	}

	list, err := query.
		Order(ent.Desc(proxyuserdevice.FieldCreatedAt)).
		Offset(int((req.Page - 1) * req.Size)).
		Limit(int(req.Size)).
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

	userSubID := int64(0)
	if req.UserSubscribeId != "" {
		userSubID, _ = strconv.ParseInt(req.UserSubscribeId, 10, 64)
	}
	if userSubID == 0 && req.SubscribeId != "" {
		userSubID, _ = strconv.ParseInt(req.SubscribeId, 10, 64)
	}
	if userSubID > 0 {
		query = query.Where(proxysystemlog.ObjectIDEQ(userSubID))
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

	if req.UserSubscribeId != "" {
		userSubID, _ := strconv.ParseInt(req.UserSubscribeId, 10, 64)
		query = query.Where(proxysystemlog.ObjectIDEQ(userSubID))
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

	if req.SubscribeId != "" {
		subscribeID, _ := strconv.ParseInt(req.SubscribeId, 10, 64)
		query = query.Where(proxytrafficlog.SubscribeIDEQ(subscribeID))
	}
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
	if req.StartTime > 0 {
		query = query.Where(proxytrafficlog.TimestampGTE(time.UnixMilli(req.StartTime)))
	}
	if req.EndTime > 0 {
		query = query.Where(proxytrafficlog.TimestampLTE(time.UnixMilli(req.EndTime)))
	}

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

// ResetUserSubscribeToken 重置用户订阅令牌
func (r *adminUserSubscribeRepo) ResetUserSubscribeToken(ctx context.Context, userSubscribeID int64) error {
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDEQ(userSubscribeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	newToken := uuidx.SubscribeToken(fmt.Sprintf("AdminUpdate:%d", time.Now().UnixMilli()))
	if err = r.data.db.ProxyUserSubscribe.UpdateOneID(userSub.ID).
		SetToken(newToken).
		SetUpdatedAt(time.Now()).
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	return nil
}

// ToggleUserSubscribeStatus 切换用户订阅状态
func (r *adminUserSubscribeRepo) ToggleUserSubscribeStatus(ctx context.Context, userSubscribeID int64) error {
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDEQ(userSubscribeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	currentStatus := int8(0)
	if userSub.Status != nil {
		currentStatus = *userSub.Status
	}

	var nextStatus int8
	switch currentStatus {
	case 2:
		nextStatus = 5
	case 5:
		nextStatus = 2
	default:
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if err = r.data.db.ProxyUserSubscribe.UpdateOneID(userSub.ID).
		SetStatus(nextStatus).
		SetUpdatedAt(time.Now()).
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	return nil
}

// ResetUserSubscribeTraffic 重置用户订阅流量
func (r *adminUserSubscribeRepo) ResetUserSubscribeTraffic(ctx context.Context, userSubscribeID int64) error {
	_, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDEQ(userSubscribeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	if err = r.data.db.ProxyUserSubscribe.UpdateOneID(userSubscribeID).
		SetDownload(0).
		SetUpload(0).
		SetUpdatedAt(time.Now()).
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	return nil
}

// 辅助函数：获取int64指针的值，nil返回0
func getInt64ValueFromPointer(p *int64) int64 {
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

func timePointerToUnixMilli(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.UnixMilli()
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
