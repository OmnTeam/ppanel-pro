package server

import (
	"context"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/internal/model/server"
	"github.com/OmnTeam/ppanel-pro/pkg/ip"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
)

// Server represents a proxy server
type Server struct {
	ID             int64
	Name           string
	Country        string
	City           string
	Address        string
	Sort           int // 匹配数据库 bigint 类型
	Protocols      []*server.Protocol
	LastReportedAt int64
	Status         *ServerStatus
	CreatedAt      int64
	UpdatedAt      int64
}

// ServerStatus represents server status information
type ServerStatus struct {
	Cpu      float64
	Mem      float64
	Disk     float64
	Protocol string // Protocol field for compatibility with old project
	Online   []*ServerOnlineUser
	Status   string // "online", "warning", "offline"
}

// ServerOnlineUser represents an online user
type ServerOnlineUser struct {
	IP          []*ServerOnlineIP
	UserID      int64
	Subscribe   string
	SubscribeID int64
	Traffic     int64
	ExpiredAt   int64
}

// ServerOnlineIP represents an online IP with protocol
type ServerOnlineIP struct {
	IP       string
	Protocol string
}

// Node represents a proxy node
type Node struct {
	ID        int64
	Name      string
	Tags      []string
	Port      uint16 // 匹配数据库 smallint unsigned 和老项目
	Address   string
	ServerID  int64
	Protocol  string
	Enabled   *bool
	Sort      uint32 // 匹配数据库 INT UNSIGNED 类型
	CreatedAt int64
	UpdatedAt int64
}

// SortItem represents a sort item
type SortItem struct {
	ID   int64
	Sort int // 使用 int以兼容 Server(bigint) 和 Node(uint32)
}

// UserSubscribeInfo represents user subscribe information
type UserSubscribeInfo struct {
	UserID      int64
	SubscribeID int64
	Subscribe   string
	Download    int64
	Upload      int64
	ExpireTime  int64
}

// ServerRepo defines the interface for server repository
type ServerRepo interface {
	CreateServer(ctx context.Context, server *Server) (*Server, error)
	UpdateServer(ctx context.Context, server *Server) (*Server, error)
	DeleteServer(ctx context.Context, id int) error
	GetServerByID(ctx context.Context, id int) (*Server, error)
	FilterServerList(ctx context.Context, page, size int32, search string) (int64, []*Server, error)
	GetServerProtocols(ctx context.Context, id int) ([]*server.Protocol, error)
	ResetServerSort(ctx context.Context, sortItems []*SortItem) error
	// Redis cache methods
	GetServerStatus(ctx context.Context, serverID int) (*ServerResourceStatus, error)
	GetOnlineUsers(ctx context.Context, serverID int64, protocol string) (map[int64][]string, error)
	// Subscribe methods
	GetUserSubscribeInfo(ctx context.Context, subscribeID int) (*UserSubscribeInfo, error)
}

// ServerResourceStatus represents server resource status from cache
type ServerResourceStatus struct {
	Cpu  float64
	Mem  float64
	Disk float64
}

// NodeRepo defines the interface for node repository
type NodeRepo interface {
	CreateNode(ctx context.Context, node *Node) (*Node, error)
	UpdateNode(ctx context.Context, node *Node) (*Node, error)
	DeleteNode(ctx context.Context, id int) error
	FilterNodeList(ctx context.Context, page, size int32, search string) (int64, []*Node, error)
	ToggleNodeStatus(ctx context.Context, id int, enable *bool) (*Node, error)
	QueryNodeTags(ctx context.Context) ([]string, error)
	ResetNodeSort(ctx context.Context, sortItems []*SortItem) error
	ClearNodeCache(ctx context.Context, serverIDs []int) error
}

// MigrationRepo defines the interface for migration operations
type MigrationRepo interface {
	HasMigrateServerNode(ctx context.Context) (bool, error)
	MigrateServerNode(ctx context.Context) (uint64, uint64, string, error)
}

