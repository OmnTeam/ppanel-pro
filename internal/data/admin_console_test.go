package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/enttest"
	"github.com/OmnTeam/ppanel-pro/internal/model"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/mattn/go-sqlite3"
)

// setupConsoleTestClient creates a test Ent client with SQLite in-memory database
func setupConsoleTestClient(t *testing.T) *ent.Client {
	opts := []enttest.Option{
		enttest.WithOptions(
			ent.Log(t.Log),
			ent.Debug(), // Enable debug mode to print all SQL
		),
	}
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1", opts...)
	return client
}

// seedConsoleTestData prepares test data for Console module
func seedConsoleTestData(t *testing.T, client *ent.Client, ctx context.Context) {
	tenantID := int64(1)

	// Create users
	users := make([]*ent.ProxyUser, 10)
	for i := 0; i < 10; i++ {
		user, err := client.ProxyUser.Create()
			SetTenantID(tenantID)
			SetPassword("password")
			SetEnable(true)
			SetIsAdmin(false)
			SetBalance(0)
			SetCreatedAt(time.Now().Add(-time.Duration(i) * 24 * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
		users[i] = user
	}

	// Create subscribe plans
	subscribe, err := client.ProxySubscribe.Create()
		SetTenantID(tenantID)
		SetName("Test Subscribe Plan")
		SetTraffic(10000000000). // 10GB
		SetShow(true)
		SetSell(true)
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create subscribe: %v", err)
	}

	// Create orders
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	// Today's orders
	for i := 0; i < 5; i++ {
		isNew := i < 3 // First 3 are new orders
		status := int8(2)
		if i == 4 {
			status = int8(5) // One completed order
		}

		orderNo := fmt.Sprintf("ORDER-TODAY-%d-%d", tenantID, i)
		_, err := client.ProxyOrder.Create()
			SetTenantID(tenantID)
			SetUserID(users[i].ID)
			SetSubscribeID(subscribe.ID)
			SetOrderNo(orderNo)
			SetIsNew(isNew)
			SetAmount(int64(100 * (i + 1))). // 100, 200, 300, 400, 500
			SetStatus(status)
			SetMethod("alipay")
			SetCreatedAt(today.Add(time.Duration(i) * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create today's order %d: %v", i, err)
		}
	}

	// Yesterday's orders
	for i := 0; i < 3; i++ {
		orderNo := fmt.Sprintf("ORDER-YESTERDAY-%d-%d", tenantID, i)
		_, err := client.ProxyOrder.Create()
			SetTenantID(tenantID)
			SetUserID(users[i].ID)
			SetSubscribeID(subscribe.ID)
			SetOrderNo(orderNo)
			SetIsNew(i == 0)
			SetAmount(int64(150 * (i + 1))). // 150, 300, 450
			SetStatus(2)
			SetMethod("alipay")
			SetCreatedAt(yesterday.Add(time.Duration(i) * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create yesterday's order %d: %v", i, err)
		}
	}

	// Create tickets with different statuses
	ticketStatuses := []int8{1, 1, 1, 2, 3, 4} // 3 Pending, 1 Waiting, 1 Processed, 1 Closed
	for i, status := range ticketStatuses {
		_, err := client.ProxyTicket.Create()
			SetTenantID(tenantID)
			SetUserID(users[i].ID)
			SetTitle("Test Ticket " + string(rune('A'+i)))
			SetDescription("Test Description")
			SetStatus(status)
			SetCreatedAt(now.Add(-time.Duration(i) * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create ticket %d: %v", i, err)
		}
	}

	// Create servers
	servers := make([]*ent.ProxyServer, 5)
	for i := 0; i < 5; i++ {
		server, err := client.ProxyServer.Create()
			SetTenantID(tenantID)
			SetName("Server-" + string(rune('A'+i)))
			SetTags("tag1,tag2")
			SetCountry("US")
			SetCity("New York")
			SetServerAddr("server" + string(rune('a'+i)) + ".example.com:443")
			SetProtocol("vmess")
			SetEnable(true)
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create server %d: %v", i, err)
		}
		servers[i] = server
	}

	// Create user subscribes
	userSubscribes := make([]*ent.ProxyUserSubscribe, 10)
	for i := 0; i < 10; i++ {
		userSub, err := client.ProxyUserSubscribe.Create()
			SetTenantID(tenantID)
			SetUserID(users[i].ID)
			SetOrderID(int64(1000 + i))
			SetSubscribeID(subscribe.ID)
			SetStatus(uint8(model.UserSubscribeStatusActive))
			SetStartTime(now.Add(-10 * 24 * time.Hour))
			SetExpireTime(now.Add(20 * 24 * time.Hour))
			SetTraffic(10000000000)
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create user subscribe %d: %v", i, err)
		}
		userSubscribes[i] = userSub
	}

	// Create traffic logs for today (with different traffic amounts to test ranking)
	todayStart := today

	// Create traffic with descending order: User 9 has most traffic, User 0 has least
	// This tests if our sorting is working correctly
	for i := 0; i < 15; i++ { // Create 15 traffic records to test Top 10
		userIndex := i % 10
		upload := int64((15 - i) * 1000000)   // 15MB, 14MB, ..., 1MB
		download := int64((15 - i) * 2000000) // 30MB, 28MB, ..., 2MB

		_, err := client.ProxyTrafficLog.Create()
			SetTenantID(tenantID)
			SetUserID(users[userIndex].ID)
			SetSubscribeID(userSubscribes[userIndex].ID)
			SetServerID(servers[i%5].ID)
			SetUpload(upload)
			SetDownload(download)
			SetTimestamp(todayStart.Add(time.Duration(i) * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create traffic log %d: %v", i, err)
		}
	}
}

// TestQueryWaitReplyTotal tests the fixed ticket status code (status=1 for Pending)
func TestQueryWaitReplyTotal(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedConsoleTestData(t, client, ctx)

	data := &Data{db: client}
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	t.Run("应该返回status=1(Pending)的工单数量", func(t *testing.T) {
		count, err := repo.QueryWaitReplyTotal(ctx, tenantID)
		if err != nil {
			t.Errorf("QueryWaitReplyTotal() error = %v", err)
			return
		}

		// We created 3 tickets with status=1 (Pending)
		expectedCount := int64(3)
		if count != expectedCount {
			t.Errorf("QueryWaitReplyTotal() = %v, want %v", count, expectedCount)
		}
	})

	t.Run("验证不会统计其他状态的工单", func(t *testing.T) {
		// Verify we're not counting status=0, 2, 3, 4
		allTickets, err := client.ProxyTicket.Query().Where().All(ctx)
		if err != nil {
			t.Fatalf("Failed to query all tickets: %v", err)
		}

		t.Logf("Total tickets created: %d", len(allTickets))
		for _, ticket := range allTickets {
			t.Logf("Ticket ID=%d, Status=%d", ticket.ID, ticket.Status)
		}

		// Verify only status=1 tickets are counted
		count, _ := repo.QueryWaitReplyTotal(ctx, tenantID)
		if count > 3 {
			t.Errorf("QueryWaitReplyTotal() counted more than Pending tickets: got %v", count)
		}
	})
}

// TestQueryTodayUserTrafficRanking tests the user traffic ranking with sorting
func TestQueryTodayUserTrafficRanking(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedConsoleTestData(t, client, ctx)

	data := &Data{db: client}
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	date := time.Now()

	t.Run("应该返回Top 10并按流量降序排序", func(t *testing.T) {
		ranking, err := repo.QueryTodayUserTrafficRanking(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryTodayUserTrafficRanking() error = %v", err)
			return
		}

		// Should return max 10 results
		if len(ranking) > 10 {
			t.Errorf("QueryTodayUserTrafficRanking() returned %d results, want max 10", len(ranking))
		}

		// Verify descending order by total traffic
		for i := 0; i < len(ranking)-1; i++ {
			total1 := ranking[i].Upload + ranking[i].Download
			total2 := ranking[i+1].Upload + ranking[i+1].Download

			if total1 < total2 {
				t.Errorf("Ranking not in descending order at position %d: total[%d]=%d < total[%d]=%d",
					i, i, total1, i+1, total2)
			}
		}

		t.Logf("Top 10 User Traffic Ranking:")
		for i, r := range ranking {
			total := r.Upload + r.Download
			t.Logf("  #%d: Subscribe ID=%d, Upload=%d, Download=%d, Total=%d",
				i+1, r.SID, r.Upload, r.Download, total)
		}
	})

	t.Run("验证返回的流量数据正确", func(t *testing.T) {
		ranking, err := repo.QueryTodayUserTrafficRanking(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryTodayUserTrafficRanking() error = %v", err)
			return
		}

		// First place should have the highest traffic
		if len(ranking) > 0 {
			firstPlace := ranking[0]
			if firstPlace.Upload == 0 && firstPlace.Download == 0 {
				t.Errorf("First place has zero traffic: %+v", firstPlace)
			}
		}
	})
}

// TestQueryTodayServerTrafficRanking tests the server traffic ranking with sorting
func TestQueryTodayServerTrafficRanking(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedConsoleTestData(t, client, ctx)

	data := &Data{db: client}
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	date := time.Now()

	t.Run("应该返回Top 10并按流量降序排序", func(t *testing.T) {
		ranking, err := repo.QueryTodayServerTrafficRanking(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryTodayServerTrafficRanking() error = %v", err)
			return
		}

		// Should return max 10 results
		if len(ranking) > 10 {
			t.Errorf("QueryTodayServerTrafficRanking() returned %d results, want max 10", len(ranking))
		}

		// Verify descending order by total traffic
		for i := 0; i < len(ranking)-1; i++ {
			total1 := ranking[i].Upload + ranking[i].Download
			total2 := ranking[i+1].Upload + ranking[i+1].Download

			if total1 < total2 {
				t.Errorf("Ranking not in descending order at position %d: total[%d]=%d < total[%d]=%d",
					i, i, total1, i+1, total2)
			}
		}

		t.Logf("Top 10 Server Traffic Ranking:")
		for i, r := range ranking {
			total := r.Upload + r.Download
			t.Logf("  #%d: Server=%s, Upload=%d, Download=%d, Total=%d",
				i+1, r.Name, r.Upload, r.Download, total)
		}
	})

	t.Run("验证服务器名称已填充", func(t *testing.T) {
		ranking, err := repo.QueryTodayServerTrafficRanking(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryTodayServerTrafficRanking() error = %v", err)
			return
		}

		// All servers should have names
		for i, r := range ranking {
			if r.Name == "" {
				t.Errorf("Server at position %d has no name: %+v", i, r)
			}
		}
	})
}

// TestQueryDateOrders tests order statistics
func TestQueryDateOrders(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedConsoleTestData(t, client, ctx)

	data := &Data{db: client}
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

		// We created 5 orders today with status=2 or 5
		// Amounts: 100, 200, 300, 400, 500 = 1500 total
		// First 3 are new orders: 100, 200, 300 = 600
		// Last 2 are renewal: 400, 500 = 900
		if result.AmountTotal != 1500 {
			t.Errorf("AmountTotal = %v, want 1500", result.AmountTotal)
		}
		if result.NewOrderAmount != 600 {
			t.Errorf("NewOrderAmount = %v, want 600", result.NewOrderAmount)
		}
		if result.RenewalOrderAmount != 900 {
			t.Errorf("RenewalOrderAmount = %v, want 900", result.RenewalOrderAmount)
		}

		t.Logf("Today's orders: Total=%d, New=%d, Renewal=%d",
			result.AmountTotal, result.NewOrderAmount, result.RenewalOrderAmount)
	})

	t.Run("查询昨日订单统计", func(t *testing.T) {
		yesterday := today.Add(-24 * time.Hour)
		result, err := repo.QueryDateOrders(ctx, tenantID, yesterday)
		if err != nil {
			t.Errorf("QueryDateOrders() error = %v", err)
			return
		}

		// Yesterday: 150, 300, 450 = 900 total
		// Only first is new: 150
		// Renewal: 300, 450 = 750
		if result.AmountTotal != 900 {
			t.Errorf("AmountTotal = %v, want 900", result.AmountTotal)
		}
		if result.NewOrderAmount != 150 {
			t.Errorf("NewOrderAmount = %v, want 150", result.NewOrderAmount)
		}
		if result.RenewalOrderAmount != 750 {
			t.Errorf("RenewalOrderAmount = %v, want 750", result.RenewalOrderAmount)
		}
	})

	t.Run("验证过滤balance支付方式", func(t *testing.T) {
		// Create an order with method=balance (should be excluded)
		_, err := client.ProxyOrder.Create()
			SetTenantID(tenantID)
			SetUserID(1)
			SetSubscribeID(1)
			SetOrderNo("ORDER-BALANCE-TEST")
			SetIsNew(true)
			SetAmount(10000). // Large amount
			SetStatus(2)
			SetMethod("balance"). // Should be filtered out
			SetCreatedAt(today.Add(12 * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create balance order: %v", err)
		}

		result, err := repo.QueryDateOrders(ctx, tenantID, today)
		if err != nil {
			t.Errorf("QueryDateOrders() error = %v", err)
			return
		}

		// Should still be 1500 (balance order excluded)
		if result.AmountTotal != 1500 {
			t.Errorf("AmountTotal = %v, want 1500 (balance order should be excluded)", result.AmountTotal)
		}
	})
}

// TestQueryDailyUserStatisticsList tests user daily statistics
func TestQueryDailyUserStatisticsList(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedConsoleTestData(t, client, ctx)

	data := &Data{db: client}
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)
	date := time.Now()

	t.Run("查询当月每日用户增长统计", func(t *testing.T) {
		stats, err := repo.QueryDailyUserStatisticsList(ctx, tenantID, date)
		if err != nil {
			t.Errorf("QueryDailyUserStatisticsList() error = %v", err)
			return
		}

		if len(stats) == 0 {
			t.Error("QueryDailyUserStatisticsList() returned empty list")
			return
		}

		t.Logf("Daily user statistics (%d days):", len(stats))
		for _, s := range stats {
			t.Logf("  Date=%s, Register=%d, NewOrderUsers=%d, RenewalOrderUsers=%d",
				s.Date, s.Register, s.NewOrderUsers, s.RenewalOrderUsers)
		}

		// Verify all entries have date field
		for i, s := range stats {
			if s.Date == "" {
				t.Errorf("Entry %d has no date", i)
			}
		}
	})
}

// TestMultiTenantIsolation tests multi-tenant data isolation
func TestConsoleMultiTenantIsolation(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()

	// Create data for two tenants
	for tenantID := int64(1); tenantID <= 2; tenantID++ {
		// Create user
		user, err := client.ProxyUser.Create()
			SetTenantID(tenantID)
			SetPassword("password")
			SetEnable(true)
			SetIsAdmin(false)
			SetBalance(0)
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create user for tenant %d: %v", tenantID, err)
		}

		// Create ticket with status=1 (Pending)
		_, err = client.ProxyTicket.Create()
			SetTenantID(tenantID)
			SetUserID(user.ID)
			SetTitle("Ticket for Tenant " + string(rune('0'+int(tenantID))))
			SetDescription("Test")
			SetStatus(1). // Pending
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create ticket for tenant %d: %v", tenantID, err)
		}
	}

	data := &Data{db: client}
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	// Test tenant 1 can only see their own data
	count1, err := repo.QueryWaitReplyTotal(ctx, 1)
	if err != nil {
		t.Errorf("QueryWaitReplyTotal() for tenant 1 error = %v", err)
	}
	if count1 != 1 {
		t.Errorf("Tenant 1 should see 1 ticket, got %v", count1)
	}

	// Test tenant 2 can only see their own data
	count2, err := repo.QueryWaitReplyTotal(ctx, 2)
	if err != nil {
		t.Errorf("QueryWaitReplyTotal() for tenant 2 error = %v", err)
	}
	if count2 != 1 {
		t.Errorf("Tenant 2 should see 1 ticket, got %v", count2)
	}
}

