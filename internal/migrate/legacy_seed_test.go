package migrate

import "testing"

func TestNormalizeLegacyStatementSkipsObsoleteSubscribeTypeSeed(t *testing.T) {
	stmt := "INSERT IGNORE INTO `subscribe_type` (`id`, `name`) VALUES (1, 'Clash')"
	if got := normalizeLegacyStatement("legacy_sql/00002_init_basic_data.up.sql", stmt); got != "" {
		t.Fatalf("expected obsolete subscribe_type seed to be skipped, got %q", got)
	}
}

func TestNormalizeLegacyStatementUsesInsertIgnoreForSubscribeApplication(t *testing.T) {
	stmt := "INSERT INTO `subscribe_application` (`id`, `name`) VALUES (1, 'Default')"
	want := "INSERT IGNORE INTO `subscribe_application` (`id`, `name`) VALUES (1, 'Default')"
	if got := normalizeLegacyStatement("legacy_sql/02101_subscribe_application.up.sql", stmt); got != want {
		t.Fatalf("unexpected normalized statement: got %q want %q", got, want)
	}
}

func TestLegacySQLMigrationsIncludeSyncedLatestVersions(t *testing.T) {
	want := []int64{2133, 2134, 2135, 2136, 2137}
	got := make([]int64, 0, len(want))
	for _, migration := range legacySQLMigrations {
		if migration.version >= 2133 {
			got = append(got, migration.version)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected synced migration count: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected synced migration versions: got %v want %v", got, want)
		}
	}
}