// ServerUsecase is the server use case
type ServerUsecase struct {
	repo     ServerRepo
	nodeRepo NodeRepo
	log      *log.Helper
}

// NewServerUsecase creates a new server use case
func NewServerUsecase(repo ServerRepo, nodeRepo NodeRepo, logger log.Logger) *ServerUsecase {
	return &ServerUsecase{
		repo:     repo,
		nodeRepo: nodeRepo,
		log:      log.NewHelper(logger),
	}
}

// CreateServer creates a new server
func (uc *ServerUsecase) CreateServer(ctx context.Context, name, country, city, address string, sort int64, protocols []*server.Protocol) (*Server, error) {
	// 1. Validate protocols
	if err := server.ValidateProtocols(protocols); err != nil {
		return nil, err
	}

	// 2. Deep copy protocols to avoid modifying caller's data
	processedProtocols := make([]*server.Protocol, len(protocols))
	for i, proto := range protocols {
		// Create a copy
		copied := &server.Protocol{}
		*copied = *proto
		processedProtocols[i] = copied
	}

	// 3. Process each protocol
	for _, protocol := range processedProtocols {
		// Handle VLESS Reality
		if protocol.Type == "vless" && protocol.Security == "reality" {
			// Generate Reality keys if not provided
			if protocol.RealityPublicKey == "" {
				public, private, err := tool.Curve25519Genkey(false, "")
				if err != nil {
					uc.log.Errorf("Failed to generate Curve25519 key: %v", err)
					return nil, err
				}
				protocol.RealityPublicKey = public
				protocol.RealityPrivateKey = private
				protocol.RealityShortId = tool.GenerateShortID(private)
			}

			// Set default Reality server address
			if protocol.RealityServerAddr == "" {
				protocol.RealityServerAddr = protocol.SNI
			}

			// Set default Reality server port
			if protocol.RealityServerPort == 0 {
				protocol.RealityServerPort = 443
			}
		}

		// Handle ShadowSocks 2022
		if protocol.Type == "shadowsocks" && strings.Contains(protocol.Cipher, "2022") {
			var length int
			switch protocol.Cipher {
			case "2022-blake3-aes-128-gcm":
				length = 16
			default:
				length = 32
			}

			// Generate server key if needed
			if len(protocol.ServerKey) != length {
				protocol.ServerKey = tool.GenerateCipher(protocol.ServerKey, length)
			}
		}
	}

	// 4. Query IP geolocation if city and country are empty (matching original logic exactly)
	if country == "" && city == "" && address != "" {
		// Query geolocation (pass address directly as in original)
		location, err := ip.GetRegionByIp(address)
		if err != nil {
			uc.log.Warnf("Failed to get geolocation for address %s: %v", address, err)
		} else if location != nil {
			country = location.Country
			city = location.City
		}
	}

	// 5. Create server
	srv := &Server{
		Name:      name,
		Country:   country,
		City:      city,
		Address:   address,
		Sort:      int(sort),
		Protocols: processedProtocols,
	}

	return uc.repo.CreateServer(ctx, srv)
}

