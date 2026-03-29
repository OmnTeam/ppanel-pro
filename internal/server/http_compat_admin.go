package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	adminticketv1 "github.com/OmnTeam/ppanel-pro/api/admin/ticket/v1"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	adminticketservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/ticket"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatUpdateTicketStatusRequest struct {
	ID     int64  `json:"id"`
	Status *uint8 `json:"status"`
}

func registerLegacyAdminCompatRoutes(r *khttp.Router, dataLayer *data.Data, adminTicket *adminticketservice.TicketService) {
	if r == nil {
		return
	}

	r.GET("/v1/admin/system/module", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			version := "unknown version"
			if value, loadErr := compatSystemValue(inner, dataLayer, "system", "Version", "version"); loadErr == nil && strings.TrimSpace(value) != "" {
				version = value
			}
			return map[string]string{
				"service_name":    "ApiService",
				"service_version": strings.ReplaceAll(version, "v", ""),
				"secret":          "",
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.PUT("/v1/admin/ticket", func(ctx khttp.Context) error {
		if adminTicket == nil {
			return compatJSONError(ctx, compatCodeError(500, "ticket service unavailable"))
		}

		var req compatUpdateTicketStatusRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)

		if req.ID == 0 {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateTicketStatusRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"))
		}
		if req.Status == nil {
			return compatJSONError(ctx, compatParamError("Key: 'UpdateTicketStatusRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"))
		}

		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatUpdateTicketStatusRequest)
			return adminTicket.UpdateTicketStatus(inner, &adminticketv1.UpdateTicketStatusRequest{
				Id:     strconv.FormatInt(in.ID, 10),
				Status: int32(*in.Status),
			})
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, nil)
	})

	r.POST("/v1/admin/subscribe/reset_all_token", func(ctx khttp.Context) error {
		if dataLayer == nil || dataLayer.DB() == nil {
			return compatJSONError(ctx, compatCodeError(500, "data layer unavailable"))
		}

		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			tx, err := dataLayer.DB().Tx(inner)
			if err != nil {
				return nil, compatCodeError(500, "failed to begin transaction")
			}

			userSubs, err := tx.ProxyUserSubscribe.Query().
				Where(proxyusersubscribe.StatusIn(1, 2)).
				All(inner)
			if err != nil {
				_ = tx.Rollback()
				return nil, compatCodeError(500, "Failed to fetch subscribe list")
			}

			nowMillis := time.Now().UnixMilli()
			for _, userSub := range userSubs {
				token := uuidx.SubscribeToken(fmt.Sprintf("%d%d", nowMillis, userSub.ID))
				subscribeUUID := uuidx.NewUUID().String()
				if _, err := tx.ProxyUserSubscribe.UpdateOneID(userSub.ID).
					SetToken(token).
					SetUUID(subscribeUUID).
					Save(inner); err != nil {
					_ = tx.Rollback()
					return nil, compatCodeError(500, fmt.Sprintf("Failed to update subscribe token for ID %d", userSub.ID))
				}
			}

			if err := tx.Commit(); err != nil {
				return nil, compatCodeError(500, "Failed to commit transaction")
			}

			for _, userSub := range userSubs {
				compatClearUserSubscribeCaches(inner, dataLayer.Redis(), userSub)
				compatClearSubscribeCaches(inner, dataLayer.Redis(), userSub.SubscribeID)
			}

			return map[string]bool{"success": true}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})
}
