package data

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystemlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
)

type adminUserRepo struct {
	data   *Data
	logger *log.Helper
}

// NewAdminUserRepo 创建用户数据仓库
func NewAdminUserRepo(data *Data, logger log.Logger) userbiz.UserRepo {
	return &adminUserRepo{
		data:   data,
		logger: log.NewHelper(logger),
	}
}

// CreateUser 创建用户
func (r *adminUserRepo) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (int64, error) {
	// 生成推荐码（如果未提供）
	referCode := req.ReferCode
	if referCode == "" {
		referCode = uuidx.UserInviteCode(time.Now().UnixMicro())
	}

	// 密码处理（如果未提供，默认使用邮箱）
	password := req.Password
	if password == "" {
		password = req.Email
	}
	encodedPwd := tool.EncodePassWord(password)

	// 查找推荐人ID
	var refererID *int64
	if req.RefererUser != "" {
		// 通过推荐人邮箱查找认证方法
		authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.AuthTypeEQ("email"),
				proxyuserauthmethod.AuthIdentifierEQ(req.RefererUser),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return 0, responsecode.NewKratosError(responsecode.ErrUserNotExist)
			}
			return 0, err
		}
		refererID = new(int64)
	*refererID = int64(authMethod.UserID)
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		exists, err := r.data.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.AuthTypeEQ("email"),
				proxyuserauthmethod.AuthIdentifierEQ(req.Email),
			).
			Exist(ctx)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, responsecode.NewKratosError(responsecode.ErrEmailExist)
		}
	}

	// 检查手机号是否已存在
	if req.TelephoneAreaCode != "" && req.Telephone != "" {
		phone := fmt.Sprintf("%s-%s", req.TelephoneAreaCode, req.Telephone)
		exists, err := r.data.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.AuthTypeEQ("mobile"),
				proxyuserauthmethod.AuthIdentifierEQ(phone),
			).
			Exist(ctx)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, responsecode.NewKratosError(responsecode.ErrTelephoneExist)
		}
	}

	// 使用事务创建用户和认证方法
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return 0, err
	}

	// 创建用户
	builder := tx.ProxyUser.Create().
		SetPassword(encodedPwd).
		SetReferCode(referCode).
		SetBalance(req.Balance).
		SetCommission(req.Commission).
		SetGiftAmount(req.GiftAmount).
		SetReferralPercentage(int(req.ReferralPercentage)).
		SetOnlyFirstPurchase(req.OnlyFirstPurchase).
		SetIsAdmin(req.IsAdmin)

	if refererID != nil {
		builder.SetRefererID(int(*refererID))
	}

	user, err := builder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	// 创建邮箱认证方法
	if req.Email != "" {
		err = tx.ProxyUserAuthMethod.Create().
			SetUserID(user.ID).
			SetAuthType("email").
			SetAuthIdentifier(req.Email).
			SetVerified(false).
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	// 创建手机号认证方法
	if req.TelephoneAreaCode != "" && req.Telephone != "" {
		phone := fmt.Sprintf("%s-%s", req.TelephoneAreaCode, req.Telephone)
		err = tx.ProxyUserAuthMethod.Create().
			SetUserID(user.ID).
			SetAuthType("mobile").
			SetAuthIdentifier(phone).
			SetVerified(false).
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int64(user.ID), nil
}

// DeleteUser 删除用户
func (r *adminUserRepo) DeleteUser(ctx context.Context, userID int64) error {
	// Demo模式保护
	isDemo := strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo"
	if userID == 2 && isDemo {
		return responsecode.NewKratosError(503) // Demo mode restriction
	}

	// 验证用户存在
	deleted, err := r.data.db.ProxyUser.Delete().
		Where(
			proxyuser.IDEQ(int(userID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}

	if deleted == 0 {
		return responsecode.NewKratosError(responsecode.ErrUserNotExist)
	}

	return nil
}

// BatchDeleteUser 批量删除用户
func (r *adminUserRepo) BatchDeleteUser(ctx context.Context, userIDs []int64) (int64, error) {
	// Demo模式保护
	isDemo := strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo"
	if isDemo {
		for _, id := range userIDs {
			if id == 2 {
				return 0, responsecode.NewKratosError(503) // Demo mode restriction
			}
		}
	}

	intUserIDs := make([]int, len(userIDs))
	for i, id := range userIDs {
		intUserIDs[i] = int(id)
	}

	deleted, err := r.data.db.ProxyUser.Delete().
		Where(
			proxyuser.IDIn(intUserIDs...),
		).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int64(deleted), nil
}

// GetUserByID 根据ID获取用户
func (r *adminUserRepo) GetUserByID(ctx context.Context, userID int64) (*ent.ProxyUser, error) {
	user, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.IDEQ(int(userID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotExist)
		}
		return nil, err
	}

	return user, nil
}

// GetUserList 获取用户列表
func (r *adminUserRepo) GetUserList(ctx context.Context, page, size int32, search string, userID, subscribeID, userSubscribeID *int64) ([]*ent.ProxyUser, int64, error) {
	query := r.data.db.ProxyUser.Query()

	// 按用户ID过滤
	if userID != nil && *userID > 0 {
		query = query.Where(proxyuser.IDEQ(int(*userID)))
	}

	// 按搜索关键字过滤（邮箱或手机号）
	if search != "" {
		// 先查找匹配的认证方法
		authMethods, err := r.data.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.AuthIdentifierContains(search),
			).
			All(ctx)
		if err != nil {
			return nil, 0, err
		}

		if len(authMethods) > 0 {
			intUserIDs := make([]int, 0, len(authMethods))
			for _, am := range authMethods {
				intUserIDs = append(intUserIDs, int(am.UserID))
			}
			query = query.Where(proxyuser.IDIn(intUserIDs...))
		} else {
			// 如果没有匹配的认证方法，返回空结果
			return []*ent.ProxyUser{}, 0, nil
		}
	}

	// TODO: 按订阅套餐ID和用户订阅ID过滤
	// 这需要关联proxy_user_subscribe表，暂时跳过

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	users, err := query.
		Order(func(s *sql.Selector) {
			s.OrderBy(sql.Desc(proxyuser.FieldCreatedAt))
		}).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, int64(total), nil
}

// UpdateUserBasicInfo 更新用户基本信息
func (r *adminUserRepo) UpdateUserBasicInfo(ctx context.Context, req *v1.UpdateUserBasicInfoRequest) error {
	// 先查询用户当前信息
	userInfo, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.IDEQ(int(req.UserId)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrUserNotExist)
		}
		return err
	}

	isDemo := strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo"

	// 头像大小验证（如果提供）
	if req.Avatar != "" && !tool.IsValidImageSize(req.Avatar, 1024) {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 余额变更日志
	if req.Balance > 0 {
		currentBalance := int64(0)
		if userInfo.Balance != nil {
			currentBalance = *userInfo.Balance
		}
		if currentBalance != req.Balance {
			change := req.Balance - currentBalance
			balanceLog := logmodel.Balance{
				Type:      logmodel.BalanceTypeAdjust,
				Amount:    change,
				OrderNo:   "",
				Balance:   req.Balance,
				Timestamp: time.Now().UnixMilli(),
			}
			content, _ := balanceLog.Marshal()

			err = r.data.db.ProxySystemLog.Create().
				SetType(int8(logmodel.TypeBalance)).
				SetDate(time.Now().Format(time.DateOnly)).
				SetObjectID(req.UserId).
				SetContent(string(content)).
				Exec(ctx)
			if err != nil {
				r.logger.Errorf("Failed to insert balance log: %v", err)
			}
		}
	}

	// 赠送金额变更日志
	if req.GiftAmount > 0 {
		currentGiftAmount := int64(0)
		if userInfo.GiftAmount != nil {
			currentGiftAmount = *userInfo.GiftAmount
		}
		if currentGiftAmount != req.GiftAmount {
			change := req.GiftAmount - currentGiftAmount
			if change != 0 {
				var changeType uint16
				if currentGiftAmount < req.GiftAmount {
					changeType = logmodel.GiftTypeIncrease
				} else {
					changeType = logmodel.GiftTypeReduce
				}
				giftLog := logmodel.Gift{
					Type:      changeType,
					Amount:    change,
					Balance:   req.GiftAmount,
					Remark:    "Admin adjustment",
					Timestamp: time.Now().UnixMilli(),
				}
				content, _ := giftLog.Marshal()

				err = r.data.db.ProxySystemLog.Create().
					SetType(int8(logmodel.TypeGift)).
					SetDate(time.Now().Format(time.DateOnly)).
					SetObjectID(req.UserId).
					SetContent(string(content)).
					Exec(ctx)
				if err != nil {
					r.logger.Errorf("Failed to insert gift log: %v", err)
				}
			}
		}
	}

	// 佣金变更日志
	if req.Commission > 0 {
		currentCommission := int64(0)
		if userInfo.Commission != nil {
			currentCommission = *userInfo.Commission
		}
		if req.Commission != currentCommission {
			commissionLog := logmodel.Commission{
				Type:      logmodel.CommissionTypeAdjust,
				Amount:    req.Commission - currentCommission,
				Timestamp: time.Now().UnixMilli(),
			}
			content, _ := commissionLog.Marshal()

			err = r.data.db.ProxySystemLog.Create().
				SetType(int8(logmodel.TypeCommission)).
				SetDate(time.Now().Format(time.DateOnly)).
				SetObjectID(req.UserId).
				SetContent(string(content)).
				Exec(ctx)
			if err != nil {
				r.logger.Errorf("Failed to insert commission log: %v", err)
			}
		}
	}

	// 构建更新
	builder := r.data.db.ProxyUser.UpdateOneID(int(req.UserId)).
				SetUpdatedAt(time.Now())

	// Demo模式密码保护
	if req.Password != "" {
		if userInfo.ID == 2 && isDemo {
			return responsecode.NewKratosError(503) // Demo mode restriction
		}
		encodedPwd := tool.EncodePassWord(req.Password)
		builder.SetPassword(encodedPwd)
	}

	// 无条件更新所有数值字段（允许设置为0）
	builder.SetBalance(req.Balance)
	builder.SetCommission(req.Commission)
	builder.SetGiftAmount(req.GiftAmount)
	builder.SetRefererID(int(req.RefererId))
	builder.SetReferralPercentage(int(req.ReferralPercentage))
	builder.SetOnlyFirstPurchase(req.OnlyFirstPurchase)
	builder.SetEnable(req.Enable)
	builder.SetIsAdmin(req.IsAdmin)
	if req.Telegram > 0 {
			builder.SetTelegram(req.Telegram)
		}

	// 字符串字段：仅在非空时更新（避免清空重要字段）
	if req.Avatar != "" {
		builder.SetAvatar(req.Avatar)
	}
	if req.ReferCode != "" {
		builder.SetReferCode(req.ReferCode)
	}

	// 执行更新
	err = builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrUserNotExist)
		}
		return err
	}

	return nil
}

