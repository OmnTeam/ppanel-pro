package main

import (
	"flag"
	"os"

	"github.com/OmnTeam/ppanel-pro/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software
	Name string
	// Version is the version of the compiled software
	Version string
	// flagconf is the config flag
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()

	// 抑制 Redis 客户端的警告信息
	os.Setenv("REDIS_LOG_LEVEL", "ERROR")

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	conf.SetLegacyDebugMode(bc.GetDebug())

	// 调试：打印bootstrap配置
	if bc.App != nil {
		log.NewHelper(logger).Infof("Bootstrap.App is not nil: %+v", bc.App)
		if bc.App.Admin != nil {
			log.NewHelper(logger).Infof("Admin config found in bootstrap: email=%s", bc.App.Admin.Email)
		} else {
			log.NewHelper(logger).Warnf("Admin config is nil in bootstrap.App")
		}
	} else {
		log.NewHelper(logger).Warnf("Bootstrap.App is nil")
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.App, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
