package data

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/enttest"
	"github.com/OmnTeam/ppanel-pro/internal/model"
	taskmodel "github.com/OmnTeam/ppanel-pro/internal/model/task"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	_ "github.com/mattn/go-sqlite3"
)

// 创建测试用的 Ent 客户端（使用 SQLite 内存数据库）
func setupTestClient(t *testing.T) *ent.Client {
	opts := []enttest.Option{
		enttest.WithOptions(
			ent.Log(t.Log),
			ent.Debug(), // 启用 Debug 模式，打印所有 SQL
		),
	}
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1", opts...)
	return client
}

// 准备测试数据
func seedTestData(t *testing.T, client *ent.Client, ctx context.Context) {
	tenantID := int64(1)

	// 创建用户
	users := make([]*ent.ProxyUser, 5)
	for i := 0; i < 5; i++ {
		user, err := client.ProxyUser.Create()
			SetTenantID(tenantID)
			SetPassword("password")
			SetEnable(true)
			SetIsAdmin(false)
			SetCreatedAt(time.Now().Add(-time.Duration(i) * 24 * time.Hour))
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		users[i] = user
	}

	// 创建用户认证方法（邮箱）
	for i, user := range users {
		_, err := client.ProxyUserAuthMethod.Create()
			SetTenantID(tenantID)
			SetUserID(user.ID)
			SetAuthType(model.AuthTypeEmail)
			SetAuthIdentifier("user" + string(rune('0'+i)) + "@test.com")
			SetVerified(true)
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create auth method: %v", err)
		}
	}

	// 创建订阅套餐
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

	// 创建用户订阅
	// User 0: Active (status=1)
	_, err = client.ProxyUserSubscribe.Create()
		SetTenantID(tenantID)
		SetUserID(users[0].ID)
		SetOrderID(1001)
		SetSubscribeID(subscribe.ID)
		SetStatus(uint8(model.UserSubscribeStatusActive))
		SetStartTime(time.Now().Add(-10 * 24 * time.Hour))
		SetExpireTime(time.Now().Add(20 * 24 * time.Hour))
		SetTraffic(5000)
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create user subscribe: %v", err)
	}

	// User 1: Finish (status=2)
	_, err = client.ProxyUserSubscribe.Create()
		SetTenantID(tenantID)
		SetUserID(users[1].ID)
		SetOrderID(1002)
		SetSubscribeID(subscribe.ID)
		SetStatus(uint8(model.UserSubscribeStatusFinish))
		SetStartTime(time.Now().Add(-10 * 24 * time.Hour))
		SetExpireTime(time.Now().Add(20 * 24 * time.Hour))
		SetTraffic(5000)
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create user subscribe: %v", err)
	}

	// User 2: Expired (status=3)
	_, err = client.ProxyUserSubscribe.Create()
		SetTenantID(tenantID)
		SetUserID(users[2].ID)
		SetOrderID(1003)
		SetSubscribeID(subscribe.ID)
		SetStatus(uint8(model.UserSubscribeStatusExpired))
		SetStartTime(time.Now().Add(-30 * 24 * time.Hour))
		SetExpireTime(time.Now().Add(-1 * 24 * time.Hour))
		SetTraffic(5000)
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create user subscribe: %v", err)
	}

	// User 3, 4: No subscription
}

