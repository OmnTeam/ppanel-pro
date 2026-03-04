package data

import (
	"context"
	"encoding/json"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyserver"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	serverbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/server"
	servermodel "github.com/OmnTeam/ppanel-pro/internal/model/server"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type adminServerRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminServerRepo creates a new admin server repository
func NewAdminServerRepo(data *Data, logger log.Logger) serverbiz.ServerRepo {
	return &adminServerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateServer creates a new server
func (r *adminServerRepo) CreateServer(ctx context.Context, server *serverbiz.Server) (*serverbiz.Server, error) {
	// Marshal protocols to JSON
	protocolsJSON, err := servermodel.MarshalProtocols(server.Protocols)
	if err != nil {
		return nil, err
	}

	// Create server
	created, err := r.data.db.ProxyServer.Create().
		SetName(server.Name).
		SetCountry(server.Country).
		SetCity(server.City).
		SetServerAddr(server.Address).
		SetSort(int(server.Sort)).
		SetProtocol(protocolsJSON).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Unmarshal protocols
	protocols, _ := servermodel.UnmarshalProtocols(created.Protocol)

	lastReportedAt := int64(0)
	if created.LastReportedAt != nil {
		lastReportedAt = created.LastReportedAt.UnixMilli() // 使用毫秒时间戳，与老项目保持一致
	}

	return &serverbiz.Server{
		ID: int64(created.ID),

		Name:           created.Name,
		Country:        created.Country,
		City:           created.City,
		Address:        created.ServerAddr,
		Sort:           created.Sort,
		Protocols:      protocols,
		LastReportedAt: lastReportedAt,
		CreatedAt:      created.CreatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
		UpdatedAt:      created.UpdatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
	}, nil
}

// UpdateServer updates an existing server
func (r *adminServerRepo) UpdateServer(ctx context.Context, server *serverbiz.Server) (*serverbiz.Server, error) {
	// Marshal protocols to JSON
	protocolsJSON, err := servermodel.MarshalProtocols(server.Protocols)
	if err != nil {
		return nil, err
	}

	// Update server
	updated, err := r.data.db.ProxyServer.UpdateOneID(server.ID).
		SetName(server.Name).
		SetCountry(server.Country).
		SetCity(server.City).
		SetServerAddr(server.Address).
		SetSort(int(server.Sort)).
		SetProtocol(protocolsJSON).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Unmarshal protocols
	protocols, _ := servermodel.UnmarshalProtocols(updated.Protocol)

	lastReportedAt := int64(0)
	if updated.LastReportedAt != nil {
		lastReportedAt = updated.LastReportedAt.UnixMilli() // 使用毫秒时间戳，与老项目保持一致
	}

	return &serverbiz.Server{
		ID: int64(updated.ID),

		Name:           updated.Name,
		Country:        updated.Country,
		City:           updated.City,
		Address:        updated.ServerAddr,
		Sort:           updated.Sort,
		Protocols:      protocols,
		LastReportedAt: lastReportedAt,
		CreatedAt:      updated.CreatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
		UpdatedAt:      updated.UpdatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
	}, nil
}

// DeleteServer deletes a server
func (r *adminServerRepo) DeleteServer(ctx context.Context, id int) error {
	// Delete the server
	err := r.data.db.ProxyServer.DeleteOneID(int64(id)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetServerByID gets a server by ID
func (r *adminServerRepo) GetServerByID(ctx context.Context, id int) (*serverbiz.Server, error) {
	server, err := r.data.db.ProxyServer.Query().
		Where(proxyserver.ID(int64(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("server not found or access denied")
		}
		return nil, err
	}

	protocols, _ := servermodel.UnmarshalProtocols(server.Protocol)
	lastReportedAt := int64(0)
	if server.LastReportedAt != nil {
		lastReportedAt = server.LastReportedAt.UnixMilli() // 使用毫秒时间戳，与老项目保持一致
	}

	return &serverbiz.Server{
		ID: int64(server.ID),

		Name:           server.Name,
		Country:        server.Country,
		City:           server.City,
		Address:        server.ServerAddr,
		Sort:           server.Sort,
		Protocols:      protocols,
		LastReportedAt: lastReportedAt,
		CreatedAt:      server.CreatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
		UpdatedAt:      server.UpdatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
	}, nil
}

// FilterServerList filters server list
func (r *adminServerRepo) FilterServerList(ctx context.Context, page, size int32, search string) (int64, []*serverbiz.Server, error) {
	query := r.data.db.ProxyServer.Query()
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(func(s *sql.Selector) {
			s.Where(sql.Or(
				sql.Like(s.C(proxyserver.FieldName), searchPattern),
				sql.Like(s.C(proxyserver.FieldServerAddr), searchPattern),
			))
		})
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return 0, nil, err
	}

	// Get list
	list, err := query.
		Order(ent.Asc(proxyserver.FieldSort)).
		Limit(int(size)).
		Offset(int((page - 1) * size)).
		All(ctx)
	if err != nil {
		return 0, nil, err
	}

	servers := make([]*serverbiz.Server, 0, len(list))
	for _, item := range list {
		protocols, _ := servermodel.UnmarshalProtocols(item.Protocol)
		lastReportedAt := int64(0)
		if item.LastReportedAt != nil {
			lastReportedAt = item.LastReportedAt.UnixMilli() // 使用毫秒时间戳，与老项目保持一致
		}
		servers = append(servers, &serverbiz.Server{
			ID: int64(item.ID),

			Name:           item.Name,
			Country:        item.Country,
			City:           item.City,
			Address:        item.ServerAddr,
			Sort:           item.Sort,
			Protocols:      protocols,
			LastReportedAt: lastReportedAt,
			CreatedAt:      item.CreatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
			UpdatedAt:      item.UpdatedAt.UnixMilli(), // 返回毫秒时间戳，与老项目保持一致
		})
	}

	return int64(total), servers, nil
}

// GetServerProtocols gets server protocols
func (r *adminServerRepo) GetServerProtocols(ctx context.Context, id int) ([]*servermodel.Protocol, error) {
	// Query server
	server, err := r.data.db.ProxyServer.Query().
		Where(proxyserver.ID(int64(id))).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return servermodel.UnmarshalProtocols(server.Protocol)
}

// ResetServerSort resets server sort order
func (r *adminServerRepo) ResetServerSort(ctx context.Context, sortItems []*serverbiz.SortItem) error {
	for _, item := range sortItems {
		// Update the server sort
		affected, err := r.data.db.ProxyServer.Update().
			Where(proxyserver.ID(item.ID)).
			SetSort(int(item.Sort)).
			Save(ctx)
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf("server %d not found", item.ID)
		}
	}
	return nil
}

// GetServerStatus gets server status from Redis cache
func (r *adminServerRepo) GetServerStatus(ctx context.Context, serverID int) (*serverbiz.ServerResourceStatus, error) {
	// Get server status from cache

	key := fmt.Sprintf(StatusCacheKey, int(serverID))

	result, err := r.data.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if result == "" {
		return nil, nil
	}

	var status serverbiz.ServerResourceStatus
	if err := json.Unmarshal([]byte(result), &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// GetOnlineUsers gets online users from Redis cache
func (r *adminServerRepo) GetOnlineUsers(ctx context.Context, serverID int64, protocol string) (map[int64][]string, error) {
	// Get online users from cache

	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, int(serverID), protocol)

	result, err := r.data.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return make(map[int64][]string), nil
		}
		return nil, err
	}

	if result == "" {
		return make(map[int64][]string), nil
	}

	var onlineUsers map[int64][]string
	if err := json.Unmarshal([]byte(result), &onlineUsers); err != nil {
		return nil, err
	}

	return onlineUsers, nil
}

// GetUserSubscribeInfo gets user subscribe information
func (r *adminServerRepo) GetUserSubscribeInfo(ctx context.Context, subscribeID int) (*serverbiz.UserSubscribeInfo, error) {
	userSubscribe, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.ID(int64(subscribeID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("subscribe %d not found", subscribeID)
		}
		return nil, err
	}

	// Get subscribe name
	subscribeName := ""
	if userSubscribe.SubscribeID != 0 {
		subscribeEntity, err := r.data.db.ProxySubscribe.Get(ctx, userSubscribe.SubscribeID)
		if err == nil && subscribeEntity != nil {
			subscribeName = subscribeEntity.Name
		}
	}

	// Handle optional fields
	var download, upload int64
	if userSubscribe.Download != nil {
		download = int64(*userSubscribe.Download)
	}
	if userSubscribe.Upload != nil {
		upload = int64(*userSubscribe.Upload)
	}

	var expireTime int64
	if userSubscribe.ExpireTime != nil {
		expireTime = userSubscribe.ExpireTime.UnixMilli()
	}

	return &serverbiz.UserSubscribeInfo{
		UserID:      int64(userSubscribe.UserID),
		SubscribeID: int64(userSubscribe.ID),
		Subscribe:   subscribeName,
		Download:    download,
		Upload:      upload,
		ExpireTime:  expireTime,
	}, nil
}
