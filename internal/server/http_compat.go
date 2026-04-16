package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	admingroupservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/group"
	adminpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/payment"
	adminsystemservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/system"
	adminticketservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/ticket"
	authservice "github.com/OmnTeam/ppanel-pro/internal/service/auth"
	authoauthservice "github.com/OmnTeam/ppanel-pro/internal/service/auth/oauth"
	publicpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/payment"

	publicauthv1 "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	authbiz "github.com/OmnTeam/ppanel-pro/internal/biz/auth"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatEnvelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type compatSliderVerifyRequest struct {
	ID    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Trail string `json:"trail"`
}

type compatDeviceLoginRequest struct {
	Identifier string `json:"identifier"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	CfToken    string `json:"cf_token"`
	ShortCode  string `json:"short_code"`
	LoginType  string `json:"login_type"`
}

func registerLegacyCompatRoutes(
	srv *khttp.Server,
	authCompat *data.AuthCompat,
	authSvc *authservice.AuthService,
	oauthSvc *authoauthservice.OAuthService,
	dataLayer *data.Data,
	appConf *conf.Application,
	adminGroup *admingroupservice.GroupService,
	adminPayment *adminpaymentservice.PaymentService,
	adminSystem *adminsystemservice.SystemService,
	adminTicket *adminticketservice.TicketService,
	publicOrder legacyPublicOrderCompat,
	publicPayment *publicpaymentservice.PaymentService,
	publicPortal legacyPublicPortalCompat,
	publicTicket legacyPublicTicketCompat,
	publicUser legacyPublicUserCompat,
	logger log.Logger,
) {
	if srv == nil {
		return
	}

	r := srv.Route("/")

	registerLegacyAuthCompatRoutes(r, authCompat, authSvc, oauthSvc)
	registerLegacyCommonCompatRoutes(r, dataLayer, appConf, logger)
	registerLegacyAdminCompatRoutes(r, dataLayer, adminGroup, adminPayment, adminSystem, adminTicket)
	registerLegacyPublicCompatRoutes(r, dataLayer, appConf, publicOrder, publicPayment, publicPortal, publicTicket, publicUser)
	registerLegacyCallbackCompatRoutes(r, dataLayer, appConf, publicPayment, logger)
	registerLegacyServerCompatRoutes(r, dataLayer)
}

func registerLegacyAuthCompatRoutes(r *khttp.Router, authCompat *data.AuthCompat, authSvc *authservice.AuthService, oauthSvc *authoauthservice.OAuthService) {
	if r == nil {
		return
	}

	if authCompat != nil {
		r.GET("/v1/common/site/config", func(ctx khttp.Context) error {
			out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
				return authCompat.GetLegacyGlobalConfig(inner)
			})
			if err != nil {
				return compatJSONError(ctx, err)
			}
			return compatJSON(ctx, out)
		})

		r.GET("/v1/common/heartbeat", func(ctx khttp.Context) error {
			out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
				return authCompat.Heartbeat(), nil
			})
			if err != nil {
				return compatJSONError(ctx, err)
			}
			return compatJSON(ctx, out)
		})

		r.POST("/v1/auth/captcha/generate", func(ctx khttp.Context) error {
			out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
				return authCompat.GenerateCaptcha(inner)
			})
			if err != nil {
				return compatJSONError(ctx, err)
			}
			return compatJSON(ctx, out)
		})

		r.POST("/v1/auth/captcha/slider/verify", sliderVerifyHandler(authCompat))
		r.POST("/v1/auth/login/device", deviceLoginHandler(authCompat))

		r.POST("/v1/auth/admin/captcha/generate", func(ctx khttp.Context) error {
			out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
				return authCompat.GenerateCaptcha(inner)
			})
			if err != nil {
				return compatJSONError(ctx, err)
			}
			return compatJSON(ctx, out)
		})

		r.POST("/v1/auth/admin/captcha/slider/verify", sliderVerifyHandler(authCompat))
		r.POST("/v1/auth/admin/login", adminLoginHandler(authCompat))
		r.POST("/v1/auth/admin/reset", adminResetHandler(authCompat))
	}

	registerLegacyPublicAuthCompatRoutes(r, authSvc, oauthSvc)
}

func sliderVerifyHandler(authCompat *data.AuthCompat) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatSliderVerifyRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.ID, "SliderVerifyCaptchaRequest", "Id"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredInt(req.X, "SliderVerifyCaptchaRequest", "X"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredInt(req.Y, "SliderVerifyCaptchaRequest", "Y"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatSliderVerifyRequest)
			return authCompat.VerifySliderCaptcha(inner, in.ID, in.X, in.Y, in.Trail)
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	}
}

func deviceLoginHandler(authCompat *data.AuthCompat) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatDeviceLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Identifier, "DeviceLoginRequest", "Identifier"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.UserAgent, "DeviceLoginRequest", "UserAgent"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatDeviceLoginRequest)
			meta := buildCompatRequestMeta(inner, in.IP, in.UserAgent, in.LoginType, in.Identifier, in.CfToken, "", "", "")
			if meta.LoginType == "" {
				meta.LoginType = "device"
			}
			return authCompat.DeviceLogin(inner, &data.DeviceLoginParams{
				Identifier: in.Identifier,
				ShortCode:  in.ShortCode,
				Meta:       meta,
			})
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		return compatJSON(ctx, loginData(out))
	}
}

func adminLoginHandler(authCompat *data.AuthCompat) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.UserLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Email, "UserLoginRequest", "Email"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Password, "UserLoginRequest", "Password"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*publicauthv1.UserLoginRequest)
			return authCompat.AdminLogin(inner, &data.AdminLoginParams{
				Email:    in.Email,
				Password: in.Password,
				Meta:     buildCompatRequestMeta(inner, in.Ip, in.UserAgent, in.LoginType, in.Identifier, in.CfToken, in.CaptchaId, in.CaptchaCode, in.SliderToken),
			})
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		return compatJSON(ctx, loginData(out))
	}
}

func adminResetHandler(authCompat *data.AuthCompat) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.ResetPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Email, "ResetPasswordRequest", "Email"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Password, "ResetPasswordRequest", "Password"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*publicauthv1.ResetPasswordRequest)
			return authCompat.AdminResetPassword(inner, &data.AdminResetPasswordParams{
				Email:    in.Email,
				Password: in.Password,
				Code:     in.Code,
				Meta:     buildCompatRequestMeta(inner, in.Ip, in.UserAgent, in.LoginType, in.Identifier, in.CfToken, in.CaptchaId, in.CaptchaCode, in.SliderToken),
			})
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		return compatJSON(ctx, loginData(out))
	}
}

func compatMiddleware(ctx khttp.Context, req interface{}, fn func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
	return ctx.Middleware(fn)(ctx, req)
}

func compatJSON(ctx khttp.Context, data interface{}) error {
	return ctx.JSON(200, compatSuccess(data))
}

func compatJSONError(ctx khttp.Context, err error) error {
	return ctx.JSON(200, compatError(err))
}

func compatSuccess(data interface{}) compatEnvelope {
	return compatEnvelope{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

func compatError(err error) compatEnvelope {
	code := responsecode.ErrInternalError
	msg := "Internal Server Error"

	if err != nil {
		msg = err.Error()
	}
	if se := kerrors.FromError(err); se != nil {
		if customCode, ok := se.Metadata["custom_code"]; ok {
			if parsed, parseErr := parseCompatCode(customCode); parseErr == nil {
				code = parsed
			}
		}
		if strings.TrimSpace(se.Message) != "" {
			msg = se.Message
		}
	}

	return compatEnvelope{
		Code: code,
		Msg:  msg,
	}
}

func parseCompatCode(raw string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(raw))
}

func compatCodeError(code int, msg string) error {
	if strings.TrimSpace(msg) == "" {
		msg = "error"
	}
	return kerrors.New(200, fmt.Sprintf("COMPAT_%d", code), msg).WithMetadata(map[string]string{
		"custom_code": strconv.Itoa(code),
	})
}

func compatParamError(msg string) error {
	return compatCodeError(400, msg)
}

func loginData(result interface{}) map[string]string {
	loginResult := result.(*authbiz.LoginResult)
	return map[string]string{"token": loginResult.Token}
}

func buildCompatRequestMeta(ctx context.Context, fallbackIP, fallbackUserAgent, fallbackLoginType, identifier, cfToken, captchaID, captchaCode, sliderToken string) authbiz.RequestMeta {
	meta := authbiz.RequestMeta{
		Identifier:  identifier,
		LoginType:   fallbackLoginType,
		IP:          fallbackIP,
		UserAgent:   fallbackUserAgent,
		CfToken:     cfToken,
		CaptchaID:   captchaID,
		CaptchaCode: captchaCode,
		SliderToken: sliderToken,
	}

	if tr, ok := transport.FromServerContext(ctx); ok {
		if loginType := firstCompatHeader(tr, "Login-Type"); loginType != "" {
			meta.LoginType = loginType
		}
		if ip := firstCompatHeader(tr, "X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"); ip != "" {
			meta.IP = firstCompatForwardedIP(ip)
		}
		if userAgent := firstCompatHeader(tr, "User-Agent"); userAgent != "" {
			meta.UserAgent = userAgent
		}
	}

	if meta.LoginType == "" {
		if value, ok := ctx.Value(constant.LoginType).(string); ok {
			meta.LoginType = value
		}
	}
	if meta.Identifier == "" {
		if value, ok := ctx.Value(constant.CtxKeyIdentifier).(string); ok {
			meta.Identifier = value
		}
	}

	return meta
}

func firstCompatHeader(tr transport.Transporter, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(tr.RequestHeader().Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func firstCompatForwardedIP(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ",")
	return strings.TrimSpace(parts[0])
}
