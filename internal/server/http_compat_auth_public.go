package server

import (
	"bytes"
	"context"
	"encoding/json"

	authoauthv1 "github.com/OmnTeam/ppanel-pro/api/auth/oauth/v1"
	publicauthv1 "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	authservice "github.com/OmnTeam/ppanel-pro/internal/service/auth"
	authoauthservice "github.com/OmnTeam/ppanel-pro/internal/service/auth/oauth"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatOAuthLoginGetTokenRequest struct {
	Method    string          `json:"method"`
	Callback  json.RawMessage `json:"callback"`
	IP        string          `json:"ip"`
	UserAgent string          `json:"user_agent"`
}

func registerLegacyPublicAuthCompatRoutes(r *khttp.Router, authSvc *authservice.AuthService, oauthSvc *authoauthservice.OAuthService) {
	if r == nil {
		return
	}

	if authSvc != nil {
		r.GET("/v1/auth/check", compatCheckUserHandler(authSvc))
		r.GET("/v1/auth/check/telephone", compatCheckUserTelephoneHandler(authSvc))
		r.POST("/v1/auth/login", compatUserLoginHandler(authSvc))
		r.POST("/v1/auth/login/telephone", compatTelephoneLoginHandler(authSvc))
		r.POST("/v1/auth/register", compatUserRegisterHandler(authSvc))
		r.POST("/v1/auth/register/telephone", compatTelephoneRegisterHandler(authSvc))
		r.POST("/v1/auth/reset", compatResetPasswordHandler(authSvc))
		r.POST("/v1/auth/reset/telephone", compatTelephoneResetPasswordHandler(authSvc))
	}

	if oauthSvc != nil {
		r.POST("/v1/auth/oauth/login", compatOAuthLoginHandler(oauthSvc))
		r.POST("/v1/auth/oauth/login/token", compatOAuthLoginGetTokenHandler(oauthSvc))
	}
}

func compatCheckUserHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.CheckUserRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Email, "CheckUserRequest", "Email"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return authSvc.CheckUser(inner, request.(*publicauthv1.CheckUserRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		reply, _ := out.(*publicauthv1.CheckUserReply)
		return compatJSON(ctx, map[string]bool{"exist": reply != nil && reply.Data != nil && reply.Data.Exist})
	}
}

func compatCheckUserTelephoneHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.CheckUserTelephoneRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Telephone, "TelephoneCheckUserRequest", "Telephone"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.TelephoneAreaCode, "TelephoneCheckUserRequest", "TelephoneAreaCode"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return authSvc.CheckUserTelephone(inner, request.(*publicauthv1.CheckUserTelephoneRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		reply, _ := out.(*publicauthv1.CheckUserTelephoneReply)
		return compatJSON(ctx, map[string]bool{"exist": reply != nil && reply.Data != nil && reply.Data.Exist})
	}
}

func compatUserLoginHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
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
			return authSvc.UserLogin(inner, request.(*publicauthv1.UserLoginRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, compatLoginReplyData(out.(*publicauthv1.LoginReply)))
	}
}

func compatTelephoneLoginHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.TelephoneLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Telephone, "TelephoneLoginRequest", "Telephone"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.TelephoneAreaCode, "TelephoneLoginRequest", "TelephoneAreaCode"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return authSvc.TelephoneLogin(inner, request.(*publicauthv1.TelephoneLoginRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, compatLoginReplyData(out.(*publicauthv1.LoginReply)))
	}
}

func compatUserRegisterHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.UserRegisterRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Email, "UserRegisterRequest", "Email"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Password, "UserRegisterRequest", "Password"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return authSvc.UserRegister(inner, request.(*publicauthv1.UserRegisterRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, compatLoginReplyData(out.(*publicauthv1.LoginReply)))
	}
}

func compatTelephoneRegisterHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.TelephoneRegisterRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Telephone, "TelephoneRegisterRequest", "Telephone"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.TelephoneAreaCode, "TelephoneRegisterRequest", "TelephoneAreaCode"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Password, "TelephoneRegisterRequest", "Password"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return authSvc.TelephoneRegister(inner, request.(*publicauthv1.TelephoneRegisterRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, compatLoginReplyData(out.(*publicauthv1.LoginReply)))
	}
}

func compatResetPasswordHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
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
			return authSvc.ResetPassword(inner, request.(*publicauthv1.ResetPasswordRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, compatLoginReplyData(out.(*publicauthv1.LoginReply)))
	}
}

func compatTelephoneResetPasswordHandler(authSvc *authservice.AuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req publicauthv1.TelephoneResetPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Telephone, "TelephoneResetPasswordRequest", "Telephone"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.TelephoneAreaCode, "TelephoneResetPasswordRequest", "TelephoneAreaCode"); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Password, "TelephoneResetPasswordRequest", "Password"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return authSvc.TelephoneResetPassword(inner, request.(*publicauthv1.TelephoneResetPasswordRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, compatLoginReplyData(out.(*publicauthv1.LoginReply)))
	}
}

func compatOAuthLoginHandler(oauthSvc *authoauthservice.OAuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req authoauthv1.OAuthLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Method, "OAthLoginRequest", "Method"); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return oauthSvc.OAuthLogin(inner, request.(*authoauthv1.OAuthLoginRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		reply, _ := out.(*authoauthv1.LoginRedirectReply)
		redirect := ""
		if reply != nil && reply.Data != nil {
			redirect = reply.Data.Redirect
		}
		return compatJSON(ctx, map[string]string{"redirect": redirect})
	}
}

func compatOAuthLoginGetTokenHandler(oauthSvc *authoauthservice.OAuthService) func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req compatOAuthLoginGetTokenRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(req.Method, "OAuthLoginGetTokenRequest", "Method"); err != nil {
			return compatJSONError(ctx, err)
		}

		callback, err := compatJSONString(req.Callback)
		if err != nil {
			return compatJSONError(ctx, err)
		}
		if err := compatValidateRequiredString(callback, "OAuthLoginGetTokenRequest", "Callback"); err != nil {
			return compatJSONError(ctx, err)
		}

		serviceReq := &authoauthv1.OAuthLoginGetTokenRequest{
			Method:    req.Method,
			Callback:  callback,
			Ip:        req.IP,
			UserAgent: req.UserAgent,
		}
		out, err := compatMiddleware(ctx, serviceReq, func(inner context.Context, request interface{}) (interface{}, error) {
			return oauthSvc.OAuthLoginGetToken(inner, request.(*authoauthv1.OAuthLoginGetTokenRequest))
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}

		reply, _ := out.(*authoauthv1.LoginTokenReply)
		token := ""
		if reply != nil && reply.Data != nil {
			token = reply.Data.Token
		}
		return compatJSON(ctx, map[string]string{"token": token})
	}
}

func compatJSONString(raw json.RawMessage) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil
	}
	if trimmed[0] == '"' {
		var result string
		if err := json.Unmarshal(trimmed, &result); err != nil {
			return "", err
		}
		return result, nil
	}
	return string(trimmed), nil
}

func compatLoginReplyData(reply *publicauthv1.LoginReply) map[string]string {
	token := ""
	if reply != nil && reply.Data != nil {
		token = reply.Data.Token
	}
	return map[string]string{"token": token}
}
