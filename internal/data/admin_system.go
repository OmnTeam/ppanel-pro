package data

import (
	"context"
	"encoding/json"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	systembiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/system"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type adminSystemRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminSystemRepo creates a new admin system repository
func NewAdminSystemRepo(data *Data, logger log.Logger) systembiz.SystemRepo {
	return &adminSystemRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetConfigByCategory 根据分类获取配置
func (r *adminSystemRepo) GetConfigByCategory(ctx context.Context, category string) ([]*tool.SystemConfig, error) {
	// Determine cache key based on category
	var cacheKey string
	switch category {
	case "currency":
		cacheKey = CurrencyConfigKey
	case "invite":
		cacheKey = InviteConfigKey
	case "server":
		cacheKey = NodeConfigKey
	case "tos":
		cacheKey = TosConfigKey
	case "register":
		cacheKey = RegisterConfigKey
	case "site":
		cacheKey = SiteConfigKey
	case "subscribe":
		cacheKey = SubscribeConfigKey
	case "verify_code":
		cacheKey = VerifyCodeConfigKey
	case "verify":
		cacheKey = VerifyConfigKey
	default:
		cacheKey = fmt.Sprintf("system:%s_config", category)
	}

	// Try to get from cache first
	var configs []*tool.SystemConfig
	result, err := r.data.rdb.Get(ctx, cacheKey).Result()
	if err == nil && result != "" {
		if err := json.Unmarshal([]byte(result), &configs); err == nil {
			return configs, nil
		}
	}

	// Query from database
	systems, err := r.data.db.ProxySystem.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ(s.C(proxysystem.FieldCategory), category))
		}).
		All(ctx)
	if err != nil {
		r.log.Errorf("[GetConfigByCategory] Failed to query system config for category %s: %v", category, err)
		return nil, responsecode.NewDatabaseQueryError()
	}

	// Convert to tool.SystemConfig
	configs = make([]*tool.SystemConfig, 0, len(systems))
	for _, sys := range systems {
		configs = append(configs, &tool.SystemConfig{
			Key:   sys.Key,
			Value: sys.Value,
			Type:  sys.Type,
		})
	}

	// Cache the result
	if data, err := json.Marshal(configs); err == nil {
		// Cache for 5 minutes
		if err := r.data.rdb.Set(ctx, cacheKey, data, 300).Err(); err != nil {
			r.log.Warnf("Failed to cache system config for category %s: %v", category, err)
		}
	}

	return configs, nil
}

