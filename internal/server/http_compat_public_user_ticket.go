package server

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	publicticketv1 "github.com/OmnTeam/ppanel-pro/api/public/ticket/v1"
	publicuserv1 "github.com/OmnTeam/ppanel-pro/api/public/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/OmnTeam/ppanel-pro/pkg/phone"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type legacyPublicTicketCompat interface {
	CreateUserTicket(context.Context, *publicticketv1.CreateUserTicketRequest) (*publicticketv1.TicketCreateReply, error)
	GetUserTicketList(context.Context, *publicticketv1.GetUserTicketListRequest) (*publicticketv1.TicketListReply, error)
	GetUserTicketDetails(context.Context, *publicticketv1.GetUserTicketDetailsRequest) (*publicticketv1.TicketDetailReply, error)
	UpdateUserTicketStatus(context.Context, *publicticketv1.UpdateUserTicketStatusRequest) (*publicticketv1.TicketStatusReply, error)
	CreateUserTicketFollow(context.Context, *publicticketv1.CreateUserTicketFollowRequest) (*publicticketv1.TicketFollowReply, error)
}

type legacyPublicUserCompat interface {
	GetLoginLog(context.Context, *publicuserv1.GetLoginLogRequest) (*publicuserv1.LoginLogReply, error)
	QueryUserBalanceLog(context.Context, *emptypb.Empty) (*publicuserv1.BalanceLogReply, error)
	QueryUserCommissionLog(context.Context, *publicuserv1.QueryUserCommissionLogRequest) (*publicuserv1.CommissionLogReply, error)
	QueryUserAffiliate(context.Context, *emptypb.Empty) (*publicuserv1.UserAffiliateReply, error)
	QueryUserAffiliateList(context.Context, *publicuserv1.QueryUserAffiliateListRequest) (*publicuserv1.UserAffiliateListReply, error)
	GetOAuthMethods(context.Context, *emptypb.Empty) (*publicuserv1.OAuthMethodsReply, error)
	GetSubscribeLog(context.Context, *publicuserv1.GetSubscribeLogRequest) (*publicuserv1.SubscribeLogReply, error)
	ResetUserSubscribeToken(context.Context, *publicuserv1.ResetUserSubscribeTokenRequest) (*publicuserv1.CommonReply, error)
	PreUnsubscribe(context.Context, *publicuserv1.PreUnsubscribeRequest) (*publicuserv1.UnsubscribeInfoReply, error)
	Unsubscribe(context.Context, *publicuserv1.UnsubscribeRequest) (*publicuserv1.CommonReply, error)
	UpdateUserNotify(context.Context, *publicuserv1.UpdateUserNotifyRequest) (*publicuserv1.CommonReply, error)
	UpdateUserPassword(context.Context, *publicuserv1.UpdateUserPasswordRequest) (*publicuserv1.CommonReply, error)
	BindTelegram(context.Context, *emptypb.Empty) (*publicuserv1.TelegramBindReply, error)
	UnbindTelegram(context.Context, *emptypb.Empty) (*publicuserv1.CommonReply, error)
	BindOAuth(context.Context, *publicuserv1.BindOAuthRequest) (*publicuserv1.OAuthBindReply, error)
	BindOAuthCallback(context.Context, *publicuserv1.BindOAuthCallbackRequest) (*publicuserv1.CommonReply, error)
	UnbindOAuth(context.Context, *publicuserv1.UnbindOAuthRequest) (*publicuserv1.CommonReply, error)
	VerifyEmail(context.Context, *publicuserv1.VerifyEmailRequest) (*publicuserv1.CommonReply, error)
	UpdateBindMobile(context.Context, *publicuserv1.UpdateBindMobileRequest) (*publicuserv1.CommonReply, error)
	UpdateBindEmail(context.Context, *publicuserv1.UpdateBindEmailRequest) (*publicuserv1.CommonReply, error)
	GetDeviceList(context.Context, *emptypb.Empty) (*publicuserv1.GetDeviceListReply, error)
	UnbindDevice(context.Context, *publicuserv1.UnbindDeviceRequest) (*publicuserv1.CommonReply, error)
	GetDeviceOnlineStatistics(context.Context, *emptypb.Empty) (*publicuserv1.GetDeviceOnlineStatisticsReply, error)
	CommissionWithdraw(context.Context, *publicuserv1.CommissionWithdrawRequest) (*publicuserv1.WithdrawalLogReply, error)
	QueryWithdrawalLog(context.Context, *publicuserv1.QueryWithdrawalLogRequest) (*publicuserv1.WithdrawalLogListReply, error)
}

type compatQuerySubscribeListRequest struct {
	Language string `form:"language"`
}

type compatTicketListRequest struct {
	Page   int64  `form:"page"`
	Size   int64  `form:"size"`
	Status *int32 `form:"status"`
	Search string `form:"search"`
}

type compatTicketDetailRequest struct {
	ID int64 `form:"id"`
}

type compatTicketFollowRequest struct {
	TicketID int64  `json:"ticket_id"`
	From     string `json:"from"`
	Type     int32  `json:"type"`
	Content  string `json:"content"`
}

type compatTicketStatusRequest struct {
	ID     int64  `json:"id"`
	Status *int32 `json:"status"`
}

type compatLoginLogRequest struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type compatCommissionLogRequest struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type compatAffiliateListRequest struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type compatSubscribeLogRequest struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type compatResetSubscribeTokenRequest struct {
	UserSubscribeID int64 `json:"user_subscribe_id"`
}

type compatPreUnsubscribeRequest struct {
	ID int64 `json:"id"`
}

type compatUnsubscribeRequest struct {
	ID int64 `json:"id"`
}

