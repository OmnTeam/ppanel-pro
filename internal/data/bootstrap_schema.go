package data

import (
	"context"
	"database/sql"
	"fmt"
)

func hasExistingLegacySchema(ctx context.Context, driverName, dataSourceName string) (bool, error) {
	// Once a core old-project table exists, startup must not let Ent rewrite the schema.
	for _, tableName := range []string{"user_subscribe", "user", "system"} {
		exists, err := hasDatabaseTable(ctx, driverName, dataSourceName, tableName)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func hasDatabaseTable(ctx context.Context, driverName, dataSourceName, tableName string) (bool, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return false, fmt.Errorf("open sql db: %w", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`,
		tableName,
	).Scan(&count); err != nil {
		return false, fmt.Errorf("query table %s existence failed: %w", tableName, err)
	}

	return count > 0, nil
}
