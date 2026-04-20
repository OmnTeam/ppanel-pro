package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystemlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/pkg/phone"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatLegacyGetUserSubscribeListRequest struct {
	Page   int    `json:"page" form:"page"`
	Size   int    `json:"size" form:"size"`
	UserID string `json:"user_id,omitempty" form:"user_id,omitempty"`
}

type compatLegacyCreateUserSubscribeRequest struct {
	UserID      string              `json:"user_id" form:"user_id"`
	ExpiredAt   compatFlexibleInt64 `json:"expired_at" form:"expired_at"`
	Traffic     compatFlexibleInt64 `json:"traffic" form:"traffic"`
	SubscribeID string              `json:"subscribe_id" form:"subscribe_id"`
}

type compatLegacyUpdateUserSubscribeRequest struct {
	UserSubscribeID string              `json:"user_subscribe_id" form:"user_subscribe_id"`
	SubscribeID     string              `json:"subscribe_id" form:"subscribe_id"`
	Traffic         compatFlexibleInt64 `json:"traffic" form:"traffic"`
	ExpiredAt       compatFlexibleInt64 `json:"expired_at" form:"expired_at"`
	Upload          compatFlexibleInt64 `json:"upload" form:"upload"`
	Download        compatFlexibleInt64 `json:"download" form:"download"`
}

type compatLegacyDeleteUserSubscribeRequest struct {
	UserSubscribeID string `json:"user_subscribe_id" form:"user_subscribe_id"`
}

type compatLegacyGetUserSubscribeByIDRequest struct {
	ID string `json:"id,omitempty" form:"id"`
}

type compatLegacyGetUserSubscribeDevicesRequest struct {
	Page            int    `json:"page" form:"page"`
	Size            int    `json:"size" form:"size"`
	UserID          string `json:"user_id,omitempty" form:"user_id,omitempty"`
	SubscribeID     string `json:"subscribe_id,omitempty" form:"subscribe_id,omitempty"`
	UserSubscribeID string `json:"user_subscribe_id,omitempty" form:"user_subscribe_id,omitempty"`
}

type compatLegacyGetUserSubscribeLogsRequest struct {
	Page            int    `json:"page" form:"page"`
	Size            int    `json:"size" form:"size"`
	UserID          string `json:"user_id,omitempty" form:"user_id,omitempty"`
	SubscribeID     string `json:"subscribe_id,omitempty" form:"subscribe_id,omitempty"`
	UserSubscribeID string `json:"user_subscribe_id,omitempty" form:"user_subscribe_id,omitempty"`
}

type compatLegacyGetUserSubscribeResetTrafficLogsRequest struct {
	Page            int    `json:"page" form:"page"`
	Size            int    `json:"size" form:"size"`
	UserSubscribeID string `json:"user_subscribe_id,omitempty" form:"user_subscribe_id,omitempty"`
}

type compatLegacyGetUserSubscribeTrafficLogsRequest struct {
	Page            int    `json:"page" form:"page"`
	Size            int    `json:"size" form:"size"`
	UserID          string `json:"user_id,omitempty" form:"user_id,omitempty"`
	SubscribeID     string `json:"subscribe_id,omitempty" form:"subscribe_id,omitempty"`
	UserSubscribeID string `json:"user_subscribe_id,omitempty" form:"user_subscribe_id,omitempty"`
	StartTime       int64  `json:"start_time,omitempty" form:"start_time,omitempty"`
	EndTime         int64  `json:"end_time,omitempty" form:"end_time,omitempty"`
	Date            string `json:"date,omitempty" form:"date,omitempty"`
}

type compatLegacyResetUserSubscribeTokenRequest struct {
	UserSubscribeID string `json:"user_subscribe_id"`
}

type compatLegacyToggleUserSubscribeStatusRequest struct {
	UserSubscribeID string `json:"user_subscribe_id"`
}

type compatLegacyResetUserSubscribeTrafficRequest struct {
	UserSubscribeID string `json:"user_subscribe_id"`
}

type compatLegacyAdminUser struct {
	ID                    int64                  `json:"id,string"`
	Avatar                string                 `json:"avatar"`
	Balance               int64                  `json:"balance"`
	Commission            int64                  `json:"commission"`
	ReferralPercentage    uint8                  `json:"referral_percentage"`
	OnlyFirstPurchase     bool                   `json:"only_first_purchase"`
	GiftAmount            int64                  `json:"gift_amount"`
	Telegram              int64                  `json:"telegram"`
	ReferCode             string                 `json:"refer_code"`
	RefererID             int64                  `json:"referer_id,string"`
	Enable                bool                   `json:"enable"`
	IsAdmin               bool                   `json:"is_admin,omitempty"`
	EnableBalanceNotify   bool                   `json:"enable_balance_notify"`
	EnableLoginNotify     bool                   `json:"enable_login_notify"`
	EnableSubscribeNotify bool                   `json:"enable_subscribe_notify"`
	EnableTradeNotify     bool                   `json:"enable_trade_notify"`
	AuthMethods           []compatUserAuthMethod `json:"auth_methods"`
	UserDevices           []compatUserDevice     `json:"user_devices"`
	Rules                 []string               `json:"rules"`
	CreatedAt             int64                  `json:"created_at"`
	UpdatedAt             int64                  `json:"updated_at"`
	DeletedAt             int64                  `json:"deleted_at,omitempty"`
	IsDel                 bool                   `json:"is_del,omitempty"`
}

