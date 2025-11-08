package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// MigrationUsecase is the migration use case
type MigrationUsecase struct {
	repo MigrationRepo
	log  *log.Helper
}

// NewMigrationUsecase creates a new migration use case
func NewMigrationUsecase(repo MigrationRepo, logger log.Logger) *MigrationUsecase {
	return &MigrationUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// HasMigrateServerNode checks if there's data to migrate
func (uc *MigrationUsecase) HasMigrateServerNode(ctx context.Context) (bool, error) {
	return uc.repo.HasMigrateServerNode(ctx)
}

// MigrateServerNode migrates server and node data
func (uc *MigrationUsecase) MigrateServerNode(ctx context.Context) (uint64, uint64, string, error) {
	return uc.repo.MigrateServerNode(ctx)
}
