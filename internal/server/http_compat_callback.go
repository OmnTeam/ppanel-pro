package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	publicpaymentv1 "github.com/OmnTeam/ppanel-pro/api/public/payment/v1"
	"github.com/OmnTeam/ppanel-pro/ent/proxypayment"
	telegramcallback "github.com/OmnTeam/ppanel-pro/internal/adapter/telegramcallback"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	publicpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/payment"
	"github.com/OmnTeam/ppanel-pro/pkg/payment"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func registerLegacyCallbackCompatRoutes(r *khttp.Router, dataLayer *data.Data, appConf *conf.Application, publicPayment *publicpaymentservice.PaymentService, logger log.Logger) {
	if r == nil {
		return
	}

	notifyHandler := compatNotifyHandler(dataLayer, publicPayment)
	for _, method := range []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	} {
		r.Handle(method, "/v1/notify/{platform}/{token}", notifyHandler)
	}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		r.Handle(method, "/v1/auth/oauth/callback/apple", compatAppleLoginCallbackHandler(dataLayer, appConf, logger))
	}

	r.POST("/v1/telegram/webhook", func(ctx khttp.Context) error {
		helper := log.NewHelper(logger)
		secret := strings.TrimSpace(ctx.Query().Get("secret"))

		botToken, err := telegramcallback.ResolveBotToken(ctx, dataLayer)
		if err != nil {
			helper.Errorf("[compatTelegramWebhook] load bot token failed: %v", err)
			return compatJSON(ctx, nil)
		}
		if secret == "" || tool.Md5Encode(botToken, false) != secret {
			helper.Errorf("[compatTelegramWebhook] secret mismatch")
			return compatJSON(ctx, nil)
		}

		var update tgbotapi.Update
		if err := json.NewDecoder(ctx.Request().Body).Decode(&update); err != nil {
			helper.Errorf("[compatTelegramWebhook] bind request failed: %v", err)
			return compatJSONError(ctx, err)
		}

		_, _ = compatMiddleware(ctx, &update, func(inner context.Context, req interface{}) (interface{}, error) {
			telegramcallback.HandleUpdate(inner, dataLayer, req.(*tgbotapi.Update), botToken, logger)
			return nil, nil
		})

		return compatJSON(ctx, nil)
	})
}

func compatNotifyHandler(dataLayer *data.Data, publicPayment *publicpaymentservice.PaymentService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		if publicPayment == nil || dataLayer == nil || dataLayer.DB() == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "payment service unavailable"})
		}

		var pathReq compatPathTokenRequest
		_ = ctx.BindVars(&pathReq)
		if pathReq.Platform == "" {
			pathReq.Platform = ctx.Vars().Get("platform")
		}
		if pathReq.Token == "" {
			pathReq.Token = ctx.Vars().Get("token")
		}

		config, err := dataLayer.DB().ProxyPayment.Query().
			Where(proxypayment.Token(pathReq.Token)).
			Only(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		switch payment.ParsePlatform(config.Platform) {
		case payment.EPay, payment.CryptoSaaS:
			req := compatBuildEPayNotifyRequest(ctx, pathReq.Token)
			out, err := compatMiddleware(ctx, req, func(inner context.Context, request interface{}) (interface{}, error) {
				return publicPayment.EPayNotify(inner, request.(*publicpaymentv1.EPayNotifyRequest))
			})
			if err != nil {
				return ctx.String(http.StatusBadRequest, err.Error())
			}
			reply, _ := out.(*publicpaymentv1.EPayNotifyReply)
			response := "success"
			if reply != nil && strings.TrimSpace(reply.Response) != "" {
				response = reply.Response
			}
			return ctx.String(http.StatusOK, response)

		case payment.Stripe:
			payload, err := io.ReadAll(ctx.Request().Body)
			if err != nil {
				return compatJSONError(ctx, err)
			}
			req := &publicpaymentv1.StripeNotifyRequest{
				Token:           pathReq.Token,
				Payload:         payload,
				StripeSignature: ctx.Request().Header.Get("Stripe-Signature"),
			}
			out, err := compatMiddleware(ctx, req, func(inner context.Context, request interface{}) (interface{}, error) {
				return publicPayment.StripeNotify(inner, request.(*publicpaymentv1.StripeNotifyRequest))
			})
			if err != nil {
				return compatJSONError(ctx, err)
			}
			reply, _ := out.(*publicpaymentv1.StripeNotifyReply)
			if reply != nil && reply.Code != 0 {
				return ctx.JSON(http.StatusOK, compatEnvelope{Code: int(reply.Code), Msg: reply.Message})
			}
			return compatJSON(ctx, nil)

		case payment.AlipayF2F:
			req := compatBuildAlipayNotifyRequest(ctx, pathReq.Token)
			out, err := compatMiddleware(ctx, req, func(inner context.Context, request interface{}) (interface{}, error) {
				return publicPayment.AlipayNotify(inner, request.(*publicpaymentv1.AlipayNotifyRequest))
			})
			if err != nil {
				return compatJSONError(ctx, err)
			}
			reply, _ := out.(*publicpaymentv1.AlipayNotifyReply)
			response := "success"
			if reply != nil && strings.TrimSpace(reply.Response) != "" {
				response = reply.Response
			}
			return ctx.String(http.StatusOK, response)

		default:
			return compatJSON(ctx, nil)
		}
	}
}