type compatLegacyAdminUserSubscribeDetail struct {
	ID          int64                 `json:"id,string"`
	UserID      int64                 `json:"user_id,string"`
	User        compatLegacyAdminUser `json:"user"`
	OrderID     int64                 `json:"order_id,string"`
	SubscribeID int64                 `json:"subscribe_id,string"`
	Subscribe   compatSubscribe       `json:"subscribe"`
	NodeGroupID int64                 `json:"node_group_id,string"`
	GroupLocked bool                  `json:"group_locked"`
	StartTime   int64                 `json:"start_time"`
	ExpireTime  int64                 `json:"expire_time"`
	ResetTime   int64                 `json:"reset_time"`
	Traffic     int64                 `json:"traffic"`
	Download    int64                 `json:"download"`
	Upload      int64                 `json:"upload"`
	Token       string                `json:"token"`
	Status      uint8                 `json:"status"`
	CreatedAt   int64                 `json:"created_at"`
	UpdatedAt   int64                 `json:"updated_at"`
}

type compatLegacyResetSubscribeTrafficLog struct {
	ID              int64  `json:"id,string"`
	Type            uint16 `json:"type"`
	UserSubscribeID int64  `json:"user_subscribe_id,string"`
	OrderNo         string `json:"order_no,omitempty"`
	Timestamp       int64  `json:"timestamp"`
}

type compatLegacyResetSubscribeTrafficLogListData struct {
	Total int64                                  `json:"total"`
	List  []compatLegacyResetSubscribeTrafficLog `json:"list"`
}

type compatLegacyTrafficLog struct {
	ID          int64 `json:"id,string"`
	ServerID    int64 `json:"server_id,string"`
	UserID      int64 `json:"user_id,string"`
	SubscribeID int64 `json:"subscribe_id,string"`
	Download    int64 `json:"download"`
	Upload      int64 `json:"upload"`
	Timestamp   int64 `json:"timestamp"`
}

type compatLegacyTrafficLogListData struct {
	Total int64                    `json:"total"`
	List  []compatLegacyTrafficLog `json:"list"`
}

func registerLegacyAdminUserSubscribeCompatRoutes(r *khttp.Router, dataLayer *data.Data) {
	if r == nil || dataLayer == nil || dataLayer.DB() == nil {
		return
	}

	r.GET("/v1/admin/user/subscribe", compatGetUserSubscribeHandler(dataLayer))
	r.POST("/v1/admin/user/subscribe", compatCreateUserSubscribeHandler(dataLayer))
	r.PUT("/v1/admin/user/subscribe", compatUpdateUserSubscribeHandler(dataLayer))
	r.DELETE("/v1/admin/user/subscribe", compatDeleteUserSubscribeHandler(dataLayer))
	r.GET("/v1/admin/user/subscribe/detail", compatGetUserSubscribeDetailHandler(dataLayer))
	r.GET("/v1/admin/user/subscribe/device", compatGetUserSubscribeDevicesHandler(dataLayer))
	r.GET("/v1/admin/user/subscribe/logs", compatGetUserSubscribeLogsHandler(dataLayer))
	r.GET("/v1/admin/user/subscribe/reset/logs", compatGetUserSubscribeResetTrafficLogsHandler(dataLayer))
	r.GET("/v1/admin/user/subscribe/traffic_logs", compatGetUserSubscribeTrafficLogsHandler(dataLayer))
	r.POST("/v1/admin/user/subscribe/reset/token", compatResetUserSubscribeTokenHandler(dataLayer))
	r.POST("/v1/admin/user/subscribe/toggle", compatToggleUserSubscribeStatusHandler(dataLayer))
	r.POST("/v1/admin/user/subscribe/reset/traffic", compatResetUserSubscribeTrafficHandler(dataLayer))
}

func compatGetUserSubscribeHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyGetUserSubscribeListRequest
		_ = ctx.BindQuery(&req)
		compatFillLegacyUserSubscribeListQuery(ctx, &req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatLegacyGetUserSubscribeListRequest)
			return compatLegacyAdminUserSubscribeList(inner, dataLayer, compatParseInt64String(in.UserID))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func compatCreateUserSubscribeHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyCreateUserSubscribeRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatCreateLegacyUserSubscribe(inner, dataLayer, request.(*compatLegacyCreateUserSubscribeRequest))
		}); err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	}
}

func compatUpdateUserSubscribeHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyUpdateUserSubscribeRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatUpdateLegacyUserSubscribe(inner, dataLayer, request.(*compatLegacyUpdateUserSubscribeRequest))
		}); err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	}
}

func compatDeleteUserSubscribeHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyDeleteUserSubscribeRequest
		_ = ctx.BindQuery(&req)
		compatFillLegacyDeleteUserSubscribeQuery(ctx, &req)

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatDeleteLegacyUserSubscribe(inner, dataLayer, compatParseInt64String(request.(*compatLegacyDeleteUserSubscribeRequest).UserSubscribeID))
		}); err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	}
}

func compatGetUserSubscribeDetailHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyGetUserSubscribeByIDRequest
		_ = ctx.BindQuery(&req)
		if strings.TrimSpace(req.ID) == "" && ctx.Request() != nil && ctx.Request().URL != nil {
			req.ID = strings.TrimSpace(ctx.Request().URL.Query().Get("id"))
		}
		if strings.TrimSpace(req.ID) == "" {
			return compatJSONError(ctx, compatRequiredFieldError("GetUserSubscribeByIdRequest", "Id"))
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLoadLegacyAdminUserSubscribeDetail(inner, dataLayer, compatParseInt64String(request.(*compatLegacyGetUserSubscribeByIDRequest).ID))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func compatGetUserSubscribeDevicesHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyGetUserSubscribeDevicesRequest
		_ = ctx.BindQuery(&req)
		compatFillLegacyUserSubscribeDevicesQuery(ctx, &req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyAdminUserSubscribeDevices(inner, dataLayer, request.(*compatLegacyGetUserSubscribeDevicesRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func compatGetUserSubscribeLogsHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyGetUserSubscribeLogsRequest
		_ = ctx.BindQuery(&req)
		compatFillLegacyUserSubscribeLogsQuery(ctx, &req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyAdminUserSubscribeLogs(inner, dataLayer, request.(*compatLegacyGetUserSubscribeLogsRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func compatGetUserSubscribeResetTrafficLogsHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyGetUserSubscribeResetTrafficLogsRequest
		_ = ctx.BindQuery(&req)
		compatFillLegacyUserSubscribeResetTrafficLogsQuery(ctx, &req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyAdminUserSubscribeResetTrafficLogs(inner, dataLayer, request.(*compatLegacyGetUserSubscribeResetTrafficLogsRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func compatGetUserSubscribeTrafficLogsHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyGetUserSubscribeTrafficLogsRequest
		_ = ctx.BindQuery(&req)
		compatFillLegacyUserSubscribeTrafficLogsQuery(ctx, &req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyAdminUserSubscribeTrafficLogs(inner, dataLayer, request.(*compatLegacyGetUserSubscribeTrafficLogsRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func compatResetUserSubscribeTokenHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyResetUserSubscribeTokenRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatResetLegacyUserSubscribeToken(inner, dataLayer, compatParseInt64String(request.(*compatLegacyResetUserSubscribeTokenRequest).UserSubscribeID))
		}); err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	}
}

func compatToggleUserSubscribeStatusHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyToggleUserSubscribeStatusRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatToggleLegacyUserSubscribeStatus(inner, dataLayer, compatParseInt64String(request.(*compatLegacyToggleUserSubscribeStatusRequest).UserSubscribeID))
		}); err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	}
}

func compatResetUserSubscribeTrafficHandler(dataLayer *data.Data) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatLegacyResetUserSubscribeTrafficRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatResetLegacyUserSubscribeTraffic(inner, dataLayer, compatParseInt64String(request.(*compatLegacyResetUserSubscribeTrafficRequest).UserSubscribeID))
		}); err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	}
}

func compatLegacyAdminUserSubscribeList(ctx context.Context, dataLayer *data.Data, userID int64) (*compatUserSubscribeListData, error) {
	query := dataLayer.DB().ProxyUserSubscribe.Query()
	if userID > 0 {
		query = query.Where(proxyusersubscribe.UserIDEQ(userID))
	}

	items, err := query.
		Where(proxyusersubscribe.StatusIn(0, 1, 2, 3, 4, 5)).
		Order(ent.Desc(proxyusersubscribe.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	subscribeIDs := make([]int64, 0, len(items))
	for _, item := range items {
		subscribeIDs = append(subscribeIDs, item.SubscribeID)
	}
	subscribeIDs = compatUniqueInt64(subscribeIDs)

	subscribeMap := map[int64]*ent.ProxySubscribe{}
	if len(subscribeIDs) > 0 {
		subscribes, err := dataLayer.DB().ProxySubscribe.Query().
			Where(proxysubscribe.IDIn(subscribeIDs...)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range subscribes {
			subscribeMap[item.ID] = item
		}
	}

	result := &compatUserSubscribeListData{List: []compatUserSubscribe{}}
	for _, item := range items {
		subscribePlan := subscribeMap[item.SubscribeID]
		token := compatStringValue(item.Token)
		short, _ := tool.FixedUniqueString(token, 8, "")
		entry := compatUserSubscribe{
			ID:          item.ID,
			IDStr:       strconv.FormatInt(item.ID, 10),
			UserID:      item.UserID,
			OrderID:     item.OrderID,
			SubscribeID: item.SubscribeID,
			Subscribe:   compatLegacySubscribeFromEntity(subscribePlan),
			NodeGroupID: item.NodeGroupID,
			GroupLocked: item.GroupLocked,
			StartTime:   item.StartTime.UnixMilli(),
			ExpireTime:  compatLegacyUnixMillis(item.ExpireTime),
			FinishedAt:  compatLegacyUnixMillis(item.FinishedAt),
			Traffic:     compatInt64Value(item.Traffic),
			Download:    compatInt64Value(item.Download),
			Upload:      compatInt64Value(item.Upload),
			Token:       token,
			Status:      uint8(compatInt8Value(item.Status)),
			Short:       short,
			CreatedAt:   item.CreatedAt.UnixMilli(),
			UpdatedAt:   item.UpdatedAt.UnixMilli(),
		}
		entry.ResetTime = compatLegacyUserResetTime(entry)
		result.List = append(result.List, entry)
	}
	result.Total = int64(len(result.List))
	return result, nil
}

func compatCreateLegacyUserSubscribe(ctx context.Context, dataLayer *data.Data, req *compatLegacyCreateUserSubscribeRequest) error {
	if req == nil {
		return compatParamError("invalid request")
	}

	userID := compatParseInt64String(req.UserID)
	subscribeID := compatParseInt64String(req.SubscribeID)
	if userID <= 0 || subscribeID <= 0 {
		return compatCodeError(500, "FindOne error: invalid user or subscribe id")
	}

	userInfo, err := dataLayer.DB().ProxyUser.Get(ctx, userID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("FindOne error: %v", err))
	}

	singleModelEnabled := dataLayer.AppConf() != nil && dataLayer.AppConf().Subscribe != nil && dataLayer.AppConf().Subscribe.SingleModel
	if singleModelEnabled {
		count, err := dataLayer.DB().ProxyUserSubscribe.Query().
			Where(proxyusersubscribe.UserIDEQ(userID)).
			Count(ctx)
		if err != nil {
			return compatCodeError(500, fmt.Sprintf("QueryUserSubscribe error: %v", err))
		}
		if count >= 1 {
			return compatCodeError(10021, "Single subscribe mode exceeds limit")
		}
	}

	subscribePlan, err := dataLayer.DB().ProxySubscribe.Get(ctx, subscribeID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("FindOne error: %v", err))
	}

	traffic := int64(req.Traffic)
	if traffic == 0 {
		traffic = subscribePlan.Traffic
	}
	var expireTime *time.Time
	if int64(req.ExpiredAt) > 0 {
		t := time.UnixMilli(int64(req.ExpiredAt))
		expireTime = &t
	}
	token := uuidx.SubscribeToken(fmt.Sprintf("adminCreate:%d", time.Now().UnixMilli()))
	subscribeUUID := uuid.New().String()

	create := dataLayer.DB().ProxyUserSubscribe.Create().
		SetUserID(userID).
		SetOrderID(0).
		SetSubscribeID(subscribeID).
		SetStartTime(time.Now()).
		SetTraffic(traffic).
		SetDownload(0).
		SetUpload(0).
		SetNodeGroupID(compatInt64Value(subscribePlan.NodeGroupID)).
		SetGroupLocked(false).
		SetToken(token).
		SetUUID(subscribeUUID).
		SetStatus(1)
	if expireTime != nil {
		create = create.SetExpireTime(*expireTime)
	}
	_, err = create.Save(ctx)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("InsertSubscribe error: %v", err))
	}

	return compatClearLegacyUserCaches(ctx, dataLayer, userInfo.ID, subscribeID)
}

