package data

import (
	"context"
	"fmt"
	"sync"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/migrate"
	"github.com/OmnTeam/ppanel-pro/internal/queue/handler"
	"github.com/OmnTeam/ppanel-pro/internal/service"
	"github.com/OmnTeam/ppanel-pro/pkg/device"

	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Global device manager instance
var (
	globalDeviceManager *device.DeviceManager
	deviceManagerOnce   sync.Once
)

// ProviderSet is data providers
var ProviderSet = wire.NewSet(
	NewData,
	NewEntClient,
	NewDeviceManager,
	NewAdsRepo,
	NewAdminAnnouncementRepo,
	NewAdminAuthMethodRepo,
	NewAdminConsoleRepo,
	NewCouponRepo,
	NewAdminDocumentRepo,
	NewAdminSystemLogRepo,
	NewAdminTrafficLogRepo,
	NewAdminLogSettingRepo,
	NewAdminMarketingRepo,
	NewOrderRepo,
	NewAdminPaymentRepo,
	NewAdminServerRepo,
	NewAdminNodeRepo,
	NewAdminMigrationRepo,
	NewSubscribeApplicationRepo,
	NewSubscribeRepo,
	NewAdminSystemRepo,
	NewTicketRepo,
	NewAdminRedemptionRepo,
	NewAdminGroupRepo,
	// Admin User模块仓储
	NewAdminUserRepo,
	NewAdminUserAuthMethodRepo,
	NewAdminUserDeviceRepo,
	NewAdminUserSubscribeRepo,
	// Auth模块仓储
	NewAuthRepo,
	// Public Common模块仓储
	NewCommonRepo,
	// Public Order模块仓储
	NewPublicOrderRepo,
	// Public Announcement模块仓储
	NewPublicAnnouncementRepo,
	// Public Document模块仓储
	NewPublicDocumentRepo,
	// Public Portal模块仓储
	NewPublicPortalRepo,
	// Public Ticket模块仓储
	NewPublicTicketRepo,
	// Public User模块仓储
	NewPublicUserRepo,
	// Public Payment模块仓储
	NewPublicPaymentRepo,
	// Public Subscribe模块仓储
	NewPublicSubscribeRepo,
	// Public Withdrawal模块仓储
	NewWithdrawalRepo,
	// Server模块仓储
	NewServerNodeRepo,
	// Auth OAuth模块仓储
	NewOAuthRepo,
)

// Data
type Data struct {
	db          *ent.Client
	rdb         *redis.Client
	queue       *asynq.Client
	queueServer *asynq.Server
	conf        *conf.Application     // 应用配置（包含JWT等配置）
	deviceMgr   *device.DeviceManager // 设备管理器
}

// DB 获取数据库客户端
func (d *Data) DB() *ent.Client {
	return d.db
}

// RDB 获取Redis客户端
func (d *Data) RDB() *redis.Client {
	return d.rdb
}

// Redis 获取Redis客户端 (为中间件兼容)
func (d *Data) Redis() *redis.Client {
	return d.rdb
}

// DeviceManager 获取设备管理器
func (d *Data) DeviceManager() *device.DeviceManager {
	return d.deviceMgr
}

// FindOne 实现用户服务接口 - 根据用户ID查找用户
func (d *Data) FindOne(ctx context.Context, userId int) (*ent.ProxyUser, error) {
	return d.db.ProxyUser.Get(ctx, int64(userId))
}

// NewData
func NewData(c *conf.Data, appConf *conf.Application, logger log.Logger) (*Data, func(), error) {
	log.NewHelper(logger).Infof("connecting to database: %s", c.Database.Source)

	client, err := ent.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		log.NewHelper(logger).Errorf("failed opening connection to database: %v", err)
		return nil, nil, err
	}
	client = client.Debug()
	// 运行自动迁移工具
	if err := client.Schema.Create(context.Background()); err != nil {
		log.NewHelper(logger).Errorf("failed creating schema resources: %v", err)
		return nil, nil, err
	}
	log.NewHelper(logger).Infof("database schema migration completed successfully")

	// 初始化默认数据 - 使用迁移系统
	log.NewHelper(logger).Infof("starting default data initialization...")
	log.NewHelper(logger).Infof("appConf value: %+v", appConf)
	if appConf != nil {
		log.NewHelper(logger).Infof("app config found, creating migrator...")
		log.NewHelper(logger).Infof("app config site: %+v", appConf.Site)
		if appConf.Admin != nil {
			log.NewHelper(logger).Infof("admin config found: email=%s", appConf.Admin.Email)
		} else {
			log.NewHelper(logger).Warnf("admin config is nil in app config")
		}
		migrator := migrate.NewMigrator(client, logger, appConf)
		log.NewHelper(logger).Infof("migrator created, starting AutoMigrateWithData...")
		if err := migrator.AutoMigrateWithData(context.Background()); err != nil {
			log.NewHelper(logger).Errorf("failed to initialize default data: %v", err)
			return nil, nil, fmt.Errorf("failed to initialize default data: %w", err)
		}
		log.NewHelper(logger).Infof("default data initialization completed successfully")
	} else {
		log.NewHelper(logger).Warnf("app config is nil, skipping default data initialization")
		// 尝试使用默认配置进行迁移
		log.NewHelper(logger).Infof("trying to use default config for migration...")
		migrator := migrate.NewMigrator(client, logger, nil)
		if err := migrator.AutoMigrateWithData(context.Background()); err != nil {
			log.NewHelper(logger).Errorf("failed to initialize default data with nil config: %v", err)
			// 不返回错误，只记录日志，允许服务启动
		} else {
			log.NewHelper(logger).Infof("default data initialization completed successfully with default config")
		}
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		PoolSize:     int(c.Redis.PoolSize),
		MinIdleConns: int(c.Redis.MinIdleConns),
	})

	// 测试 Redis 连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.NewHelper(logger).Errorf("failed connecting to redis: %v", err)
		return nil, nil, err
	}
	log.NewHelper(logger).Infof("connected to redis: %s", c.Redis.Addr)

	// 创建 asynq 客户端 - 用于发送任务到队列
	redisOpt := asynq.RedisClientOpt{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       int(c.Redis.Db),
	}
	queueClient := asynq.NewClient(redisOpt)

	// 创建 asynq 服务器 - 用于处理队列中的任务
	queueServer := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10, // 并发处理10个任务
			// Logger 使用默认的 asynq logger
		},
	)

	// 初始化设备管理器
	deviceManager := device.NewDeviceManager(logger, 60, 30) // 心跳超时60秒，检查间隔30秒

	d := &Data{
		db:          client,
		rdb:         rdb,
		queue:       queueClient,
		queueServer: queueServer,
		conf:        appConf,
		deviceMgr:   deviceManager,
	}

	// 启动 asynq 队列服务器
	mux := asynq.NewServeMux()
	// 创建缓存服务
	cacheService := service.NewCacheService(rdb, client, logger)
	handler.RegisterHandlers(mux, client, rdb, queueClient, appConf, cacheService, logger)
	go func() {
		if err := queueServer.Start(mux); err != nil {
			log.NewHelper(logger).Fatalf("Failed to start asynq server: %v", err)
		}
	}()
	log.NewHelper(logger).Info("Asynq queue server started successfully")

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if err := d.db.Close(); err != nil {
			log.NewHelper(logger).Error(err)
		}
		if err := d.rdb.Close(); err != nil {
			log.NewHelper(logger).Error(err)
		}
		if err := d.queue.Close(); err != nil {
			log.NewHelper(logger).Error(err)
		}
		// 关闭 asynq 服务器
		d.queueServer.Shutdown()
		log.NewHelper(logger).Info("asynq server stopped")
	}

	return d, cleanup, nil
}

// NewDeviceManager 提供设备管理器
func NewDeviceManager(d *Data) *device.DeviceManager {
	// 设置全局设备管理器实例
	deviceManagerOnce.Do(func() {
		globalDeviceManager = d.deviceMgr
	})
	return d.deviceMgr
}

// GetGlobalDeviceManager 获取全局设备管理器
func GetGlobalDeviceManager() *device.DeviceManager {
	return globalDeviceManager
}

// NewEntClient 提供 ent.Client 给需要直接访问数据库的服务
func NewEntClient(d *Data) *ent.Client {
	return d.db
}
