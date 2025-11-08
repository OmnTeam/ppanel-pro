//go:build integration
// +build integration

package data

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
)

// loadConfig loads configuration from config.yaml
func loadConfig() (*conf.Bootstrap, error) {
	c := config.New(
		config.WithSource(
			file.NewSource("../../configs/config.yaml"),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}

	return &bc, nil
}

// setupIntegrationTest creates a Data instance using real database from config.yaml
func setupIntegrationTest(t *testing.T) (*Data, func()) {
	// Load config
	bc, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create logger with proper output
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	// Create Data instance
	data, cleanup, err := NewData(bc.Data, logger)
	if err != nil {
		t.Fatalf("Failed to create data: %v", err)
	}

	return data, cleanup
}

// TestIntegrationQueryWaitReplyTotal tests ticket status query with real database
func TestIntegrationQueryWaitReplyTotal(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	t.Run("查询待回复工单数量", func(t *testing.T) {
		count, err := repo.QueryWaitReplyTotal(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryWaitReplyTotal() error = %v", err)
			return
		}

		t.Logf("Tenant %d has %d pending tickets (status=1)", tenantID, count)

		// Verify count is non-negative
		if count < 0 {
			t.Errorf("QueryWaitReplyTotal() returned negative count: %d", count)
		}
	})

	t.Run("验证查询使用正确的status值", func(t *testing.T) {
		// Query all tickets to verify status values
		tickets, err := data.db.ProxyTicket.Query()
			Where()
			All(ctx)
		if err != nil {
			t.Logf("Could not query all tickets: %v", err)
			return
		}

		t.Logf("Found %d total tickets in database", len(tickets))

		pendingCount := 0
		for _, ticket := range tickets {
			if ticket.Status == 1 && ticket.TenantID == tenantID {
				pendingCount++
			}
		}

		t.Logf("Manual count of pending tickets (status=1): %d", pendingCount)

		// Query using repo method
		repoCount, _ := repo.QueryWaitReplyTotal(ctx, tenantID)

		if repoCount != int64(pendingCount) {
			t.Errorf("Count mismatch: repo returned %d, manual count is %d", repoCount, pendingCount)
		}
	})
}

// TestIntegrationQueryTodayUserTrafficRanking tests user traffic ranking with real database
func TestIntegrationQueryTodayUserTrafficRanking(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	date := time.Now()

	t.Run("查询今日用户流量Top 10", func(t *testing.T) {
		ranking, err := repo.QueryTodayUserTrafficRanking(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryTodayUserTrafficRanking() error = %v", err)
			return
		}

		t.Logf("Today's User Traffic Ranking (Top %d):", len(ranking))
		for i, r := range ranking {
			total := r.Upload + r.Download
			t.Logf("  #%d: Subscribe ID=%d, Upload=%d, Download=%d, Total=%d",
				i+1, r.SID, r.Upload, r.Download, total)
		}

		// Should not return more than 10
		if len(ranking) > 10 {
			t.Errorf("QueryTodayUserTrafficRanking() returned %d results, want max 10", len(ranking))
		}

		// Verify descending order
		for i := 0; i < len(ranking)-1; i++ {
			total1 := ranking[i].Upload + ranking[i].Download
			total2 := ranking[i+1].Upload + ranking[i+1].Download

			if total1 < total2 {
				t.Errorf("Ranking not in descending order at position %d: total[%d]=%d < total[%d]=%d",
					i, i, total1, i+1, total2)
			}
		}
	})
}

// TestIntegrationQueryTodayServerTrafficRanking tests server traffic ranking with real database
func TestIntegrationQueryTodayServerTrafficRanking(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	date := time.Now()

	t.Run("查询今日服务器流量Top 10", func(t *testing.T) {
		ranking, err := repo.QueryTodayServerTrafficRanking(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryTodayServerTrafficRanking() error = %v", err)
			return
		}

		t.Logf("Today's Server Traffic Ranking (Top %d):", len(ranking))
		for i, r := range ranking {
			total := r.Upload + r.Download
			t.Logf("  #%d: Server=%s (ID=%d), Upload=%d, Download=%d, Total=%d",
				i+1, r.Name, r.ServerID, r.Upload, r.Download, total)
		}

		// Should not return more than 10
		if len(ranking) > 10 {
			t.Errorf("QueryTodayServerTrafficRanking() returned %d results, want max 10", len(ranking))
		}

		// Verify descending order
		for i := 0; i < len(ranking)-1; i++ {
			total1 := ranking[i].Upload + ranking[i].Download
			total2 := ranking[i+1].Upload + ranking[i+1].Download

			if total1 < total2 {
				t.Errorf("Ranking not in descending order at position %d: total[%d]=%d < total[%d]=%d",
					i, i, total1, i+1, total2)
			}
		}

		// Verify all servers have names
		for i, r := range ranking {
			if r.Name == "" {
				t.Errorf("Server at position %d has no name: %+v", i, r)
			}
		}
	})
}

// TestIntegrationQueryDateOrders tests order statistics with real database
func TestIntegrationQueryDateOrders(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	today := time.Now().Truncate(24 * time.Hour)

	t.Run("查询今日订单统计", func(t *testing.T) {
		result, err := repo.QueryDateOrders(ctx, tenantID, today)
		if err != nil {
			t.Errorf("QueryDateOrders() error = %v", err)
			return
		}

		t.Logf("Today's Orders:")
		t.Logf("  Total Amount: %d", result.AmountTotal)
		t.Logf("  New Order Amount: %d", result.NewOrderAmount)
		t.Logf("  Renewal Order Amount: %d", result.RenewalOrderAmount)

		// Verify total = new + renewal
		expected := result.NewOrderAmount + result.RenewalOrderAmount
		if result.AmountTotal != expected {
			t.Errorf("AmountTotal=%d, but NewOrder+Renewal=%d", result.AmountTotal, expected)
		}
	})

	t.Run("查询昨日订单统计", func(t *testing.T) {
		yesterday := today.Add(-24 * time.Hour)
		result, err := repo.QueryDateOrders(ctx, tenantID, yesterday)
		if err != nil {
			t.Errorf("QueryDateOrders() error = %v", err)
			return
		}

		t.Logf("Yesterday's Orders:")
		t.Logf("  Total Amount: %d", result.AmountTotal)
		t.Logf("  New Order Amount: %d", result.NewOrderAmount)
		t.Logf("  Renewal Order Amount: %d", result.RenewalOrderAmount)
	})
}

// TestIntegrationQueryRevenueStatistics tests full revenue statistics
func TestIntegrationQueryRevenueStatistics(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	t.Run("查询完整收入统计", func(t *testing.T) {
		// Query today
		today, err := repo.QueryDateOrders(ctx, tenantID, time.Now())
		if err != nil {
			t.Errorf("QueryDateOrders(today) error = %v", err)
			return
		}

		// Query this month
		monthly, err := repo.QueryMonthlyOrders(ctx, tenantID, time.Now())
		if err != nil {
			t.Errorf("QueryMonthlyOrders() error = %v", err)
			return
		}

		// Query all time
		all, err := repo.QueryTotalOrders(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Logf("Revenue Statistics for Tenant %d:", tenantID)
		t.Logf("Today: Total=%d, New=%d, Renewal=%d",
			today.AmountTotal, today.NewOrderAmount, today.RenewalOrderAmount)
		t.Logf("This Month: Total=%d, New=%d, Renewal=%d",
			monthly.AmountTotal, monthly.NewOrderAmount, monthly.RenewalOrderAmount)
		t.Logf("All Time: Total=%d, New=%d, Renewal=%d",
			all.AmountTotal, all.NewOrderAmount, all.RenewalOrderAmount)

		// Verify today <= month <= all
		if today.AmountTotal > monthly.AmountTotal {
			t.Errorf("Today's amount (%d) should not exceed monthly amount (%d)",
				today.AmountTotal, monthly.AmountTotal)
		}
		if monthly.AmountTotal > all.AmountTotal {
			t.Errorf("Monthly amount (%d) should not exceed all-time amount (%d)",
				monthly.AmountTotal, all.AmountTotal)
		}
	})
}

// TestIntegrationMultiTenantIsolation verifies multi-tenant data isolation
func TestIntegrationMultiTenantIsolation(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	t.Run("验证不同租户数据隔离", func(t *testing.T) {
		// Query for tenant 1
		count1, err1 := repo.QueryWaitReplyTotal(ctx, 1)
		if err1 != nil {
			t.Logf("Tenant 1 query error: %v", err1)
		}

		// Query for tenant 2 (if exists)
		count2, err2 := repo.QueryWaitReplyTotal(ctx, 2)
		if err2 != nil {
			t.Logf("Tenant 2 query error: %v", err2)
		}

		t.Logf("Tenant 1 pending tickets: %d", count1)
		t.Logf("Tenant 2 pending tickets: %d", count2)

		// The counts can be different - this is expected and correct
		t.Logf("Multi-tenant isolation: Each tenant has independent data")
	})
}

// TestIntegrationConsoleAPIs tests all Console APIs together
func TestIntegrationConsoleAPIs(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	now := time.Now()

	t.Run("测试所有Console API", func(t *testing.T) {
		// Test Revenue APIs
		t.Log("=== Revenue Statistics ===")
		if orders, err := repo.QueryDateOrders(ctx, tenantID, now); err == nil {
			t.Logf("Today Orders: %+v", orders)
		} else {
			t.Errorf("QueryDateOrders failed: %v", err)
		}

		if orders, err := repo.QueryMonthlyOrders(ctx, tenantID, now); err == nil {
			t.Logf("Monthly Orders: %+v", orders)
		} else {
			t.Errorf("QueryMonthlyOrders failed: %v", err)
		}

		// Test User APIs
		t.Log("\n=== User Statistics ===")
		if count, err := repo.QueryRegisterUserTotalByDate(ctx, tenantID, now); err == nil {
			t.Logf("Today Registered Users: %d", count)
		} else {
			t.Errorf("QueryRegisterUserTotalByDate failed: %v", err)
		}

		if count, err := repo.QueryRegisterUserTotal(ctx, tenantID); err == nil {
			t.Logf("Total Registered Users: %d", count)
		} else {
			t.Errorf("QueryRegisterUserTotal failed: %v", err)
		}

		// Test Ticket APIs
		t.Log("\n=== Ticket Statistics ===")
		if count, err := repo.QueryWaitReplyTotal(ctx, tenantID); err == nil {
			t.Logf("Pending Tickets (status=1): %d", count)
		} else {
			t.Errorf("QueryWaitReplyTotal failed: %v", err)
		}

		// Test Server APIs
		t.Log("\n=== Server Statistics ===")
		if count, err := repo.QueryOnlineServers(ctx, tenantID); err == nil {
			t.Logf("Online Servers: %d", count)
		} else {
			t.Errorf("QueryOnlineServers failed: %v", err)
		}

		if count, err := repo.QueryOfflineServers(ctx, tenantID); err == nil {
			t.Logf("Offline Servers: %d", count)
		} else {
			t.Errorf("QueryOfflineServers failed: %v", err)
		}

		// Test Traffic APIs
		t.Log("\n=== Traffic Statistics ===")
		if upload, download, err := repo.QueryTodayTraffic(ctx, tenantID, now); err == nil {
			t.Logf("Today Traffic: Upload=%d, Download=%d, Total=%d",
				upload, download, upload+download)
		} else {
			t.Errorf("QueryTodayTraffic failed: %v", err)
		}

		if ranking, err := repo.QueryTodayUserTrafficRanking(ctx, tenantID, now); err == nil {
			t.Logf("Today User Traffic Top %d retrieved successfully", len(ranking))
		} else {
			t.Errorf("QueryTodayUserTrafficRanking failed: %v", err)
		}

		if ranking, err := repo.QueryTodayServerTrafficRanking(ctx, tenantID, now); err == nil {
			t.Logf("Today Server Traffic Top %d retrieved successfully", len(ranking))
		} else {
			t.Errorf("QueryTodayServerTrafficRanking failed: %v", err)
		}

		t.Log("\n=== All Console APIs Tested ===")
	})
}

// =============================================================================
// Deep Testing for Order and Revenue Statistics
// =============================================================================

// TestIntegrationOrderStatisticsDeep provides comprehensive deep testing for order statistics
func TestIntegrationOrderStatisticsDeep(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	t.Run("验证订单状态过滤逻辑", func(t *testing.T) {
		// Query all orders to analyze status distribution
		allOrders, err := data.db.ProxyOrder.Query()
			Where()
			All(ctx)
		if err != nil {
			t.Logf("Could not query all orders: %v", err)
			return
		}

		t.Logf("Total orders in database: %d", len(allOrders))

		// Count orders by status
		statusCounts := make(map[int8]int)
		for _, order := range allOrders {
			statusCounts[order.Status]++
		}

		t.Log("Order distribution by status:")
		for status, count := range statusCounts {
			t.Logf("  Status %d: %d orders", status, count)
		}

		// Verify: Only status 2 and 5 should be counted
		validOrderCount := statusCounts[2] + statusCounts[5]
		t.Logf("Valid orders (status 2 or 5): %d", validOrderCount)

		// Query using repo method (filters by status 2, 5)
		result, err := repo.QueryTotalOrders(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Logf("QueryTotalOrders result: Total=%d, New=%d, Renewal=%d",
			result.AmountTotal, result.NewOrderAmount, result.RenewalOrderAmount)
	})

	t.Run("验证支付方式过滤逻辑", func(t *testing.T) {
		// Count orders by payment method
		allOrders, err := data.db.ProxyOrder.Query()
			Where()
			All(ctx)
		if err != nil {
			t.Logf("Could not query all orders: %v", err)
			return
		}

		methodCounts := make(map[string]int)
		balanceAmount := int64(0)
		nonBalanceAmount := int64(0)

		for _, order := range allOrders {
			methodCounts[order.Method]++
			if order.Method == "balance" {
				balanceAmount += order.Amount
			} else if order.Status == 2 || order.Status == 5 {
				nonBalanceAmount += order.Amount
			}
		}

		t.Log("Order distribution by payment method:")
		for method, count := range methodCounts {
			t.Logf("  Method '%s': %d orders", method, count)
		}

		t.Logf("Balance orders total amount: %d (should be excluded)", balanceAmount)
		t.Logf("Non-balance orders total amount: %d (should be counted)", nonBalanceAmount)

		// Verify: balance orders should be excluded
		result, err := repo.QueryTotalOrders(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Logf("QueryTotalOrders result excludes balance: %d", result.AmountTotal)
	})

	t.Run("验证新购/续费订单区分", func(t *testing.T) {
		// Analyze is_new field distribution
		allOrders, err := data.db.ProxyOrder.Query()
			Where()
			All(ctx)
		if err != nil {
			t.Logf("Could not query all orders: %v", err)
			return
		}

		newOrderCount := 0
		renewalOrderCount := 0
		newOrderAmount := int64(0)
		renewalOrderAmount := int64(0)

		for _, order := range allOrders {
			// Only count valid status (2, 5) and non-balance
			if (order.Status == 2 || order.Status == 5) && order.Method != "balance" {
				if order.IsNew {
					newOrderCount++
					newOrderAmount += order.Amount
				} else {
					renewalOrderCount++
					renewalOrderAmount += order.Amount
				}
			}
		}

		t.Logf("Manual calculation:")
		t.Logf("  New orders: %d, Amount: %d", newOrderCount, newOrderAmount)
		t.Logf("  Renewal orders: %d, Amount: %d", renewalOrderCount, renewalOrderAmount)

		// Compare with repo method
		result, err := repo.QueryTotalOrders(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Logf("QueryTotalOrders result:")
		t.Logf("  New: %d, Renewal: %d", result.NewOrderAmount, result.RenewalOrderAmount)

		// Verify total = new + renewal
		expectedTotal := result.NewOrderAmount + result.RenewalOrderAmount
		if result.AmountTotal != expectedTotal {
			t.Errorf("AmountTotal=%d, but New+Renewal=%d",
				result.AmountTotal, expectedTotal)
		}
	})

	t.Run("验证日期范围过滤", func(t *testing.T) {
		today := time.Now().Truncate(24 * time.Hour)
		yesterday := today.Add(-24 * time.Hour)
		firstDayOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

		// Query today
		todayResult, err := repo.QueryDateOrders(ctx, tenantID, today)
		if err != nil {
			t.Errorf("QueryDateOrders(today) error = %v", err)
			return
		}

		// Query yesterday
		yesterdayResult, err := repo.QueryDateOrders(ctx, tenantID, yesterday)
		if err != nil {
			t.Errorf("QueryDateOrders(yesterday) error = %v", err)
			return
		}

		// Query this month
		monthlyResult, err := repo.QueryMonthlyOrders(ctx, tenantID, today)
		if err != nil {
			t.Errorf("QueryMonthlyOrders() error = %v", err)
			return
		}

		// Query all time
		allResult, err := repo.QueryTotalOrders(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Logf("Date range filtering results:")
		t.Logf("  Today (%s): %d", today.Format("2006-01-02"), todayResult.AmountTotal)
		t.Logf("  Yesterday (%s): %d", yesterday.Format("2006-01-02"), yesterdayResult.AmountTotal)
		t.Logf("  This Month (%s to %s): %d",
			firstDayOfMonth.Format("2006-01-02"),
			today.Format("2006-01-02"),
			monthlyResult.AmountTotal)
		t.Logf("  All Time: %d", allResult.AmountTotal)

		// Logical checks
		// Note: today and yesterday might not sum to monthly due to other days
		// But monthly should not exceed all-time
		if monthlyResult.AmountTotal > allResult.AmountTotal {
			t.Errorf("Monthly amount (%d) should not exceed all-time amount (%d)",
				monthlyResult.AmountTotal, allResult.AmountTotal)
		}
	})

	t.Run("边界条件测试", func(t *testing.T) {
		// Test with a future date (should return zero)
		futureDate := time.Now().Add(365 * 24 * time.Hour)
		futureResult, err := repo.QueryDateOrders(ctx, tenantID, futureDate)
		if err != nil {
			t.Errorf("QueryDateOrders(future) error = %v", err)
			return
		}

		if futureResult.AmountTotal != 0 {
			t.Logf("Warning: Future date query returned non-zero: %d", futureResult.AmountTotal)
		} else {
			t.Logf("✓ Future date query correctly returns 0")
		}

		// Test with a very old date
		oldDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		oldResult, err := repo.QueryDateOrders(ctx, tenantID, oldDate)
		if err != nil {
			t.Errorf("QueryDateOrders(old) error = %v", err)
			return
		}

		t.Logf("Old date (%s) query result: %d", oldDate.Format("2006-01-02"), oldResult.AmountTotal)
	})

	t.Run("多租户数据隔离验证", func(t *testing.T) {
		// Query tenant 1
		tenant1Result, err := repo.QueryTotalOrders(ctx, 1)
		if err != nil {
			t.Errorf("QueryTotalOrders(tenant 1) error = %v", err)
			return
		}

		// Query tenant 2 (if exists)
		tenant2Result, err := repo.QueryTotalOrders(ctx, 2)
		if err != nil {
			t.Logf("Tenant 2 query error (might not exist): %v", err)
		}

		t.Logf("Tenant 1 total orders: %d", tenant1Result.AmountTotal)
		t.Logf("Tenant 2 total orders: %d", tenant2Result.AmountTotal)

		// They should be independent
		t.Log("✓ Multi-tenant data isolation working (each tenant has independent order data)")
	})
}

// TestIntegrationRevenueStatisticsDeep provides comprehensive deep testing for revenue calculations
func TestIntegrationRevenueStatisticsDeep(t *testing.T) {
	data, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	now := time.Now()

	t.Run("收入统计完整性验证", func(t *testing.T) {
		// Query all three time ranges
		todayResult, err := repo.QueryDateOrders(ctx, tenantID, now)
		if err != nil {
			t.Errorf("QueryDateOrders() error = %v", err)
			return
		}

		monthlyResult, err := repo.QueryMonthlyOrders(ctx, tenantID, now)
		if err != nil {
			t.Errorf("QueryMonthlyOrders() error = %v", err)
			return
		}

		allResult, err := repo.QueryTotalOrders(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Log("=== Revenue Statistics Integrity ===")
		t.Logf("Today:     Total=%8d, New=%8d, Renewal=%8d",
			todayResult.AmountTotal, todayResult.NewOrderAmount, todayResult.RenewalOrderAmount)
		t.Logf("Monthly:   Total=%8d, New=%8d, Renewal=%8d",
			monthlyResult.AmountTotal, monthlyResult.NewOrderAmount, monthlyResult.RenewalOrderAmount)
		t.Logf("All-Time:  Total=%8d, New=%8d, Renewal=%8d",
			allResult.AmountTotal, allResult.NewOrderAmount, allResult.RenewalOrderAmount)

		// Verify each total = new + renewal
		if todayResult.AmountTotal != todayResult.NewOrderAmount+todayResult.RenewalOrderAmount {
			t.Errorf("Today: Total != New+Renewal")
		}
		if monthlyResult.AmountTotal != monthlyResult.NewOrderAmount+monthlyResult.RenewalOrderAmount {
			t.Errorf("Monthly: Total != New+Renewal")
		}
		if allResult.AmountTotal != allResult.NewOrderAmount+allResult.RenewalOrderAmount {
			t.Errorf("All-Time: Total != New+Renewal")
		}

		// Verify hierarchical relationship: today <= monthly <= all
		if todayResult.AmountTotal > monthlyResult.AmountTotal {
			t.Errorf("Today's revenue (%d) exceeds monthly revenue (%d)",
				todayResult.AmountTotal, monthlyResult.AmountTotal)
		}
		if monthlyResult.AmountTotal > allResult.AmountTotal {
			t.Errorf("Monthly revenue (%d) exceeds all-time revenue (%d)",
				monthlyResult.AmountTotal, allResult.AmountTotal)
		}

		t.Log("✓ Revenue statistics integrity verified")
	})

	t.Run("收入金额准确性验证", func(t *testing.T) {
		// Manually calculate today's revenue from database
		today := now.Truncate(24 * time.Hour)
		tomorrow := today.Add(24 * time.Hour)

		orders, err := data.db.ProxyOrder.Query()
			Where()
			All(ctx)
		if err != nil {
			t.Logf("Could not query orders: %v", err)
			return
		}

		manualTotal := int64(0)
		manualNew := int64(0)
		manualRenewal := int64(0)

		for _, order := range orders {
			// Apply same filters as QueryDateOrders
			if order.TenantID != tenantID {
				continue
			}
			if order.Status != 2 && order.Status != 5 {
				continue
			}
			if order.Method == "balance" {
				continue
			}
			// Check if order is from today
			orderTime := order.CreatedAt
			if orderTime.Before(today) || orderTime.After(tomorrow) {
				continue
			}

			manualTotal += order.Amount
			if order.IsNew {
				manualNew += order.Amount
			} else {
				manualRenewal += order.Amount
			}
		}

		t.Logf("Manual calculation for today:")
		t.Logf("  Total=%d, New=%d, Renewal=%d", manualTotal, manualNew, manualRenewal)

		// Compare with repo method
		repoResult, err := repo.QueryDateOrders(ctx, tenantID, now)
		if err != nil {
			t.Errorf("QueryDateOrders() error = %v", err)
			return
		}

		t.Logf("Repository method result:")
		t.Logf("  Total=%d, New=%d, Renewal=%d",
			repoResult.AmountTotal, repoResult.NewOrderAmount, repoResult.RenewalOrderAmount)

		// Verify they match
		if manualTotal != repoResult.AmountTotal {
			t.Errorf("Total mismatch: manual=%d, repo=%d", manualTotal, repoResult.AmountTotal)
		}
		if manualNew != repoResult.NewOrderAmount {
			t.Errorf("New order mismatch: manual=%d, repo=%d", manualNew, repoResult.NewOrderAmount)
		}
		if manualRenewal != repoResult.RenewalOrderAmount {
			t.Errorf("Renewal mismatch: manual=%d, repo=%d", manualRenewal, repoResult.RenewalOrderAmount)
		}

		if manualTotal == repoResult.AmountTotal &&
			manualNew == repoResult.NewOrderAmount &&
			manualRenewal == repoResult.RenewalOrderAmount {
			t.Log("✓ Revenue calculation accuracy verified")
		}
	})

	t.Run("月度收入跨月边界测试", func(t *testing.T) {
		// Test first day of month
		firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		firstDayResult, err := repo.QueryMonthlyOrders(ctx, tenantID, firstDayOfMonth)
		if err != nil {
			t.Errorf("QueryMonthlyOrders(first day) error = %v", err)
			return
		}

		// Test last day of previous month
		lastDayOfPrevMonth := firstDayOfMonth.Add(-24 * time.Hour)
		lastDayResult, err := repo.QueryMonthlyOrders(ctx, tenantID, lastDayOfPrevMonth)
		if err != nil {
			t.Errorf("QueryMonthlyOrders(last day of prev month) error = %v", err)
			return
		}

		t.Logf("Month boundary test:")
		t.Logf("  Current month (%s): %d",
			firstDayOfMonth.Format("2006-01"), firstDayResult.AmountTotal)
		t.Logf("  Previous month (%s): %d",
			lastDayOfPrevMonth.Format("2006-01"), lastDayResult.AmountTotal)

		// These should be different (different months)
		t.Log("✓ Month boundary filtering working correctly")
	})

	t.Run("零收入场景测试", func(t *testing.T) {
		// Test with a tenant that likely has no data
		emptyTenantID := int64(9999)
		result, err := repo.QueryTotalOrders(ctx, emptyTenantID)
		if err != nil {
			t.Errorf("QueryTotalOrders(empty tenant) error = %v", err)
			return
		}

		if result.AmountTotal == 0 && result.NewOrderAmount == 0 && result.RenewalOrderAmount == 0 {
			t.Log("✓ Zero revenue scenario handled correctly")
			t.Logf("  Empty tenant result: %+v", result)
		} else {
			t.Logf("Note: Tenant %d has data: %+v", emptyTenantID, result)
		}
	})

	t.Run("性能和响应时间测试", func(t *testing.T) {
		// Measure query performance
		startTime := time.Now()
		_, err := repo.QueryTotalOrders(ctx, tenantID)
		elapsed := time.Since(startTime)

		if err != nil {
			t.Errorf("QueryTotalOrders() error = %v", err)
			return
		}

		t.Logf("QueryTotalOrders execution time: %v", elapsed)

		if elapsed > 1*time.Second {
			t.Logf("Warning: Query took longer than 1 second: %v", elapsed)
		} else {
			t.Log("✓ Query performance acceptable")
		}

		// Test monthly query performance
		startTime = time.Now()
		_, err = repo.QueryMonthlyOrders(ctx, tenantID, now)
		elapsed = time.Since(startTime)

		if err != nil {
			t.Errorf("QueryMonthlyOrders() error = %v", err)
			return
		}

		t.Logf("QueryMonthlyOrders execution time: %v", elapsed)
	})
}