func compatUpdateLegacyUserSubscribe(ctx context.Context, dataLayer *data.Data, req *compatLegacyUpdateUserSubscribeRequest) error {
	if req == nil {
		return compatParamError("invalid request")
	}

	userSubscribeID := compatParseInt64String(req.UserSubscribeID)
	userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubscribeID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("FindOneUserSubscribe failed: %v", err))
	}

	status := int8(1)
	var expireTime *time.Time
	if int64(req.ExpiredAt) > 0 {
		t := time.UnixMilli(int64(req.ExpiredAt))
		expireTime = &t
		if time.Since(t).Minutes() > 0 {
			status = 3
		}
	}

	update := dataLayer.DB().ProxyUserSubscribe.UpdateOneID(userSub.ID).
		SetTraffic(int64(req.Traffic)).
		SetDownload(int64(req.Download)).
		SetUpload(int64(req.Upload)).
		SetStatus(status).
		SetUpdatedAt(time.Now())
	if expireTime != nil {
		update = update.SetExpireTime(*expireTime)
	} else {
		update = update.ClearExpireTime()
	}

	newSubscribeID := compatParseInt64String(req.SubscribeID)
	if newSubscribeID > 0 {
		update = update.SetSubscribeID(newSubscribeID)
	}

	if err := update.Exec(ctx); err != nil {
		return compatCodeError(500, fmt.Sprintf("UpdateSubscribe failed: %v", err))
	}

	return compatClearLegacyUserCaches(ctx, dataLayer, userSub.UserID, userSub.SubscribeID, newSubscribeID)
}

func compatDeleteLegacyUserSubscribe(ctx context.Context, dataLayer *data.Data, userSubscribeID int64) error {
	userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubscribeID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("failed to find user subscribe: %v", err))
	}
	if err := dataLayer.DB().ProxyUserSubscribe.DeleteOneID(userSubscribeID).Exec(ctx); err != nil {
		return compatCodeError(500, fmt.Sprintf("failed to delete user subscribe: %v", err))
	}
	return compatClearLegacyUserCaches(ctx, dataLayer, userSub.UserID, userSub.SubscribeID)
}