// UpdateServer updates an existing server
func (uc *ServerUsecase) UpdateServer(ctx context.Context, id int, name, country, city, address string, sort int64, protocols []*server.Protocol) (*Server, error) {
	// 1. Get existing server to check address change (matching original FindOneServer)
	existingServer, err := uc.repo.GetServerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate protocols
	if err := server.ValidateProtocols(protocols); err != nil {
		return nil, err
	}

	// 3. Deep copy protocols to avoid modifying caller's data
	processedProtocols := make([]*server.Protocol, len(protocols))
	for i, proto := range protocols {
		// Create a copy
		copied := &server.Protocol{}
		*copied = *proto
		processedProtocols[i] = copied
	}

	// 4. Process each protocol
	for _, protocol := range processedProtocols {
		// Handle VLESS Reality
		if protocol.Type == "vless" && protocol.Security == "reality" {
			// Generate Reality keys if not provided
			if protocol.RealityPublicKey == "" {
				public, private, err := tool.Curve25519Genkey(false, "")
				if err != nil {
					uc.log.Errorf("Failed to generate Curve25519 key: %v", err)
					return nil, err
				}
				protocol.RealityPublicKey = public
				protocol.RealityPrivateKey = private
				protocol.RealityShortId = tool.GenerateShortID(private)
			}

			// Set default Reality server address
			if protocol.RealityServerAddr == "" {
				protocol.RealityServerAddr = protocol.SNI
			}

			// Set default Reality server port
			if protocol.RealityServerPort == 0 {
				protocol.RealityServerPort = 443
			}
		}

		// Handle ShadowSocks 2022
		if protocol.Type == "shadowsocks" && strings.Contains(protocol.Cipher, "2022") {
			var length int
			switch protocol.Cipher {
			case "2022-blake3-aes-128-gcm":
				length = 16
			default:
				length = 32
			}

			// Generate server key if needed
			if len(protocol.ServerKey) != length {
				protocol.ServerKey = tool.GenerateCipher(protocol.ServerKey, length)
			}
		}
	}

	// 5. Query IP geolocation only if address changed (matching original logic exactly)
	if address != existingServer.Address {
		// Query geolocation (pass address directly as in original)
		location, err := ip.GetRegionByIp(address)
		if err != nil {
			uc.log.Warnf("Failed to get geolocation for address %s: %v", address, err)
		} else if location != nil {
			// Update location info from geolocation result
			country = location.Country
			city = location.City
		}
	}

	// 6. Update server
	srv := &Server{
		ID:        int64(id),
		Name:      name,
		Country:   country,
		City:      city,
		Address:   address,
		Sort:      int(sort),
		Protocols: processedProtocols,
	}

	updatedServer, err := uc.repo.UpdateServer(ctx, srv)
	if err != nil {
		return nil, err
	}

	// 7. Clear node cache
	if err := uc.nodeRepo.ClearNodeCache(ctx, []int{id}); err != nil {
		uc.log.Warnf("Failed to clear node cache for server %d: %v", id, err)
		// Don't return error, just log warning
	}

	return updatedServer, nil
}

// DeleteServer deletes a server
func (uc *ServerUsecase) DeleteServer(ctx context.Context, id int) error {
	// Delete the server
	if err := uc.repo.DeleteServer(ctx, id); err != nil {
		return err
	}

	// Clear node cache for the deleted server
	if err := uc.nodeRepo.ClearNodeCache(ctx, []int{id}); err != nil {
		uc.log.Warnf("Failed to clear node cache for deleted server %d: %v", id, err)
		// Don't return error, just log warning
	}

	return nil
}

// FilterServerList filters server list
func (uc *ServerUsecase) FilterServerList(ctx context.Context, page, size int32, search string) (int64, []*Server, error) {
	total, servers, err := uc.repo.FilterServerList(ctx, page, size, search)
	if err != nil {
		return 0, nil, err
	}

	// Fill status for each server
	for _, srv := range servers {
		srv.Status = uc.buildServerStatus(ctx, srv)
	}

	return total, servers, nil
}

// GetServerProtocols gets server protocols
func (uc *ServerUsecase) GetServerProtocols(ctx context.Context, id int) ([]*server.Protocol, error) {
	return uc.repo.GetServerProtocols(ctx, id)
}

// ResetServerSort resets server sort
func (uc *ServerUsecase) ResetServerSort(ctx context.Context, sortItems []*SortItem) error {
	return uc.repo.ResetServerSort(ctx, sortItems)
}

