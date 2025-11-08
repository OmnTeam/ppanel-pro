//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build

package main

import (
	"github.com/OmnTeam/ppanel-pro/internal/biz"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/server"
	"github.com/OmnTeam/ppanel-pro/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application
func wireApp(*conf.Server, *conf.Data, *conf.Application, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