func compatAppleLoginCallbackHandler(dataLayer *data.Data, appConf *conf.Application, logger log.Logger) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		helper := log.NewHelper(logger)
		if err := ctx.Request().ParseForm(); err != nil {
			helper.Errorf("[compatAppleLoginCallback] parse form failed: %v", err)
		}

		redirectHost := ""
		if appConf != nil && appConf.Site != nil {
			redirectHost = strings.TrimSpace(appConf.Site.Host)
		}

		state := strings.TrimSpace(ctx.Request().FormValue("state"))
		code := strings.TrimSpace(ctx.Request().FormValue("code"))
		if dataLayer == nil || dataLayer.Redis() == nil || state == "" {
			http.Redirect(ctx.Response(), ctx.Request(), redirectHost, http.StatusTemporaryRedirect)
			return nil
		}

		result, err := dataLayer.Redis().Get(ctx, fmt.Sprintf("apple:%s", state)).Result()
		if err != nil || strings.TrimSpace(result) == "" {
			helper.Errorf("[compatAppleLoginCallback] load state failed: %v", err)
			http.Redirect(ctx.Response(), ctx.Request(), redirectHost, http.StatusTemporaryRedirect)
			return nil
		}

		target := fmt.Sprintf("%s?method=apple&code=%s&state=%s", result, code, state)
		http.Redirect(ctx.Response(), ctx.Request(), target, http.StatusFound)
		helper.Infof("[compatAppleLoginCallback] redirect url: %s", target)
		return nil
	}
}

func compatBuildEPayNotifyRequest(ctx khttp.Context, token string) *publicpaymentv1.EPayNotifyRequest {
	form := ctx.Request().URL.Query()
	extras := make(map[string]string)
	for key, values := range form {
		if len(values) == 0 {
			continue
		}
		extras[key] = values[0]
	}

	delete(extras, "pid")
	delete(extras, "trade_no")
	delete(extras, "out_trade_no")
	delete(extras, "type")
	delete(extras, "name")
	delete(extras, "money")
	delete(extras, "trade_status")
	delete(extras, "param")
	delete(extras, "sign")

	return &publicpaymentv1.EPayNotifyRequest{
		Token:       token,
		Pid:         form.Get("pid"),
		TradeNo:     form.Get("trade_no"),
		OutTradeNo:  form.Get("out_trade_no"),
		Type:        form.Get("type"),
		Name:        form.Get("name"),
		Money:       form.Get("money"),
		TradeStatus: form.Get("trade_status"),
		Param:       form.Get("param"),
		Sign:        form.Get("sign"),
		ExtraParams: extras,
	}
}

func compatBuildAlipayNotifyRequest(ctx khttp.Context, token string) *publicpaymentv1.AlipayNotifyRequest {
	form := ctx.Form()
	extras := make(map[string]string)
	for key, values := range form {
		if len(values) == 0 {
			continue
		}
		extras[key] = values[0]
	}

	for _, key := range []string{
		"trade_no",
		"out_trade_no",
		"trade_status",
		"total_amount",
		"receipt_amount",
		"buyer_pay_amount",
		"subject",
		"body",
		"gmt_create",
		"gmt_payment",
		"notify_time",
		"app_id",
		"seller_id",
		"seller_email",
		"notify_type",
		"auth_app_id",
		"charset",
		"version",
		"sign",
	} {
		delete(extras, key)
	}

	return &publicpaymentv1.AlipayNotifyRequest{
		Token:          token,
		TradeNo:        form.Get("trade_no"),
		OutTradeNo:     form.Get("out_trade_no"),
		TradeStatus:    form.Get("trade_status"),
		TotalAmount:    form.Get("total_amount"),
		ReceiptAmount:  form.Get("receipt_amount"),
		BuyerPayAmount: form.Get("buyer_pay_amount"),
		Subject:        form.Get("subject"),
		Body:           form.Get("body"),
		GmtCreate:      form.Get("gmt_create"),
		GmtPayment:     form.Get("gmt_payment"),
		NotifyTime:     form.Get("notify_time"),
		AppId:          form.Get("app_id"),
		SellerId:       form.Get("seller_id"),
		SellerEmail:    form.Get("seller_email"),
		NotifyType:     form.Get("notify_type"),
		AuthAppId:      form.Get("auth_app_id"),
		Charset:        form.Get("charset"),
		Version:        form.Get("version"),
		Sign:           form.Get("sign"),
		ExtraParams:    extras,
	}
}