// UpdateUserNotifySettings 更新用户通知设置
func (r *adminUserRepo) UpdateUserNotifySettings(ctx context.Context, req *v1.UpdateUserNotifySettingsRequest) error {
	err := r.data.db.ProxyUser.UpdateOneID(int(req.UserId)).
				SetEnableBalanceNotify(req.EnableBalanceNotify).
		SetEnableLoginNotify(req.EnableLoginNotify).
		SetEnableSubscribeNotify(req.EnableSubscribeNotify).
		SetEnableTradeNotify(req.EnableTradeNotify).
		SetUpdatedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrUserNotExist)
		}
		return err
	}

	return nil
}

// GetUserLoginLogs 获取用户登录日志
func (r *adminUserRepo) GetUserLoginLogs(ctx context.Context, page, size int32, userID *int64, date string) ([]*ent.ProxySystemLog, int64, error) {
	query := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(logmodel.TypeLogin)), // 登录日志类型: 30
		)

	// 按用户ID过滤
	if userID != nil && *userID > 0 {
		query = query.Where(proxysystemlog.ObjectIDEQ(*userID))
	}

	// 按日期过滤
	if date != "" {
		query = query.Where(proxysystemlog.DateEQ(date))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	logs, err := query.
		Order(func(s *sql.Selector) {
			s.OrderBy(sql.Desc(proxysystemlog.FieldCreatedAt))
		}).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return logs, int64(total), nil
}

// parseDateRange 解析日期范围（YYYY-MM-DD）
func parseDateRange(date string) (time.Time, time.Time, error) {
	startTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime := startTime.Add(24 * time.Hour).Add(-time.Second)
	return startTime, endTime, nil
}