// buildServerStatus builds server status information
func (uc *ServerUsecase) buildServerStatus(ctx context.Context, srv *Server) *ServerStatus {
	status := &ServerStatus{
		Cpu:      0,
		Mem:      0,
		Disk:     0,
		Protocol: "", // 与老项目保持一致，列表中为空字符串
		Online:   make([]*ServerOnlineUser, 0),
		Status:   uc.getServerStatusString(int(srv.LastReportedAt)),
	}

	// Get server resource status from Redis
	resourceStatus, err := uc.repo.GetServerStatus(ctx, int(srv.ID))
	if err != nil {
		uc.log.Warnf("Failed to get server status from cache for server %d: %v", srv.ID, err)
	} else if resourceStatus != nil {
		status.Cpu = resourceStatus.Cpu
		status.Mem = resourceStatus.Mem
		status.Disk = resourceStatus.Disk
	}

	// Get online users
	status.Online = uc.getOnlineUsers(ctx, srv.ID, srv.Protocols)

	return status
}

// getServerStatusString determines server status string based on last reported time
func (uc *ServerUsecase) getServerStatusString(lastReportedAt int) string {
	if lastReportedAt == 0 {
		return "offline"
	}

	// lastReportedAt 是毫秒时间戳，需要转换为秒
	lastReported := time.Unix(int64(lastReportedAt/1000), 0)
	elapsed := time.Since(lastReported)

	if elapsed > 5*time.Minute {
		return "offline"
	}
	if elapsed > 3*time.Minute {
		return "warning"
	}
	return "online"
}

// getOnlineUsers gets online users for a server
func (uc *ServerUsecase) getOnlineUsers(ctx context.Context, serverID int64, protocols []*server.Protocol) []*ServerOnlineUser {
	result := make([]*ServerOnlineUser, 0)

	// Collect online users from all protocols
	for _, protocol := range protocols {
		onlineData, err := uc.repo.GetOnlineUsers(ctx, serverID, protocol.Type)
		if err != nil {
			uc.log.Warnf("Failed to get online users for server %d protocol %s: %v", serverID, protocol.Type, err)
			continue
		}

		if len(onlineData) > 0 {
			for subscribeID, ips := range onlineData {
				ipList := make([]*ServerOnlineIP, 0, len(ips))
				for _, ip := range ips {
					ipList = append(ipList, &ServerOnlineIP{
						IP:       ip,
						Protocol: protocol.Type,
					})
				}

				result = append(result, &ServerOnlineUser{
					IP:          ipList,
					SubscribeID: subscribeID,
				})
			}
		}
	}

	// Merge same subscribe and fetch subscribe info
	mergedResult := uc.mergeOnlineUsers(ctx, result)
	return mergedResult
}

// mergeOnlineUsers merges online users with same subscribe ID and fetches subscribe info
func (uc *ServerUsecase) mergeOnlineUsers(ctx context.Context, users []*ServerOnlineUser) []*ServerOnlineUser {
	mergedMap := make(map[int64]*ServerOnlineUser)

	for _, user := range users {
		if existing, exists := mergedMap[user.SubscribeID]; exists {
			// Merge IPs
			existing.Traffic += user.Traffic
			existing.IP = append(existing.IP, user.IP...)
			mergedMap[user.SubscribeID] = existing
		} else {
			// Fetch subscribe info
			subscribeInfo, err := uc.repo.GetUserSubscribeInfo(ctx, int(user.SubscribeID))
			if err != nil {
				uc.log.Warnf("Failed to get subscribe info for subscribe %d: %v", user.SubscribeID, err)
				continue
			}

			newUser := &ServerOnlineUser{
				IP:          user.IP,
				UserID:      subscribeInfo.UserID,
				Subscribe:   subscribeInfo.Subscribe,
				SubscribeID: user.SubscribeID,
				Traffic:     subscribeInfo.Download + subscribeInfo.Upload,
				ExpiredAt:   subscribeInfo.ExpireTime,
			}
			mergedMap[user.SubscribeID] = newUser
		}
	}

	// Convert map to slice
	result := make([]*ServerOnlineUser, 0, len(mergedMap))
	for _, user := range mergedMap {
		result = append(result, user)
	}

	return result
}