func compatLoadLegacyAdminUserSubscribeDetail(ctx context.Context, dataLayer *data.Data, userSubscribeID int64) (*compatLegacyAdminUserSubscribeDetail, error) {
	userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubscribeID)
	if err != nil {
		return nil, compatCodeError(500, fmt.Sprintf("FindOneSubscribeDetailsById error: %v", err))
	}
	userInfo, err := dataLayer.DB().ProxyUser.Get(ctx, userSub.UserID)
	if err != nil {
		return nil, compatCodeError(500, fmt.Sprintf("FindOneSubscribeDetailsById error: %v", err))
	}
	subscribePlan, err := dataLayer.DB().ProxySubscribe.Get(ctx, userSub.SubscribeID)
	if err != nil {
		return nil, compatCodeError(500, fmt.Sprintf("FindOneSubscribeDetailsById error: %v", err))
	}

	adminUser, err := compatLegacyAdminUserFromEntity(ctx, dataLayer, userInfo)
	if err != nil {
		return nil, err
	}

	detail := &compatLegacyAdminUserSubscribeDetail{
		ID:          userSub.ID,
		UserID:      userSub.UserID,
		User:        *adminUser,
		OrderID:     userSub.OrderID,
		SubscribeID: userSub.SubscribeID,
		Subscribe:   compatLegacySubscribeFromEntity(subscribePlan),
		NodeGroupID: userSub.NodeGroupID,
		GroupLocked: userSub.GroupLocked,
		StartTime:   userSub.StartTime.UnixMilli(),
		ExpireTime:  compatLegacyUnixMillis(userSub.ExpireTime),
		Traffic:     compatInt64Value(userSub.Traffic),
		Download:    compatInt64Value(userSub.Download),
		Upload:      compatInt64Value(userSub.Upload),
		Token:       compatStringValue(userSub.Token),
		Status:      uint8(compatInt8Value(userSub.Status)),
		CreatedAt:   userSub.CreatedAt.UnixMilli(),
		UpdatedAt:   userSub.UpdatedAt.UnixMilli(),
	}
	detail.ResetTime = compatLegacyUserResetTime(compatUserSubscribe{
		ExpireTime: detail.ExpireTime,
		Subscribe:  detail.Subscribe,
	})
	return detail, nil
}

