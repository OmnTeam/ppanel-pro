package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyannouncement"
	"github.com/OmnTeam/ppanel-pro/ent/proxydocument"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyserver"
	"github.com/OmnTeam/ppanel-pro/ent/proxyservergroup"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdeviceonlinerecord"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatUserSubscribeNodeInfo struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Uuid            string   `json:"uuid"`
	Protocol        string   `json:"protocol"`
	Protocols       string   `json:"protocols"`
	Port            uint16   `json:"port"`
	Address         string   `json:"address"`
	Tags            []string `json:"tags"`
	Country         string   `json:"country"`
	City            string   `json:"city"`
	Longitude       string   `json:"longitude"`
	Latitude        string   `json:"latitude"`
	LatitudeCenter  string   `json:"latitude_center"`
	LongitudeCenter string   `json:"longitude_center"`
	CreatedAt       int64    `json:"created_at"`
}

type compatUserSubscribeInfo struct {
	ID          int64                          `json:"id"`
	UserID      int64                          `json:"user_id"`
	OrderID     int64                          `json:"order_id"`
	SubscribeID int64                          `json:"subscribe_id"`
	StartTime   int64                          `json:"start_time"`
	ExpireTime  int64                          `json:"expire_time"`
	FinishedAt  int64                          `json:"finished_at"`
	ResetTime   int64                          `json:"reset_time"`
	Traffic     int64                          `json:"traffic"`
	Download    int64                          `json:"download"`
	Upload      int64                          `json:"upload"`
	Token       string                         `json:"token"`
	Status      uint8                          `json:"status"`
	CreatedAt   int64                          `json:"created_at"`
	UpdatedAt   int64                          `json:"updated_at"`
	IsTryOut    bool                           `json:"is_try_out"`
	Nodes       []*compatUserSubscribeNodeInfo `json:"nodes"`
}

type compatAnnouncementRequest struct {
	Page   int   `json:"page"`
	Size   int   `json:"size"`
	Pinned *bool `json:"pinned"`
	Popup  *bool `json:"popup"`
}