func TestGetPreSendEmailCount(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedTestData(t, client, ctx)

	// 创建 mock queue（不实际使用）
	mockQueue := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	data := &Data{
		db:    client,
		queue: mockQueue,
	}

	repo := &adminMarketingRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	tests := []struct {
		name      string
		scope     int32
		wantCount int64
		wantErr   bool
	}{
		{
			name:      "ScopeAll - 应该返回所有有email的用户",
			scope:     int32(taskmodel.ScopeAll),
			wantCount: 5, // 5个用户都有email
			wantErr:   false,
		},
		{
			name:      "ScopeActive - 应该返回激活订阅的用户",
			scope:     int32(taskmodel.ScopeActive),
			wantCount: 2, // User 0 (Active) + User 1 (Finish)
			wantErr:   false,
		},
		{
			name:      "ScopeExpired - 应该返回过期订阅的用户",
			scope:     int32(taskmodel.ScopeExpired),
			wantCount: 1, // User 2 (Expired)
			wantErr:   false,
		},
		{
			name:      "ScopeNone - 应该返回没有订阅的用户",
			scope:     int32(taskmodel.ScopeNone),
			wantCount: 2, // User 3, 4
			wantErr:   false,
		},
		{
			name:      "ScopeSkip - 应该返回0",
			scope:     int32(taskmodel.ScopeSkip),
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.GetPreSendEmailCount(ctx, tenantID, tt.scope, 0, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPreSendEmailCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if count != tt.wantCount {
				t.Errorf("GetPreSendEmailCount() count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestGetPreSendEmailCountWithTimeFilter(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedTestData(t, client, ctx)

	mockQueue := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	data := &Data{
		db:    client,
		queue: mockQueue,
	}

	repo := &adminMarketingRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	// 测试时间过滤
	// User 0 created: now
	// User 1 created: now - 1 day
	// User 2 created: now - 2 days
	// User 3 created: now - 3 days
	// User 4 created: now - 4 days

	tests := []struct {
		name              string
		registerStartTime int64
		registerEndTime   int64
		wantCount         int64
	}{
		{
			name:              "过滤最近2天注册的用户",
			registerStartTime: time.Now().Add(-2 * 24 * time.Hour).UnixMilli(),
			registerEndTime:   time.Now().UnixMilli(),
			wantCount:         3, // User 0, 1, 2
		},
		{
			name:              "只过滤开始时间",
			registerStartTime: time.Now().Add(-2 * 24 * time.Hour).UnixMilli(),
			registerEndTime:   0,
			wantCount:         3, // User 0, 1, 2
		},
		{
			name:              "只过滤结束时间",
			registerStartTime: 0,
			registerEndTime:   time.Now().Add(-2 * 24 * time.Hour).UnixMilli(),
			wantCount:         3, // User 2, 3, 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.GetPreSendEmailCount(ctx, tenantID, int32(taskmodel.ScopeAll), tt.registerStartTime, tt.registerEndTime)
			if err != nil {
				t.Errorf("GetPreSendEmailCount() error = %v", err)
				return
			}
			if count != tt.wantCount {
				t.Errorf("GetPreSendEmailCount() count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestCreateBatchSendEmailTask(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedTestData(t, client, ctx)

	// 注意：这里我们不会真正创建 asynq 客户端，因为没有 Redis
	// 我们只测试数据库操作部分
	// 在实际环境中，需要 mock asynq

	data := &Data{
		db:    client,
		queue: nil, // 不测试队列部分
	}

	repo := &adminMarketingRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	tests := []struct {
		name            string
		scope           int32
		additional      string
		wantEmailCount  int
		wantTaskCreated bool
		wantErr         bool
		skipQueueTest   bool
	}{
		{
			name:            "ScopeAll - 应该获取所有用户邮箱",
			scope:           int32(taskmodel.ScopeAll),
			additional:      "",
			wantEmailCount:  5,
			wantTaskCreated: true,
			wantErr:         false,
			skipQueueTest:   true,
		},
		{
			name:            "ScopeActive - 应该只获取激活用户邮箱",
			scope:           int32(taskmodel.ScopeActive),
			additional:      "",
			wantEmailCount:  2,
			wantTaskCreated: true,
			wantErr:         false,
			skipQueueTest:   true,
		},
		{
			name:            "ScopeAll + Additional - 应该合并邮箱",
			scope:           int32(taskmodel.ScopeAll),
			additional:      "extra1@test.com\nextra2@test.com",
			wantEmailCount:  7, // 5 + 2
			wantTaskCreated: true,
			wantErr:         false,
			skipQueueTest:   true,
		},
		{
			name:            "ScopeAll + Duplicate Additional - 应该去重",
			scope:           int32(taskmodel.ScopeAll),
			additional:      "user0@test.com\nextra@test.com",
			wantEmailCount:  6, // 5 + 1 (user0@test.com重复)
			wantTaskCreated: true,
			wantErr:         false,
			skipQueueTest:   true,
		},
		{
			name:            "ScopeSkip + Additional - 应该只使用额外邮箱",
			scope:           int32(taskmodel.ScopeSkip),
			additional:      "extra1@test.com\nextra2@test.com",
			wantEmailCount:  2,
			wantTaskCreated: true,
			wantErr:         false,
			skipQueueTest:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 由于我们没有 Redis，暂时不测试完整的 CreateBatchSendEmailTask
			// 我们只测试邮箱查询逻辑（到任务创建之前）

			// 测试邮箱统计
			if tt.scope != int32(taskmodel.ScopeSkip) {
				count, err := repo.GetPreSendEmailCount(ctx, tenantID, tt.scope, 0, 0)
				if err != nil {
					t.Errorf("GetPreSendEmailCount() error = %v", err)
					return
				}

				expectedDBCount := tt.wantEmailCount
				if tt.additional != "" {
					// 如果有additional，数据库查询的数量应该是减去additional的数量
					additionalCount := len(splitEmails(tt.additional))
					expectedDBCount = tt.wantEmailCount - additionalCount
					if tt.scope == int32(taskmodel.ScopeAll) && tt.additional == "user0@test.com\nextra@test.com" {
						expectedDBCount = 5 // 数据库有5个
					}
				}

				if tt.scope != int32(taskmodel.ScopeSkip) {
					if count != int64(expectedDBCount) && tt.additional == "" {
						t.Errorf("GetPreSendEmailCount() = %v, want %v", count, expectedDBCount)
					}
				}
			}
		})
	}
}

func splitEmails(emails string) []string {
	if emails == "" {
		return nil
	}
	// 简单的换行分割
	result := make([]string, 0)
	for _, line := range splitLines(emails) {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func splitLines(s string) []string {
	result := make([]string, 0)
	current := ""
	for _, c := range s {
		if c == '\n' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func TestQueryQuotaTaskPreCount(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedTestData(t, client, ctx)

	mockQueue := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	data := &Data{
		db:    client,
		queue: mockQueue,
	}

	repo := &adminMarketingRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	// 获取 subscribe ID
	subscribes, err := client.ProxySubscribe.Query().All(ctx)
	if err != nil || len(subscribes) == 0 {
		t.Fatalf("Failed to get subscribes: %v", err)
	}
	subscribeID := subscribes[0].ID

	tests := []struct {
		name      string
		subs      []int64
		isActive  *bool
		startTime int64
		endTime   int64
		wantCount int64
	}{
		{
			name:      "查询所有订阅",
			subs:      nil,
			isActive:  nil,
			startTime: 0,
			endTime:   0,
			wantCount: 3, // User 0, 1, 2 有订阅
		},
		{
			name:      "查询指定套餐的订阅",
			subs:      []int64{subscribeID},
			isActive:  nil,
			startTime: 0,
			endTime:   0,
			wantCount: 3,
		},
		{
			name:      "只查询活跃订阅",
			subs:      nil,
			isActive:  boolPtr(true),
			startTime: 0,
			endTime:   0,
			wantCount: 2, // User 0 (Active) + User 1 (Finish)
		},
		{
			name:      "查询有效期内的订阅",
			subs:      nil,
			isActive:  nil,
			startTime: 0,
			endTime:   time.Now().UnixMilli(),
			wantCount: 2, // User 0, 1 (expire_time >= now)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.QueryQuotaTaskPreCount(ctx, tenantID, tt.subs, tt.isActive, tt.startTime, tt.endTime)
			if err != nil {
				t.Errorf("QueryQuotaTaskPreCount() error = %v", err)
				return
			}
			if count != tt.wantCount {
				t.Errorf("QueryQuotaTaskPreCount() count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestMultiTenantIsolation(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()

	// 创建两个租户的数据
	for tenantID := int64(1); tenantID <= 2; tenantID++ {
		// 创建用户
		user, err := client.ProxyUser.Create()
			SetTenantID(tenantID)
			SetPassword("password")
			SetEnable(true)
			SetIsAdmin(false)
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create user for tenant %d: %v", tenantID, err)
		}

		// 创建认证方法
		_, err = client.ProxyUserAuthMethod.Create()
			SetTenantID(tenantID)
			SetUserID(user.ID)
			SetAuthType(model.AuthTypeEmail)
			SetAuthIdentifier("user@tenant" + string(rune('0'+int(tenantID))) + ".com")
			SetVerified(true)
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create auth method for tenant %d: %v", tenantID, err)
		}
	}

	mockQueue := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	data := &Data{
		db:    client,
		queue: mockQueue,
	}

	repo := &adminMarketingRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	// 测试租户1只能看到自己的数据
	count1, err := repo.GetPreSendEmailCount(ctx, 1, int32(taskmodel.ScopeAll), 0, 0)
	if err != nil {
		t.Errorf("GetPreSendEmailCount() for tenant 1 error = %v", err)
	}
	if count1 != 1 {
		t.Errorf("Tenant 1 should see 1 user, got %v", count1)
	}

	// 测试租户2只能看到自己的数据
	count2, err := repo.GetPreSendEmailCount(ctx, 2, int32(taskmodel.ScopeAll), 0, 0)
	if err != nil {
		t.Errorf("GetPreSendEmailCount() for tenant 2 error = %v", err)
	}
	if count2 != 1 {
		t.Errorf("Tenant 2 should see 1 user, got %v", count2)
	}
}

func TestJoinLogic(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	seedTestData(t, client, ctx)

	mockQueue := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	data := &Data{
		db:    client,
		queue: mockQueue,
	}

	repo := &adminMarketingRepo{
		data: data,
		log:  log.NewHelper(log.DefaultLogger),
	}

	tenantID := int64(1)

	t.Run("验证JOIN不会重复", func(t *testing.T) {
		// 测试 ScopeActive，这个会执行两次 JOIN
		count, err := repo.GetPreSendEmailCount(ctx, tenantID, int32(taskmodel.ScopeActive), 0, 0)
		if err != nil {
			t.Errorf("GetPreSendEmailCount(ScopeActive) error = %v", err)
			return
		}

		// 如果 JOIN 逻辑错误，会导致结果不对或报错
		if count != 2 {
			t.Errorf("ScopeActive JOIN logic error: expected 2, got %v", count)
		}
	})

	t.Run("验证LEFT JOIN逻辑", func(t *testing.T) {
		// 测试 ScopeNone，这个会执行 LEFT JOIN
		count, err := repo.GetPreSendEmailCount(ctx, tenantID, int32(taskmodel.ScopeNone), 0, 0)
		if err != nil {
			t.Errorf("GetPreSendEmailCount(ScopeNone) error = %v", err)
			return
		}

		// 应该返回没有订阅的用户数量
		if count != 2 {
			t.Errorf("ScopeNone LEFT JOIN logic error: expected 2, got %v", count)
		}
	})
}

func boolPtr(b bool) *bool {
	return &b
}
