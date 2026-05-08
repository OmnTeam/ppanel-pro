package server

import (
	"github.com/OmnTeam/npanel-pro/internal/conf"
	"github.com/OmnTeam/npanel-pro/internal/data"
	middleware2 "github.com/OmnTeam/npanel-pro/internal/middleware"
	"github.com/google/wire"
)

// ProviderSet is server providers
var ProviderSet = wire.NewSet(NewMiddlewareServiceContext, NewGRPCServer, NewHTTPServer)

func NewMiddlewareServiceContext(c *conf.Server, appConf *conf.Application, d *data.Data) *middleware2.ServiceContext {
	deviceConfig := middleware2.DeviceConfig{}
	if appConf != nil && appConf.Device != nil {
		deviceConfig = middleware2.DeviceConfig{
			Enable:         appConf.Device.Enable,
			SecuritySecret: appConf.Device.SecuritySecret,
		}
	}
	return &middleware2.ServiceContext{
		Config:       c,
		Redis:        d.Redis(),
		UserModel:    d,
		DeviceConfig: deviceConfig,
	}
}