type compatAnnouncement struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Show      *bool  `json:"show"`
	Pinned    *bool  `json:"pinned"`
	Popup     *bool  `json:"popup"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type compatAnnouncementListData struct {
	Total int64                 `json:"total"`
	List  []*compatAnnouncement `json:"announcements"`
}

type compatDocumentDetailRequest struct {
	ID int64 `json:"id"`
}

type compatDocument struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags"`
	Show      bool     `json:"show"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

type compatDocumentListData struct {
	Total int64             `json:"total"`
	List  []*compatDocument `json:"list"`
}

func registerLegacyPublicCompatRoutes(r *khttp.Router, dataLayer *data.Data, appConf *conf.Application, publicOrder legacyPublicOrderCompat, publicPayment legacyPublicPaymentCompat, publicPortal legacyPublicPortalCompat, publicTicket legacyPublicTicketCompat, publicUser legacyPublicUserCompat) {
	if r == nil {
		return
	}

	registerLegacyPublicSubscribeCompatRoutes(r, dataLayer)
	registerLegacyPublicTicketCompatRoutes(r, publicTicket)
	registerLegacyPublicUserCompatRoutes(r, dataLayer, publicUser)
	registerLegacyPublicOrderCompatRoutes(r, dataLayer, publicOrder)
	registerLegacyPublicPaymentCompatRoutes(r, publicPayment)
	registerLegacyPublicPortalCompatRoutes(r, dataLayer, publicPortal)

	r.GET("/v1/public/announcement/list", func(ctx khttp.Context) error {
		req := compatAnnouncementRequest{}
		if rawPage := strings.TrimSpace(ctx.Query().Get("page")); rawPage != "" {
			if parsed, err := strconv.Atoi(rawPage); err == nil {
				req.Page = parsed
			}
		}
		if rawSize := strings.TrimSpace(ctx.Query().Get("size")); rawSize == "" {
			req.Size = 15
		} else if parsed, err := strconv.Atoi(rawSize); err == nil {
			req.Size = parsed
		}
		if rawPinned := strings.TrimSpace(ctx.Query().Get("pinned")); rawPinned != "" {
			if parsed, err := strconv.ParseBool(rawPinned); err == nil {
				req.Pinned = compatBoolPointer(parsed)
			}
		}
		if rawPopup := strings.TrimSpace(ctx.Query().Get("popup")); rawPopup != "" {
			if parsed, err := strconv.ParseBool(rawPopup); err == nil {
				req.Popup = compatBoolPointer(parsed)
			}
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatAnnouncementRequest)

			query := dataLayer.DB().ProxyAnnouncement.Query().
				Where(proxyannouncement.ShowEQ(true))
			if in.Pinned != nil {
				query = query.Where(proxyannouncement.PinnedEQ(*in.Pinned))
			}
			if in.Popup != nil {
				query = query.Where(proxyannouncement.PopupEQ(*in.Popup))
			}

			total, err := query.Count(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			size := in.Size
			if size == 0 {
				size = 10
			}
			announcements, err := query.
				Offset((in.Page - 1) * size).
				Limit(size).
				All(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			list := make([]*compatAnnouncement, 0, len(announcements))
			for _, item := range announcements {
				list = append(list, &compatAnnouncement{
					ID:        item.ID,
					Title:     item.Title,
					Content:   item.Content,
					Show:      compatBoolPointer(item.Show),
					Pinned:    compatBoolPointer(item.Pinned),
					Popup:     compatBoolPointer(item.Popup),
					CreatedAt: item.CreatedAt.UnixMilli(),
					UpdatedAt: item.UpdatedAt.UnixMilli(),
				})
			}

			return &compatAnnouncementListData{
				Total: int64(total),
				List:  list,
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/document/list", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			query := dataLayer.DB().ProxyDocument.Query().
				Where(proxydocument.ShowEQ(true))

			total, err := query.Count(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			documents, err := query.All(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			list := make([]*compatDocument, 0, len(documents))
			for _, item := range documents {
				list = append(list, &compatDocument{
					ID:        item.ID,
					Title:     item.Title,
					Content:   "",
					Tags:      compatSplitAndUniqueCSV(item.Tags),
					Show:      false,
					CreatedAt: 0,
					UpdatedAt: item.UpdatedAt.UnixMilli(),
				})
			}

			return &compatDocumentListData{
				Total: int64(total),
				List:  list,
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/document/detail", func(ctx khttp.Context) error {
		req := compatDocumentDetailRequest{}
		if rawID := strings.TrimSpace(ctx.Query().Get("id")); rawID != "" {
			if parsed, err := strconv.ParseInt(rawID, 10, 64); err == nil {
				req.ID = parsed
			}
		}
		if req.ID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'QueryDocumentDetailRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"))
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatDocumentDetailRequest)
			document, err := dataLayer.DB().ProxyDocument.Query().
				Where(proxydocument.IDEQ(in.ID)).
				Only(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			return &compatDocument{
				ID:        document.ID,
				Title:     document.Title,
				Content:   document.Content,
				Tags:      compatSplitAndUniqueCSV(document.Tags),
				Show:      document.Show,
				CreatedAt: document.CreatedAt.UnixMilli(),
				UpdatedAt: document.UpdatedAt.UnixMilli(),
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/subscribe/node/list", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			user, err := compatCurrentUser(inner)
			if err != nil {
				return nil, err
			}

			userSubs, err := dataLayer.DB().ProxyUserSubscribe.Query().
				Where(
					proxyusersubscribe.UserIDEQ(user.ID),
					proxyusersubscribe.StatusIn(0, 1, 2, 3),
				).
				Order(ent.Asc(proxyusersubscribe.FieldID)).
				All(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			userSubs = compatFilterLegacyUserSubscribes(userSubs, time.Now())
			list := make([]*compatUserSubscribeInfo, 0, len(userSubs))
			groupEnabled := compatLegacyGroupEnabled(inner, dataLayer)

			for _, userSub := range userSubs {
				subscribePlan, subErr := dataLayer.DB().ProxySubscribe.Query().
					Where(proxysubscribe.IDEQ(userSub.SubscribeID)).
					Only(inner)
				if subErr != nil {
					return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
				}

				nodes, nodeErr := compatLegacyNodeList(inner, dataLayer, userSub, subscribePlan, groupEnabled)
				if nodeErr != nil {
					return nil, nodeErr
				}

				item := &compatUserSubscribeInfo{
					ID:          userSub.ID,
					UserID:      userSub.UserID,
					OrderID:     userSub.OrderID,
					SubscribeID: userSub.SubscribeID,
					StartTime:   userSub.StartTime.Unix(),
					ExpireTime:  compatLegacyUnix(userSub.ExpireTime),
					FinishedAt:  compatUnix(userSub.FinishedAt),
					ResetTime:   0,
					Traffic:     compatInt64Value(userSub.Traffic),
					Download:    compatInt64Value(userSub.Download),
					Upload:      compatInt64Value(userSub.Upload),
					Token:       compatStringValue(userSub.Token),
					Status:      uint8(compatInt8Value(userSub.Status)),
					CreatedAt:   userSub.CreatedAt.Unix(),
					UpdatedAt:   userSub.UpdatedAt.Unix(),
					IsTryOut:    appConf != nil && appConf.Register != nil && appConf.Register.EnableTrial && appConf.Register.TrialSubscribe == userSub.SubscribeID,
					Nodes:       nodes,
				}
				list = append(list, item)
			}

			return map[string]interface{}{"list": list}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/public/user/traffic_stats", func(ctx khttp.Context) error {
		rawSubID := strings.TrimSpace(ctx.Query().Get("user_subscribe_id"))
		rawDays := strings.TrimSpace(ctx.Query().Get("days"))
		if rawSubID == "" {
			return compatJSONError(ctx, compatParamError("Key: 'GetUserTrafficStatsRequest.UserSubscribeId' Error:Field validation for 'UserSubscribeId' failed on the 'required' tag"))
		}
		if rawDays == "" {
			return compatJSONError(ctx, compatParamError("Key: 'GetUserTrafficStatsRequest.Days' Error:Field validation for 'Days' failed on the 'required' tag"))
		}
		days, err := strconv.Atoi(rawDays)
		if err != nil || (days != 7 && days != 30) {
			return compatJSONError(ctx, compatParamError("Key: 'GetUserTrafficStatsRequest.Days' Error:Field validation for 'Days' failed on the 'oneof' tag"))
		}

		out, callErr := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			user, err := compatCurrentUser(inner)
			if err != nil {
				return nil, err
			}

			userSubscribeID, err := strconv.ParseInt(rawSubID, 10, 64)
			if err != nil {
				return nil, compatCodeError(responsecode.ErrInvalidAccess, "Invalid subscription ID")
			}

			userSub, err := dataLayer.DB().ProxyUserSubscribe.Query().
				Where(proxyusersubscribe.IDEQ(userSubscribeID)).
				Only(inner)
			if err != nil {
				if ent.IsNotFound(err) {
					return nil, compatCodeError(responsecode.ErrInvalidAccess, "Subscription not found")
				}
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}
			if userSub.UserID != user.ID {
				return nil, compatCodeError(responsecode.ErrInvalidAccess, "Invalid Access")
			}

			now := time.Now()
			startDate := now.AddDate(0, 0, -days+1)
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
			endDate := startDate.AddDate(0, 0, days).Add(-time.Nanosecond)

			logs, err := dataLayer.DB().ProxyTrafficLog.Query().
				Where(
					proxytrafficlog.UserIDEQ(user.ID),
					proxytrafficlog.SubscribeIDEQ(userSubscribeID),
					proxytrafficlog.TimestampGTE(startDate),
					proxytrafficlog.TimestampLTE(endDate),
				).
				All(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}

			trafficByDay := make(map[string]compatDailyTrafficStats)
			for _, item := range logs {
				dateKey := item.Timestamp.In(time.Local).Format("2006-01-02")
				stat := trafficByDay[dateKey]
				stat.Date = dateKey
				stat.Upload += item.Upload
				stat.Download += item.Download
				stat.Total = stat.Upload + stat.Download
				trafficByDay[dateKey] = stat
			}

			resp := &compatTrafficStatsData{
				List: make([]compatDailyTrafficStats, 0, days),
			}
			for i := 0; i < days; i++ {
				currentDate := startDate.AddDate(0, 0, i)
				dateKey := currentDate.Format("2006-01-02")
				stat, ok := trafficByDay[dateKey]
				if !ok {
					stat = compatDailyTrafficStats{Date: dateKey}
				}
				resp.List = append(resp.List, stat)
				resp.TotalUpload += stat.Upload
				resp.TotalDownload += stat.Download
			}
			resp.TotalTraffic = resp.TotalUpload + resp.TotalDownload

			return resp, nil
		})
		if callErr != nil {
			return compatJSONError(ctx, callErr)
		}
		return compatJSON(ctx, out)
	})

	r.PUT("/v1/public/user/rules", func(ctx khttp.Context) error {
		var req compatUpdateUserRulesRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if len(req.Rules) == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateUserRulesRequest.Rules' Error:Field validation for 'Rules' failed on the 'required' tag"))
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			user, err := compatCurrentUser(inner)
			if err != nil {
				return nil, err
			}
			payload, err := json.Marshal(request.(*compatUpdateUserRulesRequest).Rules)
			if err != nil {
				return nil, compatCodeError(responsecode.ErrInternalError, err.Error())
			}
			if _, err := dataLayer.DB().ProxyUser.UpdateOneID(user.ID).SetRules(string(payload)).Save(inner); err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
			}
			return nil, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.PUT("/v1/public/user/subscribe_note", func(ctx khttp.Context) error {
		var req compatUpdateUserSubscribeNoteRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if req.UserSubscribeID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateUserSubscribeNoteRequest.UserSubscribeId' Error:Field validation for 'UserSubscribeId' failed on the 'required' tag"))
		}
		if len(req.Note) > 500 {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateUserSubscribeNoteRequest.Note' Error:Field validation for 'Note' failed on the 'max' tag"))
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			user, err := compatCurrentUser(inner)
			if err != nil {
				return nil, err
			}
			in := request.(*compatUpdateUserSubscribeNoteRequest)
			userSub, err := dataLayer.DB().ProxyUserSubscribe.Query().
				Where(proxyusersubscribe.IDEQ(int64(in.UserSubscribeID))).
				Only(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
			}
			if userSub.UserID != user.ID {
				return nil, compatCodeError(responsecode.ErrInvalidAccess, "UserSubscribeId does not belong to the current user")
			}
			if _, err := dataLayer.DB().ProxyUserSubscribe.UpdateOneID(userSub.ID).SetNote(in.Note).Save(inner); err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
			}
			compatClearUserSubscribeCaches(inner, dataLayer.Redis(), userSub)
			compatClearSubscribeCaches(inner, dataLayer.Redis(), userSub.SubscribeID)
			return nil, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.DELETE("/v1/public/user/current_user_account", func(ctx khttp.Context) error {
		_, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			user, err := compatCurrentUser(inner)
			if err != nil {
				return nil, err
			}

			emails, _ := compatUserEmails(inner, dataLayer, user.ID)
			userSubs, _ := dataLayer.DB().ProxyUserSubscribe.Query().
				Where(proxyusersubscribe.UserIDEQ(user.ID)).
				All(inner)

			tx, err := dataLayer.DB().Tx(inner)
			if err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
			}
			defer func() {
				if recover() != nil {
					_ = tx.Rollback()
					panic("compat current_user_account panic")
				}
			}()

			if _, err := tx.ProxyUserDeviceOnlineRecord.Delete().Where(proxyuserdeviceonlinerecord.UserIDEQ(user.ID)).Exec(inner); err != nil {
				_ = tx.Rollback()
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}
			if _, err := tx.ProxyUserDevice.Delete().Where(proxyuserdevice.UserIDEQ(user.ID)).Exec(inner); err != nil {
				_ = tx.Rollback()
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}
			if _, err := tx.ProxyUserAuthMethod.Delete().Where(proxyuserauthmethod.UserIDEQ(user.ID)).Exec(inner); err != nil {
				_ = tx.Rollback()
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}
			if _, err := tx.ProxyUserSubscribe.Delete().Where(proxyusersubscribe.UserIDEQ(user.ID)).Exec(inner); err != nil {
				_ = tx.Rollback()
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}
			if _, err := tx.ProxyUser.Update().
				Where(proxyuser.IDEQ(user.ID)).
				SetDeletedAt(time.Now()).
				SetIsDel(0).
				Save(inner); err != nil {
				_ = tx.Rollback()
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}
			if err := tx.Commit(); err != nil {
				return nil, responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}

			for _, item := range userSubs {
				compatClearUserSubscribeCaches(inner, dataLayer.Redis(), item)
				compatClearSubscribeCaches(inner, dataLayer.Redis(), item.SubscribeID)
			}
			compatClearUserCache(inner, dataLayer.Redis(), user.ID, emails...)
			if sessionID := compatCurrentSessionID(inner); sessionID != "" {
				compatDeleteKeys(inner, dataLayer.Redis(), fmt.Sprintf("%s:%s", constant.SessionIdKey, sessionID))
			}
			return nil, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})
}

func compatFilterLegacyUserSubscribes(items []*ent.ProxyUserSubscribe, now time.Time) []*ent.ProxyUserSubscribe {
	filtered := make([]*ent.ProxyUserSubscribe, 0, len(items))
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	for _, item := range items {
		if compatIsLegacyUnlimitedTime(item.ExpireTime) {
			filtered = append(filtered, item)
			continue
		}
		if item.ExpireTime != nil && item.ExpireTime.After(now) {
			filtered = append(filtered, item)
			continue
		}
		if item.FinishedAt != nil && (item.FinishedAt.After(sevenDaysAgo) || item.FinishedAt.Equal(sevenDaysAgo)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func compatIsLegacyUnlimitedTime(value *time.Time) bool {
	return value != nil && value.Unix() == 0
}

func compatLegacyUnix(value *time.Time) int64 {
	if compatIsLegacyUnlimitedTime(value) {
		return 0
	}
	return compatUnix(value)
}

func compatLegacyGroupEnabled(ctx context.Context, dataLayer *data.Data) bool {
	value, err := compatSystemValue(ctx, dataLayer, "group", "enabled")
	if err != nil {
		return false
	}
	return value == "true" || value == "1"
}

func compatLegacyNodeList(ctx context.Context, dataLayer *data.Data, userSub *ent.ProxyUserSubscribe, subscribePlan *ent.ProxySubscribe, groupEnabled bool) ([]*compatUserSubscribeNodeInfo, error) {
	now := time.Now()
	if !compatIsLegacyUnlimitedTime(userSub.ExpireTime) && userSub.ExpireTime != nil && userSub.ExpireTime.Before(now) {
		return compatLegacyExpiredNodes(ctx, dataLayer, userSub)
	}

	enabledNodes, err := dataLayer.DB().ProxyNode.Query().
		Where(proxynode.EnabledEQ(true)).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	selected := make([]*ent.ProxyNode, 0)
	seen := make(map[int64]struct{})

	if groupEnabled {
		nodeGroupID := compatResolveNodeGroupID(userSub, subscribePlan)
		directNodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)

		for _, node := range enabledNodes {
			if len(node.NodeGroupIds) == 0 {
				if _, ok := seen[node.ID]; !ok {
					seen[node.ID] = struct{}{}
					selected = append(selected, node)
				}
				continue
			}
			if nodeGroupID != 0 && tool.Contains(node.NodeGroupIds, nodeGroupID) {
				if _, ok := seen[node.ID]; !ok {
					seen[node.ID] = struct{}{}
					selected = append(selected, node)
				}
			}
		}

		for _, node := range enabledNodes {
			if tool.Contains(directNodeIDs, node.ID) {
				if _, ok := seen[node.ID]; !ok {
					seen[node.ID] = struct{}{}
					selected = append(selected, node)
				}
			}
		}
	} else {
		nodeIDs := tool.StringToInt64Slice(subscribePlan.Nodes)
		tags := make([]string, 0)
		for _, item := range strings.Split(subscribePlan.NodeTags, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}

		for _, node := range enabledNodes {
			if len(nodeIDs) > 0 && !tool.Contains(nodeIDs, node.ID) {
				continue
			}
			if len(tags) > 0 && !compatNodeMatchesTags(node.Tags, tags) {
				continue
			}
			if len(nodeIDs) == 0 && len(tags) == 0 {
				continue
			}
			selected = append(selected, node)
		}
	}

	return compatBuildLegacyNodeInfos(ctx, dataLayer, userSub, selected)
}

func compatLegacyExpiredNodes(ctx context.Context, dataLayer *data.Data, userSub *ent.ProxyUserSubscribe) ([]*compatUserSubscribeNodeInfo, error) {
	expiredGroup, err := dataLayer.DB().ProxyServerGroup.Query().
		Where(proxyservergroup.IsExpiredGroupEQ(true)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if userSub.ExpireTime == nil {
		return nil, nil
	}

	expiredDays := int(time.Since(*userSub.ExpireTime).Hours() / 24)
	if expiredDays > expiredGroup.ExpiredDaysLimit {
		return nil, nil
	}
	if expiredGroup.MaxTrafficGBExpired != nil && *expiredGroup.MaxTrafficGBExpired > 0 {
		usedTrafficGB := float64(compatInt64Value(userSub.ExpiredDownload)+compatInt64Value(userSub.ExpiredUpload)) / (1024 * 1024 * 1024)
		if usedTrafficGB >= float64(*expiredGroup.MaxTrafficGBExpired) {
			return nil, nil
		}
	}

	nodes, err := dataLayer.DB().ProxyNode.Query().
		Where(proxynode.EnabledEQ(true)).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	selected := make([]*ent.ProxyNode, 0)
	for _, node := range nodes {
		if tool.Contains(node.NodeGroupIds, expiredGroup.ID) {
			selected = append(selected, node)
		}
	}
	return compatBuildLegacyNodeInfos(ctx, dataLayer, userSub, selected)
}

func compatResolveNodeGroupID(userSub *ent.ProxyUserSubscribe, subscribePlan *ent.ProxySubscribe) int64 {
	if userSub != nil && userSub.NodeGroupID != 0 {
		return userSub.NodeGroupID
	}
	if subscribePlan != nil && subscribePlan.NodeGroupID != nil && *subscribePlan.NodeGroupID != 0 {
		return *subscribePlan.NodeGroupID
	}
	if subscribePlan != nil && len(subscribePlan.NodeGroupIds) > 0 {
		return subscribePlan.NodeGroupIds[0]
	}
	return 0
}

func compatNodeMatchesTags(nodeTags string, tags []string) bool {
	for _, tag := range tags {
		for _, item := range strings.Split(nodeTags, ",") {
			if strings.TrimSpace(item) == tag {
				return true
			}
		}
	}
	return false
}

func compatBuildLegacyNodeInfos(ctx context.Context, dataLayer *data.Data, userSub *ent.ProxyUserSubscribe, nodes []*ent.ProxyNode) ([]*compatUserSubscribeNodeInfo, error) {
	if len(nodes) == 0 {
		return []*compatUserSubscribeNodeInfo{}, nil
	}

	serverIDs := make([]int64, 0, len(nodes))
	serverSeen := make(map[int64]struct{}, len(nodes))
	for _, node := range nodes {
		if _, ok := serverSeen[node.ServerID]; ok {
			continue
		}
		serverSeen[node.ServerID] = struct{}{}
		serverIDs = append(serverIDs, node.ServerID)
	}

	servers, err := dataLayer.DB().ProxyServer.Query().
		Where(proxyserver.IDIn(serverIDs...)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	serverMap := make(map[int64]*ent.ProxyServer, len(servers))
	for _, server := range servers {
		serverMap[server.ID] = server
	}

	result := make([]*compatUserSubscribeNodeInfo, 0, len(nodes))
	for _, node := range nodes {
		server := serverMap[node.ServerID]
		if server == nil {
			continue
		}
		result = append(result, &compatUserSubscribeNodeInfo{
			ID:              node.ID,
			Name:            node.Name,
			Uuid:            compatStringValue(userSub.UUID),
			Protocol:        node.Protocol,
			Protocols:       server.Protocol,
			Port:            node.Port,
			Address:         node.Address,
			Tags:            strings.Split(node.Tags, ","),
			Country:         server.Country,
			City:            server.City,
			Longitude:       server.Longitude,
			Latitude:        server.Latitude,
			LatitudeCenter:  server.LatitudeCenter,
			LongitudeCenter: server.LongitudeCenter,
			CreatedAt:       node.CreatedAt.Unix(),
		})
	}
	return result, nil
}

func compatBoolPointer(value bool) *bool {
	out := value
	return &out
}

func compatSplitAndUniqueCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return result
}