type compatUpdateUserNotifyRequest struct {
	EnableLoginNotify     bool `json:"enable_login_notify"`
	EnableBalanceNotify   bool `json:"enable_balance_notify"`
	EnableSubscribeNotify bool `json:"enable_subscribe_notify"`
	EnableTradeNotify     bool `json:"enable_trade_notify"`
}

type compatUpdateUserPasswordRequest struct {
	Password string `json:"password"`
}

type compatBindOAuthRequest struct {
	Method   string `json:"method"`
	Redirect string `json:"redirect"`
}

type compatBindOAuthCallbackRequest struct {
	Method   string          `json:"method"`
	Callback json.RawMessage `json:"callback"`
}

type compatUnbindOAuthRequest struct {
	Method string `json:"method"`
}

type compatVerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type compatBindMobileRequest struct {
	AreaCode string `json:"area_code"`
	Mobile   string `json:"mobile"`
	Code     string `json:"code"`
}

type compatBindEmailRequest struct {
	Email string `json:"email"`
}

type compatUnbindDeviceRequest struct {
	ID int64 `json:"id"`
}

type compatCommissionWithdrawRequest struct {
	Amount  int64  `json:"amount"`
	Content string `json:"content"`
}

type compatWithdrawalLogRequest struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type compatTicketFollow struct {
	ID        int64  `json:"id"`
	TicketID  int64  `json:"ticket_id"`
	From      string `json:"from"`
	Type      uint8  `json:"type"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

type compatTicket struct {
	ID          int64                `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	UserID      int64                `json:"user_id"`
	Follows     []compatTicketFollow `json:"follow,omitempty"`
	Status      uint8                `json:"status"`
	CreatedAt   int64                `json:"created_at"`
	UpdatedAt   int64                `json:"updated_at"`
}

type compatTicketListData struct {
	Total int64          `json:"total"`
	List  []compatTicket `json:"list"`
}

type compatUserAuthMethod struct {
	AuthType       string `json:"auth_type"`
	AuthIdentifier string `json:"auth_identifier"`
	Verified       bool   `json:"verified"`
}

type compatUserDevice struct {
	ID         int64  `json:"id"`
	IP         string `json:"ip"`
	Identifier string `json:"identifier"`
	UserAgent  string `json:"user_agent"`
	Online     bool   `json:"online"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type compatUserInfo struct {
	ID                    int64                  `json:"id"`
	Avatar                string                 `json:"avatar"`
	Balance               int64                  `json:"balance"`
	Commission            int64                  `json:"commission"`
	ReferralPercentage    uint8                  `json:"referral_percentage"`
	OnlyFirstPurchase     bool                   `json:"only_first_purchase"`
	GiftAmount            int64                  `json:"gift_amount"`
	Telegram              int64                  `json:"telegram"`
	ReferCode             string                 `json:"refer_code"`
	RefererID             int64                  `json:"referer_id"`
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

type compatLoginLog struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	LoginIP   string `json:"login_ip"`
	UserAgent string `json:"user_agent"`
	Success   bool   `json:"success"`
	Timestamp int64  `json:"timestamp"`
}

type compatLoginLogListData struct {
	Total int64            `json:"total"`
	List  []compatLoginLog `json:"list"`
}

type compatBalanceLog struct {
	Type      uint16 `json:"type"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
	OrderNo   string `json:"order_no"`
	Balance   int64  `json:"balance"`
	Timestamp int64  `json:"timestamp"`
}

type compatBalanceLogListData struct {
	Total int64              `json:"total"`
	List  []compatBalanceLog `json:"list"`
}

type compatCommissionLog struct {
	Type      uint16 `json:"type"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
	OrderNo   string `json:"order_no"`
	Timestamp int64  `json:"timestamp"`
}

type compatCommissionLogListData struct {
	Total int64                 `json:"total"`
	List  []compatCommissionLog `json:"list"`
}

type compatAffiliateCountData struct {
	Registers       int64 `json:"registers"`
	TotalCommission int64 `json:"total_commission"`
}

type compatUserAffiliate struct {
	Avatar       string `json:"avatar"`
	Identifier   string `json:"identifier"`
	RegisteredAt int64  `json:"registered_at"`
	Enable       bool   `json:"enable"`
}

type compatAffiliateListData struct {
	Total int64                 `json:"total"`
	List  []compatUserAffiliate `json:"list"`
}

type compatOAuthMethodsData struct {
	Methods []compatUserAuthMethod `json:"methods"`
}

type compatUserSubscribe struct {
	ID          int64           `json:"id"`
	IDStr       string          `json:"id_str"`
	UserID      int64           `json:"user_id"`
	OrderID     int64           `json:"order_id"`
	SubscribeID int64           `json:"subscribe_id"`
	Subscribe   compatSubscribe `json:"subscribe"`
	NodeGroupID int64           `json:"node_group_id"`
	GroupLocked bool            `json:"group_locked"`
	StartTime   int64           `json:"start_time"`
	ExpireTime  int64           `json:"expire_time"`
	FinishedAt  int64           `json:"finished_at"`
	ResetTime   int64           `json:"reset_time"`
	Traffic     int64           `json:"traffic"`
	Download    int64           `json:"download"`
	Upload      int64           `json:"upload"`
	Token       string          `json:"token"`
	Status      uint8           `json:"status"`
	Short       string          `json:"short"`
	CreatedAt   int64           `json:"created_at"`
	UpdatedAt   int64           `json:"updated_at"`
}

type compatUserSubscribeListData struct {
	Total int64                 `json:"total"`
	List  []compatUserSubscribe `json:"list"`
}

type compatUserSubscribeLog struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"user_id"`
	UserSubscribeID int64  `json:"user_subscribe_id"`
	Token           string `json:"token"`
	IP              string `json:"ip"`
	UserAgent       string `json:"user_agent"`
	Timestamp       int64  `json:"timestamp"`
}

