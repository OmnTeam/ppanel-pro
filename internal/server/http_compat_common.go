package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	publicpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/payment"
	serverservice "github.com/OmnTeam/ppanel-pro/internal/service/server"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatEnvelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func registerLegacyCompatRoutes(
	srv *khttp.Server,
	_ *data.AuthCompat,
	dataLayer *data.Data,
	appConf *conf.Application,
	publicPayment *publicpaymentservice.PaymentService,
	serverService *serverservice.ServerService,
	logger log.Logger,
) {
	if srv == nil {
		return
	}

	r := srv.Route("/")
	registerLegacyCallbackCompatRoutes(r, dataLayer, appConf, publicPayment, logger)
	registerLegacyServerCompatRoutes(r, dataLayer, serverService)
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