// TestQueryMonthlyOrdersList tests monthly order statistics
func TestQueryMonthlyOrdersList(t *testing.T) {
	client := setupConsoleTestClient(t)
	defer client.Close()

	ctx := context.Background()

	tenantID := int64(1)
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	// Create subscribe
	subscribe, err := client.ProxySubscribe.Create()
		SetTenantID(tenantID)
		SetName("Test Subscribe")
		SetTraffic(10000)
		SetShow(true)
		SetSell(true)
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create subscribe: %v", err)
	}

	// Create user
	user, err := client.ProxyUser.Create()
		SetTenantID(tenantID)
		SetPassword("password")
		SetEnable(true)
		SetIsAdmin(false)
		SetBalance(0)
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create orders for last 6 months
	for i := 0; i < 6; i++ {
		monthStart := currentMonth.AddDate(0, -i, 0)
		orderNo := fmt.Sprintf("ORDER-MONTH-%d-%d", tenantID, i)

		_, err := client.ProxyOrder.Create()
			SetTenantID(tenantID)
			SetUserID(user.ID)
			SetSubscribeID(subscribe.ID)
			SetOrderNo(orderNo)
			SetIsNew(i%2 == 0)
			SetAmount(int64(1000 * (i + 1)))
			SetStatus(2)
			SetMethod("alipay")
			SetCreatedAt(monthStart.Add(15 * 24 * time.Hour)). // Mid-month
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create order for month -%d: %v", i, err)
		}
	}

	data := &Data{db: client}
	repo := &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	t.Run("查询近6个月订单统计", func(t *testing.T) {
		stats, err := repo.QueryMonthlyOrdersList(ctx, tenantID, now)
		if err != nil {
			t.Errorf("QueryMonthlyOrdersList() error = %v", err)
			return
		}

		// Should have up to 6 months of data
		if len(stats) > 6 {
			t.Errorf("QueryMonthlyOrdersList() returned %d months, want max 6", len(stats))
		}

		t.Logf("Monthly order statistics:")
		for _, s := range stats {
			t.Logf("  Month=%s, Total=%d, New=%d, Renewal=%d",
				s.Date, s.AmountTotal, s.NewOrderAmount, s.RenewalOrderAmount)
		}

		// Verify first month (most recent) has correct amount
		if len(stats) > 0 {
			if stats[0].AmountTotal != 1000 {
				t.Errorf("Most recent month total = %v, want 1000", stats[0].AmountTotal)
			}
		}
	})
}