type compatUserSubscribeLogListData struct {
	Total int64                    `json:"total"`
	List  []compatUserSubscribeLog `json:"list"`
}

type compatPreUnsubscribeData struct {
	DeductionAmount int64 `json:"deduction_amount"`
}

type compatBindTelegramData struct {
	URL       string `json:"url"`
	ExpiredAt int64  `json:"expired_at"`
}

type compatBindOAuthData struct {
	Redirect string `json:"redirect"`
}

type compatUserDeviceListData struct {
	Total int64              `json:"total"`
	List  []compatUserDevice `json:"list"`
}

type compatWeeklyStat struct {
	Day     int32   `json:"day"`
	DayName string  `json:"day_name"`
	Hours   float64 `json:"hours"`
}

type compatConnectionRecords struct {
	CurrentContinuousDays   int64 `json:"current_continuous_days"`
	HistoryContinuousDays   int64 `json:"history_continuous_days"`
	LongestSingleConnection int64 `json:"longest_single_connection"`
}

type compatDeviceOnlineStatisticsData struct {
	WeeklyStats       []compatWeeklyStat      `json:"weekly_stats"`
	ConnectionRecords compatConnectionRecords `json:"connection_records"`
}

type compatWithdrawalLog struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
	Content   string `json:"content"`
	Status    uint8  `json:"status"`
	Reason    string `json:"reason,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type compatWithdrawalLogListData struct {
	Total int64                 `json:"total"`
	List  []compatWithdrawalLog `json:"list"`
}

func registerLegacyPublicSubscribeCompatRoutes(r *khttp.Router, dataLayer *data.Data) {
	if r == nil || dataLayer == nil {
		return
	}

	r.GET("/v1/public/subscribe/list", func(ctx khttp.Context) error {
		var req compatQuerySubscribeListRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatQuerySubscribeListRequest)
			list, total, queryErr := compatLegacyQuerySubscribeList(inner, dataLayer, in.Language)
			if queryErr != nil {
				return nil, queryErr
			}
			return map[string]interface{}{
				"list":  list,
				"total": total,
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})
}

func registerLegacyPublicTicketCompatRoutes(r *khttp.Router, publicTicket legacyPublicTicketCompat) {
	if r == nil || publicTicket == nil {
		return
	}

	r.POST("/v1/public/ticket", func(ctx khttp.Context) error {
		var req publicticketv1.CreateUserTicketRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			_, err := publicTicket.CreateUserTicket(inner, request.(*publicticketv1.CreateUserTicketRequest))
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.GET("/v1/public/ticket/list", func(ctx khttp.Context) error {
		var req compatTicketListRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatTicketListRequest)
			protoReq := &publicticketv1.GetUserTicketListRequest{
				Page: int32(in.Page),
				Size: int32(in.Size),
			}
			if in.Status != nil {
				protoReq.Status = in.Status
			}
			if strings.TrimSpace(in.Search) != "" {
				search := strings.TrimSpace(in.Search)
				protoReq.Search = &search
			}
			reply, err := publicTicket.GetUserTicketList(inner, protoReq)
			if err != nil {
				return nil, err
			}
			return compatLegacyTicketList(reply), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/ticket/detail", func(ctx khttp.Context) error {
		var req compatTicketDetailRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if req.ID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'GetUserTicketDetailRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"))
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatTicketDetailRequest)
			reply, err := publicTicket.GetUserTicketDetails(inner, &publicticketv1.GetUserTicketDetailsRequest{
				Id: strconv.FormatInt(in.ID, 10),
			})
			if err != nil {
				return nil, err
			}
			return compatLegacyTicket(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.PUT("/v1/public/ticket", func(ctx khttp.Context) error {
		var req compatTicketStatusRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if req.ID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateUserTicketStatusRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"))
		}
		if req.Status == nil {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateUserTicketStatusRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"))
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatTicketStatusRequest)
			_, err := publicTicket.UpdateUserTicketStatus(inner, &publicticketv1.UpdateUserTicketStatusRequest{
				Id:     strconv.FormatInt(in.ID, 10),
				Status: *in.Status,
			})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.POST("/v1/public/ticket/follow", func(ctx khttp.Context) error {
		var req compatTicketFollowRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if req.TicketID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'CreateUserTicketFollowRequest.TicketId' Error:Field validation for 'TicketId' failed on the 'required' tag"))
		}
		if err := compatValidateRequiredString(req.From, "CreateUserTicketFollowRequest", "From"); err != nil {
			return compatJSONError(ctx, err)
		}
		if req.Type == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'CreateUserTicketFollowRequest.Type' Error:Field validation for 'Type' failed on the 'required' tag"))
		}
		if err := compatValidateRequiredString(req.Content, "CreateUserTicketFollowRequest", "Content"); err != nil {
			return compatJSONError(ctx, err)
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatTicketFollowRequest)
			_, err := publicTicket.CreateUserTicketFollow(inner, &publicticketv1.CreateUserTicketFollowRequest{
				TicketId: strconv.FormatInt(in.TicketID, 10),
				From:     in.From,
				Type:     in.Type,
				Content:  in.Content,
			})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})
}

func compatLegacyQuerySubscribeList(ctx context.Context, dataLayer *data.Data, language string) ([]compatSubscribe, int64, error) {
	querySubscribes := func(lang string) ([]*ent.ProxySubscribe, int, error) {
		query := dataLayer.DB().ProxySubscribe.Query().
			Where(proxysubscribe.SellEQ(true))
		if strings.TrimSpace(lang) != "" {
			query = query.Where(proxysubscribe.LanguageEQ(strings.TrimSpace(lang)))
		} else {
			query = query.Where(proxysubscribe.LanguageEQ(""))
		}

		total, err := query.Count(ctx)
		if err != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		items, err := query.Order(ent.Asc(proxysubscribe.FieldSort)).All(ctx)
		if err != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		return items, total, nil
	}

	items, total, err := querySubscribes(language)
	if err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(language) != "" && total == 0 {
		items, total, err = querySubscribes("")
		if err != nil {
			return nil, 0, err
		}
	}

	list := make([]compatSubscribe, 0, len(items))
	for _, item := range items {
		list = append(list, compatLegacySubscribeFromEntity(item))
	}
	return list, int64(total), nil
}

func compatLegacyTicketList(reply *publicticketv1.TicketListReply) *compatTicketListData {
	result := &compatTicketListData{List: []compatTicket{}}
	if reply == nil || reply.Data == nil {
		return result
	}
	result.Total = int64(reply.Data.Total)
	for _, item := range reply.Data.List {
		result.List = append(result.List, compatLegacyTicket(item))
	}
	return result
}

func compatLegacyTicket(item *publicticketv1.TicketInfo) compatTicket {
	result := compatTicket{Follows: []compatTicketFollow{}}
	if item == nil {
		return result
	}
	result.ID = compatParseInt64String(item.Id)
	result.Title = item.Title
	result.Description = item.Description
	result.UserID = compatParseInt64String(item.UserId)
	result.Status = uint8(item.Status)
	result.CreatedAt = compatParseInt64String(item.CreatedAt)
	result.UpdatedAt = compatParseInt64String(item.UpdatedAt)
	for _, follow := range item.Follow {
		result.Follows = append(result.Follows, compatTicketFollow{
			ID:        compatParseInt64String(follow.Id),
			TicketID:  compatParseInt64String(follow.TicketId),
			From:      follow.From,
			Type:      uint8(follow.Type),
			Content:   follow.Content,
			CreatedAt: compatParseInt64String(follow.CreatedAt),
		})
	}
	return result
}

func registerLegacyPublicUserCompatRoutes(r *khttp.Router, dataLayer *data.Data, publicUser legacyPublicUserCompat) {
	if r == nil || dataLayer == nil || publicUser == nil {
		return
	}

	r.GET("/v1/public/user/info", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			return compatLegacyUserInfo(inner, dataLayer)
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/login_log", func(ctx khttp.Context) error {
		var req compatLoginLogRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatLoginLogRequest)
			reply, err := publicUser.GetLoginLog(inner, &publicuserv1.GetLoginLogRequest{Page: in.Page, Size: in.Size})
			if err != nil {
				return nil, err
			}
			return compatLegacyLoginLog(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/balance_log", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			reply, err := publicUser.QueryUserBalanceLog(inner, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}
			return compatLegacyBalanceLog(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/commission_log", func(ctx khttp.Context) error {
		var req compatCommissionLogRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatCommissionLogRequest)
			reply, err := publicUser.QueryUserCommissionLog(inner, &publicuserv1.QueryUserCommissionLogRequest{Page: in.Page, Size: in.Size})
			if err != nil {
				return nil, err
			}
			return compatLegacyCommissionLog(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/affiliate/count", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			reply, err := publicUser.QueryUserAffiliate(inner, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}
			return compatLegacyAffiliateCount(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/affiliate/list", func(ctx khttp.Context) error {
		var req compatAffiliateListRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatAffiliateListRequest)
			reply, err := publicUser.QueryUserAffiliateList(inner, &publicuserv1.QueryUserAffiliateListRequest{Page: in.Page, Size: in.Size})
			if err != nil {
				return nil, err
			}
			return compatLegacyAffiliateList(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/oauth_methods", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			reply, err := publicUser.GetOAuthMethods(inner, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}
			return compatLegacyOAuthMethods(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/subscribe", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			return compatLegacyUserSubscribe(inner, dataLayer)
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/subscribe_log", func(ctx khttp.Context) error {
		var req compatSubscribeLogRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatSubscribeLogRequest)
			reply, err := publicUser.GetSubscribeLog(inner, &publicuserv1.GetSubscribeLogRequest{Page: in.Page, Size: in.Size})
			if err != nil {
				return nil, err
			}
			return compatLegacySubscribeLog(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.PUT("/v1/public/user/subscribe_token", func(ctx khttp.Context) error {
		var req compatResetSubscribeTokenRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if req.UserSubscribeID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'ResetUserSubscribeTokenRequest.UserSubscribeId' Error:Field validation for 'UserSubscribeId' failed on the 'required' tag"))
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatResetSubscribeTokenRequest)
			_, err := publicUser.ResetUserSubscribeToken(inner, &publicuserv1.ResetUserSubscribeTokenRequest{UserSubscribeId: in.UserSubscribeID})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.POST("/v1/public/user/unsubscribe/pre", func(ctx khttp.Context) error {
		var req compatPreUnsubscribeRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatPreUnsubscribeRequest)
			reply, err := publicUser.PreUnsubscribe(inner, &publicuserv1.PreUnsubscribeRequest{Id: in.ID})
			if err != nil {
				return nil, err
			}
			return compatLegacyPreUnsubscribe(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/public/user/unsubscribe", func(ctx khttp.Context) error {
		var req compatUnsubscribeRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatUnsubscribeRequest)
			_, err := publicUser.Unsubscribe(inner, &publicuserv1.UnsubscribeRequest{Id: in.ID})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.PUT("/v1/public/user/notify", func(ctx khttp.Context) error {
		var req compatUpdateUserNotifyRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatUpdateUserNotifyRequest)
			_, err := publicUser.UpdateUserNotify(inner, &publicuserv1.UpdateUserNotifyRequest{
				EnableLoginNotify:     in.EnableLoginNotify,
				EnableBalanceNotify:   in.EnableBalanceNotify,
				EnableSubscribeNotify: in.EnableSubscribeNotify,
				EnableTradeNotify:     in.EnableTradeNotify,
			})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.PUT("/v1/public/user/password", func(ctx khttp.Context) error {
		var req compatUpdateUserPasswordRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if err := compatValidateRequiredString(req.Password, "UpdateUserPasswordRequest", "Password"); err != nil {
			return compatJSONError(ctx, err)
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatUpdateUserPasswordRequest)
			_, err := publicUser.UpdateUserPassword(inner, &publicuserv1.UpdateUserPasswordRequest{Password: in.Password})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.GET("/v1/public/user/bind_telegram", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			reply, err := publicUser.BindTelegram(inner, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}
			return compatLegacyBindTelegram(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/public/user/unbind_telegram", func(ctx khttp.Context) error {
		_, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			_, err := publicUser.UnbindTelegram(inner, &emptypb.Empty{})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.POST("/v1/public/user/bind_oauth", func(ctx khttp.Context) error {
		var req compatBindOAuthRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatBindOAuthRequest)
			reply, err := publicUser.BindOAuth(inner, &publicuserv1.BindOAuthRequest{
				Method:   in.Method,
				Redirect: in.Redirect,
			})
			if err != nil {
				return nil, err
			}
			return compatLegacyBindOAuth(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/public/user/bind_oauth/callback", func(ctx khttp.Context) error {
		var req compatBindOAuthCallbackRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatBindOAuthCallbackRequest)
			callback := strings.TrimSpace(string(in.Callback))
			if callback == "" || callback == "null" {
				callback = "{}"
			}
			_, err := publicUser.BindOAuthCallback(inner, &publicuserv1.BindOAuthCallbackRequest{
				Method:   in.Method,
				Callback: callback,
			})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.POST("/v1/public/user/unbind_oauth", func(ctx khttp.Context) error {
		var req compatUnbindOAuthRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatUnbindOAuthRequest)
			_, err := publicUser.UnbindOAuth(inner, &publicuserv1.UnbindOAuthRequest{Method: in.Method})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.POST("/v1/public/user/verify_email", func(ctx khttp.Context) error {
		var req compatVerifyEmailRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if err := compatValidateRequiredString(req.Email, "VerifyEmailRequest", "Email"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Code, "VerifyEmailRequest", "Code"); err != nil {
			return compatJSONError(ctx, err)
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatVerifyEmailRequest)
			_, err := publicUser.VerifyEmail(inner, &publicuserv1.VerifyEmailRequest{
				Email: in.Email,
				Code:  in.Code,
			})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.PUT("/v1/public/user/bind_mobile", func(ctx khttp.Context) error {
		var req compatBindMobileRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if err := compatValidateRequiredString(req.AreaCode, "UpdateBindMobileRequest", "AreaCode"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Mobile, "UpdateBindMobileRequest", "Mobile"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Code, "UpdateBindMobileRequest", "Code"); err != nil {
			return compatJSONError(ctx, err)
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatBindMobileRequest)
			_, err := publicUser.UpdateBindMobile(inner, &publicuserv1.UpdateBindMobileRequest{
				AreaCode: in.AreaCode,
				Mobile:   in.Mobile,
				Code:     in.Code,
			})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.PUT("/v1/public/user/bind_email", func(ctx khttp.Context) error {
		var req compatBindEmailRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if err := compatValidateRequiredString(req.Email, "UpdateBindEmailRequest", "Email"); err != nil {
			return compatJSONError(ctx, err)
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatBindEmailRequest)
			_, err := publicUser.UpdateBindEmail(inner, &publicuserv1.UpdateBindEmailRequest{Email: in.Email})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.GET("/v1/public/user/devices", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			reply, err := publicUser.GetDeviceList(inner, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}
			return compatLegacyDeviceList(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.PUT("/v1/public/user/unbind_device", func(ctx khttp.Context) error {
		var req compatUnbindDeviceRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if req.ID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'UnbindDeviceRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"))
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatUnbindDeviceRequest)
			_, err := publicUser.UnbindDevice(inner, &publicuserv1.UnbindDeviceRequest{Id: in.ID})
			return nil, err
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.GET("/v1/public/user/device_online_statistics", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			reply, err := publicUser.GetDeviceOnlineStatistics(inner, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}
			return compatLegacyDeviceOnlineStatistics(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/device_ws_connect", func(ctx khttp.Context) error {
		_, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			return nil, compatLegacyDeviceWSConnect(inner, ctx, dataLayer)
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return nil
	})

	r.POST("/v1/public/user/commission_withdraw", func(ctx khttp.Context) error {
		var req compatCommissionWithdrawRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatCommissionWithdrawRequest)
			reply, err := publicUser.CommissionWithdraw(inner, &publicuserv1.CommissionWithdrawRequest{
				Amount:  strconv.FormatInt(in.Amount, 10),
				Content: in.Content,
			})
			if err != nil {
				return nil, err
			}
			return compatLegacyCreatedWithdrawal(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/withdrawal_log", func(ctx khttp.Context) error {
		var req compatWithdrawalLogRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatWithdrawalLogRequest)
			reply, err := publicUser.QueryWithdrawalLog(inner, &publicuserv1.QueryWithdrawalLogRequest{Page: in.Page, Size: in.Size})
			if err != nil {
				return nil, err
			}
			return compatLegacyWithdrawalLog(reply.GetData()), nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})
}

func compatLegacySubscribeFromEntity(item *ent.ProxySubscribe) compatSubscribe {
	if item == nil {
		return compatSubscribe{}
	}

	nodeGroupID := int64(0)
	if item.NodeGroupID != nil {
		nodeGroupID = *item.NodeGroupID
	}
	deductionRatio := int64(0)
	if item.DeductionRatio != nil {
		deductionRatio = *item.DeductionRatio
	}
	resetCycle := int64(0)
	if item.ResetCycle != nil {
		resetCycle = *item.ResetCycle
	}

	return compatSubscribe{
		ID:                item.ID,
		Name:              item.Name,
		Language:          item.Language,
		Description:       compatStringPointer(item.Description),
		UnitPrice:         item.UnitPrice,
		UnitTime:          item.UnitTime,
		Discount:          compatLegacySubscribeDiscounts(item.Discount),
		Replacement:       item.Replacement,
		Inventory:         item.Inventory,
		Traffic:           item.Traffic,
		SpeedLimit:        item.SpeedLimit,
		DeviceLimit:       item.DeviceLimit,
		Quota:             item.Quota,
		Nodes:             tool.StringToInt64Slice(item.Nodes),
		NodeTags:          compatSplitAndUniqueCSV(item.NodeTags),
		NodeGroupIds:      append([]int64{}, item.NodeGroupIds...),
		NodeGroupId:       nodeGroupID,
		TrafficLimit:      compatLegacyTrafficLimits(item.TrafficLimit),
		Show:              item.Show,
		Sell:              item.Sell,
		Sort:              item.Sort,
		DeductionRatio:    deductionRatio,
		AllowDeduction:    item.AllowDeduction,
		ResetCycle:        resetCycle,
		RenewalReset:      item.RenewalReset,
		ShowOriginalPrice: item.ShowOriginalPrice,
		CreatedAt:         item.CreatedAt.UnixMilli(),
		UpdatedAt:         item.UpdatedAt.UnixMilli(),
	}
}

func compatLegacySubscribeDiscounts(raw *string) []compatSubscribeDiscount {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return []compatSubscribeDiscount{}
	}
	type legacyDiscount struct {
		Quantity int64   `json:"quantity"`
		Discount float64 `json:"discount"`
	}
	var items []legacyDiscount
	if err := json.Unmarshal([]byte(*raw), &items); err != nil {
		return []compatSubscribeDiscount{}
	}
	result := make([]compatSubscribeDiscount, 0, len(items))
	for _, item := range items {
		result = append(result, compatSubscribeDiscount{
			Quantity: item.Quantity,
			Discount: item.Discount,
		})
	}
	return result
}

func compatLegacyTrafficLimits(raw *string) []compatTrafficLimit {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return []compatTrafficLimit{}
	}
	var items []compatTrafficLimit
	if err := json.Unmarshal([]byte(*raw), &items); err != nil {
		return []compatTrafficLimit{}
	}
	return items
}

func compatLegacyUserInfo(ctx context.Context, dataLayer *data.Data) (*compatUserInfo, error) {
	user, err := compatCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	methods, err := dataLayer.DB().ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserIDEQ(user.ID)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	devices, err := dataLayer.DB().ProxyUserDevice.Query().
		Where(proxyuserdevice.UserIDEQ(user.ID)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	authMethods := make([]compatUserAuthMethod, 0, len(methods))
	for _, item := range methods {
		identifier := item.AuthIdentifier
		switch strings.ToLower(strings.TrimSpace(item.AuthType)) {
		case "mobile":
			identifier = phone.MaskPhoneNumber(identifier)
		case "email":
		default:
			identifier = compatMaskOpenID(identifier)
		}
		authMethods = append(authMethods, compatUserAuthMethod{
			AuthType:       item.AuthType,
			AuthIdentifier: identifier,
			Verified:       item.Verified,
		})
	}
	sort.Slice(authMethods, func(i, j int) bool {
		return compatAuthTypePriority(authMethods[i].AuthType) < compatAuthTypePriority(authMethods[j].AuthType)
	})

	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })
	userDevices := make([]compatUserDevice, 0, len(devices))
	for _, item := range devices {
		userDevices = append(userDevices, compatUserDevice{
			ID:         item.ID,
			IP:         compatStringPointer(item.IP),
			Identifier: compatStringPointer(item.Identifier),
			UserAgent:  compatStringPointer(item.UserAgent),
			Online:     item.Online,
			Enabled:    item.Enabled,
			CreatedAt:  item.CreatedAt.UnixMilli(),
			UpdatedAt:  item.UpdatedAt.UnixMilli(),
		})
	}

	return &compatUserInfo{
		ID:                    user.ID,
		Avatar:                compatStringPointer(user.Avatar),
		Balance:               compatInt64Pointer(user.Balance),
		Commission:            compatInt64Pointer(user.Commission),
		ReferralPercentage:    uint8(user.ReferralPercentage),
		OnlyFirstPurchase:     user.OnlyFirstPurchase,
		GiftAmount:            compatInt64Pointer(user.GiftAmount),
		Telegram:              compatInt64Pointer(user.Telegram),
		ReferCode:             compatStringPointer(user.ReferCode),
		RefererID:             compatInt64Pointer(user.RefererID),
		Enable:                user.Enable,
		IsAdmin:               user.IsAdmin,
		EnableBalanceNotify:   user.EnableBalanceNotify,
		EnableLoginNotify:     user.EnableLoginNotify,
		EnableSubscribeNotify: user.EnableSubscribeNotify,
		EnableTradeNotify:     user.EnableTradeNotify,
		AuthMethods:           authMethods,
		UserDevices:           userDevices,
		Rules:                 compatJSONStringArray(user.Rules),
		CreatedAt:             user.CreatedAt.UnixMilli(),
		UpdatedAt:             user.UpdatedAt.UnixMilli(),
		DeletedAt:             compatUnixMillis(user.DeletedAt),
		IsDel:                 user.IsDel != nil && *user.IsDel != 0,
	}, nil
}

func compatLegacyUserSubscribe(ctx context.Context, dataLayer *data.Data) (*compatUserSubscribeListData, error) {
	user, err := compatCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	items, err := dataLayer.DB().ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.UserIDEQ(user.ID),
			proxyusersubscribe.StatusIn(0, 1, 2, 3),
		).
		Order(ent.Desc(proxyusersubscribe.FieldID)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	items = compatFilterLegacyUserSubscribes(items, time.Now())

	result := &compatUserSubscribeListData{List: []compatUserSubscribe{}}
	for _, item := range items {
		subscribePlan, err := dataLayer.DB().ProxySubscribe.Get(ctx, item.SubscribeID)
		if err != nil {
			return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
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

func compatLegacyLoginLog(data *publicuserv1.LoginLogData) *compatLoginLogListData {
	result := &compatLoginLogListData{List: []compatLoginLog{}}
	if data == nil {
		return result
	}
	result.Total = data.Total
	for _, item := range data.List {
		result.List = append(result.List, compatLoginLog{
			ID:        item.Id,
			UserID:    item.UserId,
			LoginIP:   item.LoginIp,
			UserAgent: item.UserAgent,
			Success:   item.Success,
			Timestamp: item.Timestamp,
		})
	}
	return result
}

func compatLegacyBalanceLog(data *publicuserv1.BalanceLogData) *compatBalanceLogListData {
	result := &compatBalanceLogListData{List: []compatBalanceLog{}}
	if data == nil {
		return result
	}
	result.Total = data.Total
	for _, item := range data.List {
		result.List = append(result.List, compatBalanceLog{
			Type:      uint16(item.Type),
			UserID:    item.UserId,
			Amount:    item.Amount,
			OrderNo:   item.OrderNo,
			Balance:   item.Balance,
			Timestamp: item.Timestamp,
		})
	}
	return result
}

func compatLegacyCommissionLog(data *publicuserv1.CommissionLogData) *compatCommissionLogListData {
	result := &compatCommissionLogListData{List: []compatCommissionLog{}}
	if data == nil {
		return result
	}
	result.Total = data.Total
	for _, item := range data.List {
		result.List = append(result.List, compatCommissionLog{
			Type:      uint16(item.Type),
			UserID:    item.UserId,
			Amount:    item.Amount,
			OrderNo:   item.OrderNo,
			Timestamp: item.Timestamp,
		})
	}
	return result
}

func compatLegacyAffiliateCount(data *publicuserv1.UserAffiliateData) *compatAffiliateCountData {
	if data == nil {
		return &compatAffiliateCountData{}
	}
	return &compatAffiliateCountData{
		Registers:       data.Registers,
		TotalCommission: data.TotalCommission,
	}
}

func compatLegacyAffiliateList(data *publicuserv1.UserAffiliateListData) *compatAffiliateListData {
	result := &compatAffiliateListData{List: []compatUserAffiliate{}}
	if data == nil {
		return result
	}
	result.Total = data.Total
	for _, item := range data.List {
		result.List = append(result.List, compatUserAffiliate{
			Avatar:       item.Avatar,
			Identifier:   item.Identifier,
			RegisteredAt: item.RegisteredAt,
			Enable:       item.Enable,
		})
	}
	return result
}

func compatLegacyOAuthMethods(data *publicuserv1.OAuthMethodsData) *compatOAuthMethodsData {
	result := &compatOAuthMethodsData{Methods: []compatUserAuthMethod{}}
	if data == nil {
		return result
	}
	for _, item := range data.Methods {
		result.Methods = append(result.Methods, compatUserAuthMethod{
			AuthType:       item.AuthType,
			AuthIdentifier: item.AuthIdentifier,
			Verified:       item.Verified,
		})
	}
	return result
}

func compatLegacySubscribeLog(data *publicuserv1.SubscribeLogData) *compatUserSubscribeLogListData {
	result := &compatUserSubscribeLogListData{List: []compatUserSubscribeLog{}}
	if data == nil {
		return result
	}
	result.Total = data.Total
	for _, item := range data.List {
		result.List = append(result.List, compatUserSubscribeLog{
			ID:              item.Id,
			UserID:          item.UserId,
			UserSubscribeID: item.UserSubscribeId,
			Token:           item.Token,
			IP:              item.Ip,
			UserAgent:       item.UserAgent,
			Timestamp:       item.Timestamp,
		})
	}
	return result
}

func compatLegacyPreUnsubscribe(data *publicuserv1.UnsubscribeInfoData) *compatPreUnsubscribeData {
	if data == nil {
		return &compatPreUnsubscribeData{}
	}
	return &compatPreUnsubscribeData{DeductionAmount: data.DeductionAmount}
}

func compatLegacyBindTelegram(data *publicuserv1.TelegramBindData) *compatBindTelegramData {
	if data == nil {
		return &compatBindTelegramData{}
	}
	return &compatBindTelegramData{
		URL:       data.Url,
		ExpiredAt: data.ExpiredAt,
	}
}

func compatLegacyBindOAuth(data *publicuserv1.OAuthBindData) *compatBindOAuthData {
	if data == nil {
		return &compatBindOAuthData{}
	}
	return &compatBindOAuthData{Redirect: data.Redirect}
}

func compatLegacyDeviceList(data *publicuserv1.GetDeviceListData) *compatUserDeviceListData {
	result := &compatUserDeviceListData{List: []compatUserDevice{}}
	if data == nil {
		return result
	}
	result.Total = data.Total
	for _, item := range data.List {
		result.List = append(result.List, compatUserDevice{
			ID:         item.Id,
			IP:         item.Ip,
			Identifier: item.Identifier,
			UserAgent:  item.UserAgent,
			Online:     item.Online,
			Enabled:    item.Enabled,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}
	return result
}

func compatLegacyDeviceOnlineStatistics(data *publicuserv1.GetDeviceOnlineStatisticsData) *compatDeviceOnlineStatisticsData {
	result := &compatDeviceOnlineStatisticsData{WeeklyStats: []compatWeeklyStat{}}
	if data == nil {
		return result
	}
	for _, item := range data.WeeklyStats {
		result.WeeklyStats = append(result.WeeklyStats, compatWeeklyStat{
			Day:     item.Day,
			DayName: item.DayName,
			Hours:   item.Hours,
		})
	}
	if data.ConnectionRecords != nil {
		result.ConnectionRecords = compatConnectionRecords{
			CurrentContinuousDays:   data.ConnectionRecords.CurrentContinuousDays,
			HistoryContinuousDays:   data.ConnectionRecords.HistoryContinuousDays,
			LongestSingleConnection: data.ConnectionRecords.LongestSingleConnection,
		}
	}
	return result
}

func compatLegacyCreatedWithdrawal(data *publicuserv1.WithdrawalLogData) *compatWithdrawalLog {
	if data == nil {
		return &compatWithdrawalLog{}
	}
	return &compatWithdrawalLog{
		ID:        0,
		UserID:    compatParseInt64String(data.UserId),
		Amount:    compatParseInt64String(data.Amount),
		Content:   data.Content,
		Status:    uint8(data.Status),
		Reason:    data.Reason,
		CreatedAt: compatParseInt64String(data.CreatedAt),
		UpdatedAt: 0,
	}
}

func compatLegacyWithdrawalLog(data *publicuserv1.WithdrawalLogListData) *compatWithdrawalLogListData {
	result := &compatWithdrawalLogListData{List: []compatWithdrawalLog{}}
	if data == nil {
		return result
	}
	result.Total = compatParseInt64String(data.Total)
	for _, item := range data.List {
		result.List = append(result.List, compatWithdrawalLog{
			ID:        compatParseInt64String(item.Id),
			UserID:    compatParseInt64String(item.UserId),
			Amount:    compatParseInt64String(item.Amount),
			Content:   item.Content,
			Status:    uint8(item.Status),
			Reason:    item.Reason,
			CreatedAt: compatParseInt64String(item.CreatedAt),
			UpdatedAt: compatParseInt64String(item.UpdatedAt),
		})
	}
	return result
}

func compatLegacyDeviceWSConnect(ctx context.Context, httpCtx khttp.Context, dataLayer *data.Data) error {
	user, err := compatCurrentUser(ctx)
	if err != nil {
		return err
	}

	identifier := ""
	if raw, ok := ctx.Value(constant.CtxKeyIdentifier).(string); ok {
		identifier = strings.TrimSpace(raw)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(httpCtx.Query().Get("identifier"))
	}
	if identifier == "" {
		return compatCodeError(400, "identifier is empty")
	}

	deviceInfo, err := dataLayer.DB().ProxyUserDevice.Query().
		Where(proxyuserdevice.IdentifierEQ(identifier)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if ent.IsNotFound(err) {
		if _, err := dataLayer.DB().ProxyUserDevice.Create().
			SetIdentifier(identifier).
			SetUserID(user.ID).
			SetOnline(true).
			SetEnabled(true).
			Save(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}
		deviceInfo = nil
	}
	if deviceInfo != nil && deviceInfo.UserID != user.ID {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	maxDevice := 3
	subs, err := dataLayer.DB().ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.UserIDEQ(user.ID),
			proxyusersubscribe.StatusIn(1, 2),
		).
		All(ctx)
	if err == nil {
		for _, item := range subs {
			if item.ExpireTime == nil || !time.Now().Before(*item.ExpireTime) {
				continue
			}
			subscribePlan, subErr := dataLayer.DB().ProxySubscribe.Get(ctx, item.SubscribeID)
			if subErr != nil {
				continue
			}
			if int(subscribePlan.DeviceLimit) > maxDevice {
				maxDevice = int(subscribePlan.DeviceLimit)
			}
		}
	}

	if dataLayer.DeviceManager() == nil {
		return compatCodeError(responsecode.ErrInternalError, "device manager unavailable")
	}

	return dataLayer.DeviceManager().AddDevice(
		httpCtx.Response(),
		httpCtx.Request(),
		compatCurrentSessionID(ctx),
		user.ID,
		identifier,
		maxDevice,
	)
}

func compatAuthTypePriority(authType string) int {
	switch strings.ToLower(strings.TrimSpace(authType)) {
	case "email":
		return 1
	case "mobile":
		return 2
	default:
		return 100
	}
}

func compatMaskOpenID(openID string) string {
	if len(openID) <= 6 {
		return "***"
	}
	return openID[:3] + strings.Repeat("*", len(openID)-6) + openID[len(openID)-3:]
}

func compatJSONStringArray(raw *string) []string {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(*raw), &result); err != nil {
		return []string{}
	}
	return result
}

func compatStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func compatInt64Pointer(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func compatUnixMillis(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.UnixMilli()
}

func compatLegacyUnixMillis(value *time.Time) int64 {
	if value == nil || value.Unix() == 0 {
		return 0
	}
	return value.UnixMilli()
}

func compatLegacyUserResetTime(item compatUserSubscribe) int64 {
	resetTime := time.UnixMilli(item.ExpireTime)
	now := time.Now()
	switch item.Subscribe.ResetCycle {
	case 0:
		return 0
	case 1:
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UnixMilli()
	case 2:
		if resetTime.Day() > now.Day() {
			return time.Date(now.Year(), now.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		}
		return time.Date(now.Year(), now.Month()+1, resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	case 3:
		target := time.Date(now.Year(), resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		if target.Before(now) {
			target = time.Date(now.Year()+1, resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		}
		return target.UnixMilli()
	default:
		return 0
	}
}
