package data

import (
	"context"

	serverbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/server"
	"github.com/go-kratos/kratos/v2/log"
)

type adminMigrationRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminMigrationRepo creates a new admin migration repository
func NewAdminMigrationRepo(data *Data, logger log.Logger) serverbiz.MigrationRepo {
	return &adminMigrationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// HasMigrateServerNode checks if there's any data to migrate
// 新项目不需要迁移，直接返回 false
func (r *adminMigrationRepo) HasMigrateServerNode(ctx context.Context) (bool, error) {
	// 新项目不需要迁移数据，始终返回 false
	r.log.Infof("[HasMigrateServerNode] New project, no migration needed")
	return false, nil
}

// MigrateServerNode migrates server and node data
func (r *adminMigrationRepo) MigrateServerNode(ctx context.Context) (uint64, uint64, string, error) {
	// This would perform actual migration from old tables
	// For now, return empty result
	return 0, 0, "Migration not implemented", nil
}