func compatLegacyAdminUserSubscribeDevices(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetUserSubscribeDevicesRequest) (*compatUserDeviceListData, error) {
	userID, subscribeID, err := compatResolveLegacyUserSubscribeDeviceTarget(ctx, dataLayer, req)
	if err != nil {
		return nil, err
	}

	query := dataLayer.DB().ProxyUserDevice.Query()
	if userID > 0 {
		query = query.Where(proxyuserdevice.UserIDEQ(userID))
	}
	if subscribeID > 0 {
		query = query.Where(proxyuserdevice.SubscribeIDEQ(subscribeID))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	if req != nil && req.Page > 0 && req.Size > 0 {
		query = query.Offset((req.Page - 1) * req.Size).Limit(req.Size)
	}

	items, err := query.
		Order(ent.Desc(proxyuserdevice.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := &compatUserDeviceListData{Total: int64(total), List: []compatUserDevice{}}
	for _, item := range items {
		result.List = append(result.List, compatUserDevice{
			ID:         item.ID,
			IP:         compatStringValue(item.IP),
			Identifier: compatStringValue(item.Identifier),
			UserAgent:  compatStringValue(item.UserAgent),
			Online:     item.Online,
			Enabled:    item.Enabled,
			CreatedAt:  item.CreatedAt.UnixMilli(),
			UpdatedAt:  item.UpdatedAt.UnixMilli(),
		})
	}
	return result, nil
}

func compatLegacyAdminUserSubscribeLogs(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetUserSubscribeLogsRequest) (*compatUserSubscribeLogListData, error) {
	query := dataLayer.DB().ProxySystemLog.Query().
		Where(proxysystemlog.TypeEQ(int8(logmodel.TypeSubscribe)))

	userSubID := compatParseInt64String(req.UserSubscribeID)
	if userSubID == 0 {
		userSubID = compatParseInt64String(req.SubscribeID)
	}
	if userSubID > 0 {
		query = query.Where(proxysystemlog.ObjectIDEQ(userSubID))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	if req != nil && req.Page > 0 && req.Size > 0 {
		query = query.Offset((req.Page - 1) * req.Size).Limit(req.Size)
	}

	items, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := &compatUserSubscribeLogListData{Total: int64(total), List: []compatUserSubscribeLog{}}
	for _, item := range items {
		content := &logmodel.Subscribe{}
		_ = content.Unmarshal([]byte(item.Content))
		userID := int64(0)
		if item.ObjectID > 0 {
			if userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, item.ObjectID); err == nil {
				userID = userSub.UserID
			}
		}
		result.List = append(result.List, compatUserSubscribeLog{
			ID:              item.ID,
			UserID:          userID,
			UserSubscribeID: item.ObjectID,
			Token:           content.Token,
			IP:              content.ClientIP,
			UserAgent:       content.UserAgent,
			Timestamp:       item.CreatedAt.UnixMilli(),
		})
	}
	return result, nil
}

func compatLegacyAdminUserSubscribeResetTrafficLogs(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetUserSubscribeResetTrafficLogsRequest) (*compatLegacyResetSubscribeTrafficLogListData, error) {
	query := dataLayer.DB().ProxySystemLog.Query().
		Where(proxysystemlog.TypeEQ(int8(logmodel.TypeResetSubscribe)))

	if req != nil {
		if userSubID := compatParseInt64String(req.UserSubscribeID); userSubID > 0 {
			query = query.Where(proxysystemlog.ObjectIDEQ(userSubID))
		}
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	if req != nil && req.Page > 0 && req.Size > 0 {
		query = query.Offset((req.Page - 1) * req.Size).Limit(req.Size)
	}

	items, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := &compatLegacyResetSubscribeTrafficLogListData{Total: int64(total), List: []compatLegacyResetSubscribeTrafficLog{}}
	for _, item := range items {
		content := &logmodel.ResetSubscribe{}
		_ = content.Unmarshal([]byte(item.Content))
		result.List = append(result.List, compatLegacyResetSubscribeTrafficLog{
			ID:              item.ID,
			Type:            content.Type,
			UserSubscribeID: item.ObjectID,
			OrderNo:         content.OrderNo,
			Timestamp:       content.Timestamp,
		})
	}
	return result, nil
}

func compatLegacyAdminUserSubscribeTrafficLogs(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetUserSubscribeTrafficLogsRequest) (*compatLegacyTrafficLogListData, error) {
	query := dataLayer.DB().ProxyTrafficLog.Query()

	if userID := compatParseInt64String(req.UserID); userID > 0 {
		query = query.Where(proxytrafficlog.UserIDEQ(userID))
	}

	subscribeID := compatParseInt64String(req.SubscribeID)
	if userSubID := compatParseInt64String(req.UserSubscribeID); userSubID > 0 {
		if userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubID); err == nil {
			subscribeID = userSub.SubscribeID
		}
	}
	if subscribeID > 0 {
		query = query.Where(proxytrafficlog.SubscribeIDEQ(subscribeID))
	}

	if req.StartTime > 0 {
		query = query.Where(proxytrafficlog.TimestampGTE(time.UnixMilli(req.StartTime)))
	}
	if req.EndTime > 0 {
		query = query.Where(proxytrafficlog.TimestampLTE(time.UnixMilli(req.EndTime)))
	}
	if start, end, ok := compatParseDateRange(req.Date); ok {
		query = query.Where(proxytrafficlog.TimestampGTE(start), proxytrafficlog.TimestampLT(end))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	if req.Page > 0 && req.Size > 0 {
		query = query.Offset((req.Page - 1) * req.Size).Limit(req.Size)
	}

	items, err := query.
		Order(ent.Desc(proxytrafficlog.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := &compatLegacyTrafficLogListData{Total: int64(total), List: []compatLegacyTrafficLog{}}
	for _, item := range items {
		result.List = append(result.List, compatLegacyTrafficLog{
			ID:          item.ID,
			ServerID:    item.ServerID,
			UserID:      item.UserID,
			SubscribeID: item.SubscribeID,
			Download:    item.Download,
			Upload:      item.Upload,
			Timestamp:   item.Timestamp.UnixMilli(),
		})
	}
	return result, nil
}

func compatResetLegacyUserSubscribeToken(ctx context.Context, dataLayer *data.Data, userSubscribeID int64) error {
	userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubscribeID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("FindOneSubscribe error: %v", err))
	}
	token := uuidx.SubscribeToken(fmt.Sprintf("AdminUpdate:%d", time.Now().UnixMilli()))
	if err := dataLayer.DB().ProxyUserSubscribe.UpdateOneID(userSubscribeID).
		SetToken(token).
		SetUpdatedAt(time.Now()).
		Exec(ctx); err != nil {
		return compatCodeError(500, fmt.Sprintf("UpdateSubscribe error: %v", err))
	}
	return compatClearLegacyUserCaches(ctx, dataLayer, userSub.UserID, userSub.SubscribeID)
}

func compatToggleLegacyUserSubscribeStatus(ctx context.Context, dataLayer *data.Data, userSubscribeID int64) error {
	userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubscribeID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("FindOneSubscribe error: %v", err))
	}

	currentStatus := compatInt8Value(userSub.Status)
	nextStatus := currentStatus
	switch currentStatus {
	case 1, 2:
		nextStatus = 5
	case 5:
		nextStatus = 1
	default:
		return compatCodeError(500, fmt.Sprintf("invalid user subscribe status: %d", currentStatus))
	}

	if err := dataLayer.DB().ProxyUserSubscribe.UpdateOneID(userSubscribeID).
		SetStatus(nextStatus).
		SetUpdatedAt(time.Now()).
		Exec(ctx); err != nil {
		return compatCodeError(500, fmt.Sprintf("UpdateSubscribe error: %v", err))
	}
	return compatClearLegacyUserCaches(ctx, dataLayer, userSub.UserID, userSub.SubscribeID)
}

func compatResetLegacyUserSubscribeTraffic(ctx context.Context, dataLayer *data.Data, userSubscribeID int64) error {
	userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubscribeID)
	if err != nil {
		return compatCodeError(500, fmt.Sprintf("FindOneSubscribe error: %v", err))
	}
	if err := dataLayer.DB().ProxyUserSubscribe.UpdateOneID(userSubscribeID).
		SetDownload(0).
		SetUpload(0).
		SetUpdatedAt(time.Now()).
		Exec(ctx); err != nil {
		return compatCodeError(500, fmt.Sprintf("UpdateSubscribe error: %v", err))
	}
	return compatClearLegacyUserCaches(ctx, dataLayer, userSub.UserID, userSub.SubscribeID)
}

func compatLegacyAdminUserFromEntity(ctx context.Context, dataLayer *data.Data, item *ent.ProxyUser) (*compatLegacyAdminUser, error) {
	methods, err := dataLayer.DB().ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserIDEQ(item.ID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	devices, err := dataLayer.DB().ProxyUserDevice.Query().
		Where(proxyuserdevice.UserIDEQ(item.ID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	authMethods := make([]compatUserAuthMethod, 0, len(methods))
	for _, method := range methods {
		authMethods = append(authMethods, compatUserAuthMethod{
			AuthType:       method.AuthType,
			AuthIdentifier: compatLegacyAdminAuthIdentifier(method),
			Verified:       method.Verified,
		})
	}

	userDevices := make([]compatUserDevice, 0, len(devices))
	for _, device := range devices {
		userDevices = append(userDevices, compatUserDevice{
			ID:         device.ID,
			IP:         compatStringValue(device.IP),
			Identifier: compatStringValue(device.Identifier),
			UserAgent:  compatStringValue(device.UserAgent),
			Online:     device.Online,
			Enabled:    device.Enabled,
			CreatedAt:  device.CreatedAt.UnixMilli(),
			UpdatedAt:  device.UpdatedAt.UnixMilli(),
		})
	}

	return &compatLegacyAdminUser{
		ID:                    item.ID,
		Avatar:                compatStringValue(item.Avatar),
		Balance:               compatInt64Value(item.Balance),
		Commission:            compatInt64Value(item.Commission),
		ReferralPercentage:    uint8(item.ReferralPercentage),
		OnlyFirstPurchase:     item.OnlyFirstPurchase,
		GiftAmount:            compatInt64Value(item.GiftAmount),
		Telegram:              compatInt64Value(item.Telegram),
		ReferCode:             compatStringValue(item.ReferCode),
		RefererID:             compatInt64Value(item.RefererID),
		Enable:                item.Enable,
		IsAdmin:               item.IsAdmin,
		EnableBalanceNotify:   item.EnableBalanceNotify,
		EnableLoginNotify:     item.EnableLoginNotify,
		EnableSubscribeNotify: item.EnableSubscribeNotify,
		EnableTradeNotify:     item.EnableTradeNotify,
		AuthMethods:           authMethods,
		UserDevices:           userDevices,
		Rules:                 compatJSONStringArray(item.Rules),
		CreatedAt:             item.CreatedAt.UnixMilli(),
		UpdatedAt:             item.UpdatedAt.UnixMilli(),
		DeletedAt:             compatUnixMillis(item.DeletedAt),
		IsDel:                 item.IsDel != nil && *item.IsDel != 0,
	}, nil
}

func compatLegacyAdminAuthIdentifier(item *ent.ProxyUserAuthMethod) string {
	switch strings.ToLower(strings.TrimSpace(item.AuthType)) {
	case "mobile":
		return phone.MaskPhoneNumber(item.AuthIdentifier)
	default:
		return item.AuthIdentifier
	}
}

func compatResolveLegacyUserSubscribeDeviceTarget(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetUserSubscribeDevicesRequest) (int64, int64, error) {
	if userSubID := compatParseInt64String(req.UserSubscribeID); userSubID > 0 {
		if userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, userSubID); err == nil {
			return userSub.UserID, userSub.SubscribeID, nil
		}
	}
	if rawID := compatParseInt64String(req.SubscribeID); rawID > 0 {
		if userSub, err := dataLayer.DB().ProxyUserSubscribe.Get(ctx, rawID); err == nil {
			return userSub.UserID, userSub.SubscribeID, nil
		}
	}
	return compatParseInt64String(req.UserID), compatParseInt64String(req.SubscribeID), nil
}

func compatClearLegacyUserCaches(ctx context.Context, dataLayer *data.Data, userID int64, subscribeIDs ...int64) error {
	emails, _ := compatUserEmails(ctx, dataLayer, userID)
	compatClearUserCache(ctx, dataLayer.Redis(), userID, emails...)

	if userSubs, err := dataLayer.DB().ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.UserIDEQ(userID)).
		All(ctx); err == nil {
		for _, item := range userSubs {
			compatClearUserSubscribeCaches(ctx, dataLayer.Redis(), item)
		}
	}

	for _, subscribeID := range subscribeIDs {
		if subscribeID > 0 {
			compatClearSubscribeCaches(ctx, dataLayer.Redis(), subscribeID)
		}
	}

	if err := data.ClearLegacyServerAllCaches(ctx, dataLayer.Redis()); err != nil {
		return compatCodeError(500, fmt.Sprintf("failed to clear server cache: %v", err))
	}
	return nil
}

func compatParseDateRange(raw string) (time.Time, time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	return parsed, parsed.Add(24 * time.Hour), true
}

func compatFillLegacyUserSubscribeListQuery(ctx khttp.Context, req *compatLegacyGetUserSubscribeListRequest) {
	if ctx == nil || req == nil || ctx.Request() == nil || ctx.Request().URL == nil {
		return
	}
	query := ctx.Request().URL.Query()
	if req.Page <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("page"))); err == nil {
			req.Page = value
		}
	}
	if req.Size <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("size"))); err == nil {
			req.Size = value
		}
	}
	if strings.TrimSpace(req.UserID) == "" {
		req.UserID = strings.TrimSpace(query.Get("user_id"))
	}
}