// UpdateConfigByCategory 根据分类更新配置
func (r *adminSystemRepo) UpdateConfigByCategory(ctx context.Context, category string, configs map[string]*tool.SystemConfig) error {
	// Start transaction
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		r.log.Errorf("[UpdateConfigByCategory] Failed to start transaction for category %s: %v", category, err)
		return responsecode.NewDatabaseUpdateError()
	}

	// Update each config
	for key, config := range configs {
		// Check if config exists
		exists, err := tx.ProxySystem.Query().
			Where(func(s *sql.Selector) {
				s.Where(sql.And(
					sql.EQ(s.C(proxysystem.FieldCategory), category),
					sql.EQ(s.C(proxysystem.FieldKey), key),
				))
			}).
			Exist(ctx)
		if err != nil {
			r.log.Errorf("[UpdateConfigByCategory] Failed to check config existence for category %s, key %s: %v", category, key, err)
			tx.Rollback()
			return responsecode.NewDatabaseQueryError()
		}

		if exists {
			// Update existing config
			affected, err := tx.ProxySystem.Update().
				Where(func(s *sql.Selector) {
					s.Where(sql.And(
						sql.EQ(s.C(proxysystem.FieldCategory), category),
						sql.EQ(s.C(proxysystem.FieldKey), key),
					))
				}).
				SetValue(config.Value).
				SetType(config.Type).
				Save(ctx)
			if err != nil {
				r.log.Errorf("[UpdateConfigByCategory] Failed to update config for category %s, key %s: %v", category, key, err)
				tx.Rollback()
				return responsecode.NewDatabaseUpdateError()
			}
			if affected == 0 {
				r.log.Warnf("[UpdateConfigByCategory] Config not found for category %s, key %s", category, key)
				tx.Rollback()
				return responsecode.NewSystemNotFoundError()
			}
		} else {
			// Create new config
			_, err := tx.ProxySystem.Create().
				SetCategory(category).
				SetKey(key).
				SetValue(config.Value).
				SetType(config.Type).
				SetDesc("").
				Save(ctx)
			if err != nil {
				r.log.Errorf("[UpdateConfigByCategory] Failed to create config for category %s, key %s: %v", category, key, err)
				tx.Rollback()
				return responsecode.NewDatabaseUpdateError()
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		r.log.Errorf("[UpdateConfigByCategory] Failed to commit transaction for category %s: %v", category, err)
		return responsecode.NewDatabaseUpdateError()
	}

	// Clear cache
	var cacheKeys []string
	switch category {
	case "currency":
		cacheKeys = []string{
			CurrencyConfigKey,
			GlobalConfigKey,
		}
	case "invite":
		cacheKeys = []string{
			InviteConfigKey,
			GlobalConfigKey,
		}
	case "server":
		cacheKeys = []string{
			NodeConfigKey,
			GlobalConfigKey,
		}
	case "tos":
		cacheKeys = []string{
			TosConfigKey,
			GlobalConfigKey,
		}
	case "register":
		cacheKeys = []string{
			RegisterConfigKey,
			GlobalConfigKey,
		}
	case "site":
		cacheKeys = []string{
			SiteConfigKey,
			GlobalConfigKey,
		}
	case "subscribe":
		cacheKeys = []string{
			SubscribeConfigKey,
			GlobalConfigKey,
		}
	case "verify_code":
		cacheKeys = []string{
			VerifyCodeConfigKey,
			GlobalConfigKey,
		}
	case "verify":
		cacheKeys = []string{
			VerifyConfigKey,
			GlobalConfigKey,
		}
	default:
		cacheKeys = []string{
			fmt.Sprintf("system:%s_config", category),
			GlobalConfigKey,
		}
	}

	for _, cacheKey := range cacheKeys {
		if err := r.data.rdb.Del(ctx, cacheKey).Err(); err != nil && err != redis.Nil {
			r.log.Warnf("Failed to delete cache key %s: %v", cacheKey, err)
		}
	}

	return nil
}

// GetNodeMultiplier 获取节点倍率配置
func (r *adminSystemRepo) GetNodeMultiplier(ctx context.Context) (string, error) {
	// Query node_multiplier config from database
	system, err := r.data.db.ProxySystem.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.And(
				sql.EQ(s.C(proxysystem.FieldCategory), "server"),
				sql.EQ(s.C(proxysystem.FieldKey), "NodeMultiplier"),
			))
		}).
		Only(ctx)
	if err != nil {
		// If not found, return empty string (not an error)
		if ent.IsNotFound(err) {
			return "", nil
		}
		r.log.Errorf("[GetNodeMultiplier] Failed to query node multiplier: %v", err)
		return "", responsecode.NewDatabaseQueryError()
	}

	return system.Value, nil
}

// UpdateNodeMultiplier 更新节点倍率配置
func (r *adminSystemRepo) UpdateNodeMultiplier(ctx context.Context, value string) error {
	// Check if config exists
	exists, err := r.data.db.ProxySystem.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.And(
				sql.EQ(s.C(proxysystem.FieldCategory), "server"),
				sql.EQ(s.C(proxysystem.FieldKey), "NodeMultiplier"),
			))
		}).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("[UpdateNodeMultiplier] Failed to check node multiplier existence: %v", err)
		return responsecode.NewDatabaseQueryError()
	}

	if exists {
		// Update existing config
		affected, err := r.data.db.ProxySystem.Update().
			Where(func(s *sql.Selector) {
				s.Where(sql.And(
					sql.EQ(s.C(proxysystem.FieldCategory), "server"),
					sql.EQ(s.C(proxysystem.FieldKey), "NodeMultiplier"),
				))
			}).
			SetValue(value).
			Save(ctx)
		if err != nil {
			r.log.Errorf("[UpdateNodeMultiplier] Failed to update node multiplier: %v", err)
			return responsecode.NewDatabaseUpdateError()
		}
		if affected == 0 {
			r.log.Warnf("[UpdateNodeMultiplier] Node multiplier not found")
			return responsecode.NewSystemNotFoundError()
		}
	} else {
		// Create new config
		_, err := r.data.db.ProxySystem.Create().
			SetCategory("server").
			SetKey("NodeMultiplier").
			SetValue(value).
			SetType("string").
			SetDesc("节点倍率配置").
			Save(ctx)
		if err != nil {
			r.log.Errorf("[UpdateNodeMultiplier] Failed to create node multiplier: %v", err)
			return responsecode.NewDatabaseUpdateError()
		}
	}

	return nil
}