func compatFillLegacyDeleteUserSubscribeQuery(ctx khttp.Context, req *compatLegacyDeleteUserSubscribeRequest) {
	if ctx == nil || req == nil || ctx.Request() == nil || ctx.Request().URL == nil {
		return
	}
	if strings.TrimSpace(req.UserSubscribeID) == "" {
		req.UserSubscribeID = strings.TrimSpace(ctx.Request().URL.Query().Get("user_subscribe_id"))
	}
}

func compatFillLegacyUserSubscribeDevicesQuery(ctx khttp.Context, req *compatLegacyGetUserSubscribeDevicesRequest) {
	if ctx == nil || req == nil || ctx.Request() == nil || ctx.Request().URL == nil {
		return
	}
	query := ctx.Request().URL.Query()
	if req.Page <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("page"))); err == nil {
			req.Page = value
		}
	}
	if req.Size <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("size"))); err == nil {
			req.Size = value
		}
	}
	if strings.TrimSpace(req.UserID) == "" {
		req.UserID = strings.TrimSpace(query.Get("user_id"))
	}
	if strings.TrimSpace(req.SubscribeID) == "" {
		req.SubscribeID = strings.TrimSpace(query.Get("subscribe_id"))
	}
	if strings.TrimSpace(req.UserSubscribeID) == "" {
		req.UserSubscribeID = strings.TrimSpace(query.Get("user_subscribe_id"))
	}
}

func compatFillLegacyUserSubscribeLogsQuery(ctx khttp.Context, req *compatLegacyGetUserSubscribeLogsRequest) {
	if ctx == nil || req == nil || ctx.Request() == nil || ctx.Request().URL == nil {
		return
	}
	query := ctx.Request().URL.Query()
	if req.Page <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("page"))); err == nil {
			req.Page = value
		}
	}
	if req.Size <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("size"))); err == nil {
			req.Size = value
		}
	}
	if strings.TrimSpace(req.UserID) == "" {
		req.UserID = strings.TrimSpace(query.Get("user_id"))
	}
	if strings.TrimSpace(req.SubscribeID) == "" {
		req.SubscribeID = strings.TrimSpace(query.Get("subscribe_id"))
	}
	if strings.TrimSpace(req.UserSubscribeID) == "" {
		req.UserSubscribeID = strings.TrimSpace(query.Get("user_subscribe_id"))
	}
}

func compatFillLegacyUserSubscribeResetTrafficLogsQuery(ctx khttp.Context, req *compatLegacyGetUserSubscribeResetTrafficLogsRequest) {
	if ctx == nil || req == nil || ctx.Request() == nil || ctx.Request().URL == nil {
		return
	}
	query := ctx.Request().URL.Query()
	if req.Page <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("page"))); err == nil {
			req.Page = value
		}
	}
	if req.Size <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("size"))); err == nil {
			req.Size = value
		}
	}
	if strings.TrimSpace(req.UserSubscribeID) == "" {
		req.UserSubscribeID = strings.TrimSpace(query.Get("user_subscribe_id"))
	}
}

func compatFillLegacyUserSubscribeTrafficLogsQuery(ctx khttp.Context, req *compatLegacyGetUserSubscribeTrafficLogsRequest) {
	if ctx == nil || req == nil || ctx.Request() == nil || ctx.Request().URL == nil {
		return
	}
	query := ctx.Request().URL.Query()
	if req.Page <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("page"))); err == nil {
			req.Page = value
		}
	}
	if req.Size <= 0 {
		if value, err := strconv.Atoi(strings.TrimSpace(query.Get("size"))); err == nil {
			req.Size = value
		}
	}
	if strings.TrimSpace(req.UserID) == "" {
		req.UserID = strings.TrimSpace(query.Get("user_id"))
	}
	if strings.TrimSpace(req.SubscribeID) == "" {
		req.SubscribeID = strings.TrimSpace(query.Get("subscribe_id"))
	}
	if strings.TrimSpace(req.UserSubscribeID) == "" {
		req.UserSubscribeID = strings.TrimSpace(query.Get("user_subscribe_id"))
	}
	if req.StartTime == 0 {
		if value, err := strconv.ParseInt(strings.TrimSpace(query.Get("start_time")), 10, 64); err == nil {
			req.StartTime = value
		}
	}
	if req.EndTime == 0 {
		if value, err := strconv.ParseInt(strings.TrimSpace(query.Get("end_time")), 10, 64); err == nil {
			req.EndTime = value
		}
	}
	if strings.TrimSpace(req.Date) == "" {
		req.Date = strings.TrimSpace(query.Get("date"))
	}
}
