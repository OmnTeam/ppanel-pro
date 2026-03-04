package server

import (
	"context"
	"strconv"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/server/v1"
	serverbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/server"
	servermodel "github.com/OmnTeam/ppanel-pro/internal/model/server"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

// ServerService is the server service
type ServerService struct {
	v1.UnimplementedServerServiceServer

	serverUc    *serverbiz.ServerUsecase
	nodeUc      *serverbiz.NodeUsecase
	migrationUc *serverbiz.MigrationUsecase
	log         *log.Helper
}

// NewServerService creates a new server service
func NewServerService(serverUc *serverbiz.ServerUsecase, nodeUc *serverbiz.NodeUsecase, migrationUc *serverbiz.MigrationUsecase, logger log.Logger) *ServerService {
	return &ServerService{
		serverUc:    serverUc,
		nodeUc:      nodeUc,
		migrationUc: migrationUc,
		log:         log.NewHelper(logger),
	}
}

// CreateServer creates a new server
func (s *ServerService) CreateServer(ctx context.Context, req *v1.CreateServerRequest) (*v1.CreateServerReply, error) {
	protocols := protosToModelProtocols(req.Protocols)

	server, err := s.serverUc.CreateServer(ctx, req.Name, req.Country, req.City, req.Address, int64(req.Sort), protocols)
	if err != nil {
		return nil, err
	}

	return &v1.CreateServerReply{
		Code:    responsecode.AdminCreateServerSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminCreateServerSuccess],
		Data: &v1.CreateServerData{
			Server: serverToProto(server),
		},
	}, nil
}

// UpdateServer updates an existing server
func (s *ServerService) UpdateServer(ctx context.Context, req *v1.UpdateServerRequest) (*v1.UpdateServerReply, error) {
	protocols := protosToModelProtocols(req.Protocols)

	server, err := s.serverUc.UpdateServer(ctx, int(parseInt64(req.Id)), req.Name, req.Country, req.City, req.Address, int64(req.Sort), protocols)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateServerReply{
		Code:    responsecode.AdminUpdateServerSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminUpdateServerSuccess],
		Data: &v1.UpdateServerData{
			Server: serverToProto(server),
		},
	}, nil
}

// DeleteServer deletes a server
func (s *ServerService) DeleteServer(ctx context.Context, req *v1.DeleteServerRequest) (*v1.DeleteServerReply, error) {
	err := s.serverUc.DeleteServer(ctx, int(parseInt64(req.Id)))
	if err != nil {
		return nil, err
	}

	return &v1.DeleteServerReply{
		Code:    responsecode.AdminDeleteServerSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminDeleteServerSuccess],
		Data: &v1.DeleteServerData{
			Success: true,
		},
	}, nil
}

// FilterServerList filters server list
func (s *ServerService) FilterServerList(ctx context.Context, req *v1.FilterServerListRequest) (*v1.FilterServerListReply, error) {
	total, servers, err := s.serverUc.FilterServerList(ctx, int32(req.Page), int32(req.Size), req.Search)
	if err != nil {
		return nil, err
	}

	list := make([]*v1.Server, 0, len(servers))
	for _, server := range servers {
		list = append(list, serverToProto(server))
	}

	return &v1.FilterServerListReply{
		Code:    responsecode.AdminFilterServerListSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminFilterServerListSuccess],
		Data: &v1.FilterServerListData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// GetServerProtocols gets server protocols
func (s *ServerService) GetServerProtocols(ctx context.Context, req *v1.GetServerProtocolsRequest) (*v1.GetServerProtocolsReply, error) {
	protocols, err := s.serverUc.GetServerProtocols(ctx, int(parseInt64(req.Id)))
	if err != nil {
		return nil, err
	}

	protoProtocols := modelProtocolsToProtos(protocols)

	return &v1.GetServerProtocolsReply{
		Code:    responsecode.AdminGetServerProtocolsSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminGetServerProtocolsSuccess],
		Data: &v1.GetServerProtocolsData{
			Protocols: protoProtocols,
		},
	}, nil
}

// CreateNode creates a new node
func (s *ServerService) CreateNode(ctx context.Context, req *v1.CreateNodeRequest) (*v1.CreateNodeReply, error) {
	node, err := s.nodeUc.CreateNode(ctx, req.Name, req.Tags, uint16(req.Port), req.Address, parseInt64(req.ServerId), req.Protocol, req.Enabled)
	if err != nil {
		return nil, err
	}

	return &v1.CreateNodeReply{
		Code:    responsecode.AdminCreateNodeSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminCreateNodeSuccess],
		Data: &v1.CreateNodeData{
			Node: nodeToProto(node),
		},
	}, nil
}

// UpdateNode updates an existing node
func (s *ServerService) UpdateNode(ctx context.Context, req *v1.UpdateNodeRequest) (*v1.UpdateNodeReply, error) {
	node, err := s.nodeUc.UpdateNode(ctx, int(parseInt64(req.Id)), req.Name, req.Tags, uint16(req.Port), req.Address, parseInt64(req.ServerId), req.Protocol, req.Enabled)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateNodeReply{
		Code:    responsecode.AdminUpdateNodeSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminUpdateNodeSuccess],
		Data: &v1.UpdateNodeData{
			Node: nodeToProto(node),
		},
	}, nil
}

// DeleteNode deletes a node
func (s *ServerService) DeleteNode(ctx context.Context, req *v1.DeleteNodeRequest) (*v1.DeleteNodeReply, error) {
	err := s.nodeUc.DeleteNode(ctx, int(parseInt64(req.Id)))
	if err != nil {
		return nil, err
	}

	return &v1.DeleteNodeReply{
		Code:    responsecode.AdminDeleteNodeSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminDeleteNodeSuccess],
		Data: &v1.DeleteNodeData{
			Success: true,
		},
	}, nil
}

// FilterNodeList filters node list
func (s *ServerService) FilterNodeList(ctx context.Context, req *v1.FilterNodeListRequest) (*v1.FilterNodeListReply, error) {
	total, nodes, err := s.nodeUc.FilterNodeList(ctx, int32(req.Page), int32(req.Size), req.Search)
	if err != nil {
		return nil, err
	}

	list := make([]*v1.Node, 0, len(nodes))
	for _, node := range nodes {
		list = append(list, nodeToProto(node))
	}

	return &v1.FilterNodeListReply{
		Code:    responsecode.AdminFilterNodeListSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminFilterNodeListSuccess],
		Data: &v1.FilterNodeListData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// ToggleNodeStatus toggles node status
func (s *ServerService) ToggleNodeStatus(ctx context.Context, req *v1.ToggleNodeStatusRequest) (*v1.ToggleNodeStatusReply, error) {
	node, err := s.nodeUc.ToggleNodeStatus(ctx, int(parseInt64(req.Id)), req.Enable)
	if err != nil {
		return nil, err
	}

	return &v1.ToggleNodeStatusReply{
		Code:    responsecode.AdminToggleNodeStatusSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminToggleNodeStatusSuccess],
		Data: &v1.ToggleNodeStatusData{
			Node: nodeToProto(node),
		},
	}, nil
}

// QueryNodeTag queries all node tags
func (s *ServerService) QueryNodeTag(ctx context.Context, req *v1.QueryNodeTagRequest) (*v1.QueryNodeTagReply, error) {
	tags, err := s.nodeUc.QueryNodeTags(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.QueryNodeTagReply{
		Code:    responsecode.AdminQueryNodeTagSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminQueryNodeTagSuccess],
		Data: &v1.QueryNodeTagData{
			Tags: tags,
		},
	}, nil
}

// HasMigrateServerNode checks if there's data to migrate
func (s *ServerService) HasMigrateServerNode(ctx context.Context, req *v1.HasMigrateServerNodeRequest) (*v1.HasMigrateServerNodeReply, error) {
	hasMigrate, err := s.migrationUc.HasMigrateServerNode(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.HasMigrateServerNodeReply{
		Code:    responsecode.AdminHasMigrateServerNodeSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminHasMigrateServerNodeSuccess],
		Data: &v1.HasMigrateServerNodeData{
			HasMigrate: hasMigrate,
		},
	}, nil
}

// MigrateServerNode migrates server and node data
func (s *ServerService) MigrateServerNode(ctx context.Context, req *v1.MigrateServerNodeRequest) (*v1.MigrateServerNodeReply, error) {
	success, fail, _, err := s.migrationUc.MigrateServerNode(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.MigrateServerNodeReply{
		Code:    responsecode.AdminMigrateServerNodeSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminMigrateServerNodeSuccess],
		Data: &v1.MigrateServerNodeData{
			Success: success,
			Fail:    fail,
		},
	}, nil
}

// ResetSortWithServer resets server sort
func (s *ServerService) ResetSortWithServer(ctx context.Context, req *v1.ResetSortRequest) (*v1.ResetSortReply, error) {
	sortItems := make([]*serverbiz.SortItem, 0, len(req.Sort))
	for _, item := range req.Sort {
		sortItems = append(sortItems, &serverbiz.SortItem{
			ID:   parseInt64(item.Id),
			Sort: int(parseInt64(item.Sort)),
		})
	}

	err := s.serverUc.ResetServerSort(ctx, sortItems)
	if err != nil {
		return nil, err
	}

	return &v1.ResetSortReply{
		Code:    responsecode.AdminResetSortWithServerSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminResetSortWithServerSuccess],
		Data: &v1.ResetSortData{
			Success: true,
		},
	}, nil
}

// ResetSortWithNode resets node sort
func (s *ServerService) ResetSortWithNode(ctx context.Context, req *v1.ResetSortRequest) (*v1.ResetSortReply, error) {
	sortItems := make([]*serverbiz.SortItem, 0, len(req.Sort))
	for _, item := range req.Sort {
		sortItems = append(sortItems, &serverbiz.SortItem{
			ID:   parseInt64(item.Id),
			Sort: int(parseInt64(item.Sort)),
		})
	}

	err := s.nodeUc.ResetNodeSort(ctx, sortItems)
	if err != nil {
		return nil, err
	}

	return &v1.ResetSortReply{
		Code:    responsecode.AdminResetSortWithNodeSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminResetSortWithNodeSuccess],
		Data: &v1.ResetSortData{
			Success: true,
		},
	}, nil
}

// Helper functions for conversion

func serverToProto(s *serverbiz.Server) *v1.Server {
	if s == nil {
		return nil
	}

	var status *v1.ServerStatus
	if s.Status != nil {
		onlineUsers := make([]*v1.ServerOnlineUser, 0, len(s.Status.Online))
		for _, user := range s.Status.Online {
			ips := make([]*v1.ServerOnlineIP, 0, len(user.IP))
			for _, ip := range user.IP {
				ips = append(ips, &v1.ServerOnlineIP{
					Ip:       ip.IP,
					Protocol: ip.Protocol,
				})
			}
			onlineUsers = append(onlineUsers, &v1.ServerOnlineUser{
				Ip:          ips,
				UserId:      strconv.FormatInt(user.UserID, 10),
				Subscribe:   user.Subscribe,
				SubscribeId: strconv.FormatInt(user.SubscribeID, 10),
				Traffic:     strconv.FormatInt(user.Traffic, 10),
				ExpiredAt:   strconv.FormatInt(user.ExpiredAt, 10),
			})
		}

		status = &v1.ServerStatus{
			Cpu:    s.Status.Cpu,
			Mem:    s.Status.Mem,
			Disk:   s.Status.Disk,
			Online: onlineUsers,
			Status: s.Status.Status,
		}
	}

	return &v1.Server{
		Id:             strconv.FormatInt(s.ID, 10),
		Name:           s.Name,
		Country:        s.Country,
		City:           s.City,
		Address:        s.Address,
		Sort:           strconv.FormatInt(int64(s.Sort), 10),
		Protocols:      modelProtocolsToProtos(s.Protocols),
		LastReportedAt: int32(s.LastReportedAt),
		Status:         status,
		CreatedAt:      strconv.FormatInt(s.CreatedAt, 10),
		UpdatedAt:      strconv.FormatInt(s.UpdatedAt, 10),
	}
}

func nodeToProto(n *serverbiz.Node) *v1.Node {
	if n == nil {
		return nil
	}

	return &v1.Node{
		Id:        strconv.FormatInt(n.ID, 10),
		Name:      n.Name,
		Tags:      n.Tags,
		Port:      uint32(n.Port), // Convert uint16 to uint32 for Proto
		Address:   n.Address,
		ServerId:  strconv.FormatInt(n.ServerID, 10),
		Protocol:  n.Protocol,
		Enabled:   n.Enabled, // 直接使用 *bool，与老项目一致
		Sort:      n.Sort,
		CreatedAt: strconv.FormatInt(n.CreatedAt, 10),
		UpdatedAt: strconv.FormatInt(n.UpdatedAt, 10),
	}
}

func protosToModelProtocols(protos []*v1.Protocol) []*servermodel.Protocol {
	if protos == nil {
		return nil
	}

	protocols := make([]*servermodel.Protocol, 0, len(protos))
	for _, p := range protos {
		protocols = append(protocols, protoToModelProtocol(p))
	}
	return protocols
}

func protoToModelProtocol(p *v1.Protocol) *servermodel.Protocol {
	if p == nil {
		return nil
	}

	return &servermodel.Protocol{
		Type:                    p.Type,
		Port:                    p.Port,
		Enable:                  p.Enable,
		Security:                p.Security,
		SNI:                     p.Sni,
		AllowInsecure:           p.AllowInsecure,
		Fingerprint:             p.Fingerprint32,
		RealityServerAddr:       p.RealityServerAddr,
		RealityServerPort:       p.RealityServerPort,
		RealityPrivateKey:       p.RealityPrivateKey,
		RealityPublicKey:        p.RealityPublicKey,
		RealityShortId:          p.RealityShortId,
		Transport:               p.Transport,
		Host:                    p.Host,
		Path:                    p.Path,
		ServiceName:             p.ServiceName,
		Cipher:                  p.Cipher,
		ServerKey:               p.ServerKey,
		Flow:                    p.Flow,
		HopPorts:                p.HopPorts,
		HopInterval:             p.HopInterval,
		ObfsPassword:            p.ObfsPassword,
		DisableSNI:              p.DisableSni,
		ReduceRtt:               p.ReduceRtt,
		UDPRelayMode:            p.UdpRelayMode,
		CongestionController:    p.CongestionController,
		Multiplex:               p.Multiplex,
		PaddingScheme:           p.PaddingScheme,
		UpMbps:                  p.UpMbps,
		DownMbps:                p.DownMbps,
		Obfs:                    p.Obfs,
		ObfsHost:                p.ObfsHost,
		ObfsPath:                p.ObfsPath,
		XhttpMode:               p.XhttpMode,
		XhttpExtra:              p.XhttpExtra,
		Encryption:              p.Encryption,
		EncryptionMode:          p.EncryptionMode,
		EncryptionRtt:           p.EncryptionRtt,
		EncryptionTicket:        p.EncryptionTicket,
		EncryptionServerPadding: p.EncryptionServerPadding,
		EncryptionPrivateKey:    p.EncryptionPrivateKey,
		EncryptionClientPadding: p.EncryptionClientPadding,
		EncryptionPassword:      p.EncryptionPassword,
		Ratio:                   p.Ratio,
		CertMode:                p.CertMode,
		CertDNSProvider:         p.CertDnsProvider,
		CertDNSEnv:              p.CertDnsEnv,
	}
}

func modelProtocolsToProtos(models []*servermodel.Protocol) []*v1.Protocol {
	if models == nil {
		return nil
	}

	protos := make([]*v1.Protocol, 0, len(models))
	for _, m := range models {
		protos = append(protos, modelProtocolToProto(m))
	}
	return protos
}

func modelProtocolToProto(m *servermodel.Protocol) *v1.Protocol {
	if m == nil {
		return nil
	}

	return &v1.Protocol{
		Type:                    m.Type,
		Port:                    m.Port,
		Enable:                  m.Enable,
		Security:                m.Security,
		Sni:                     m.SNI,
		AllowInsecure:           m.AllowInsecure,
		Fingerprint32:           m.Fingerprint,
		RealityServerAddr:       m.RealityServerAddr,
		RealityServerPort:       m.RealityServerPort,
		RealityPrivateKey:       m.RealityPrivateKey,
		RealityPublicKey:        m.RealityPublicKey,
		RealityShortId:          m.RealityShortId,
		Transport:               m.Transport,
		Host:                    m.Host,
		Path:                    m.Path,
		ServiceName:             m.ServiceName,
		Cipher:                  m.Cipher,
		ServerKey:               m.ServerKey,
		Flow:                    m.Flow,
		HopPorts:                m.HopPorts,
		HopInterval:             m.HopInterval,
		ObfsPassword:            m.ObfsPassword,
		DisableSni:              m.DisableSNI,
		ReduceRtt:               m.ReduceRtt,
		UdpRelayMode:            m.UDPRelayMode,
		CongestionController:    m.CongestionController,
		Multiplex:               m.Multiplex,
		PaddingScheme:           m.PaddingScheme,
		UpMbps:                  m.UpMbps,
		DownMbps:                m.DownMbps,
		Obfs:                    m.Obfs,
		ObfsHost:                m.ObfsHost,
		ObfsPath:                m.ObfsPath,
		XhttpMode:               m.XhttpMode,
		XhttpExtra:              m.XhttpExtra,
		Encryption:              m.Encryption,
		EncryptionMode:          m.EncryptionMode,
		EncryptionRtt:           m.EncryptionRtt,
		EncryptionTicket:        m.EncryptionTicket,
		EncryptionServerPadding: m.EncryptionServerPadding,
		EncryptionPrivateKey:    m.EncryptionPrivateKey,
		EncryptionClientPadding: m.EncryptionClientPadding,
		EncryptionPassword:      m.EncryptionPassword,
		Ratio:                   m.Ratio,
		CertMode:                m.CertMode,
		CertDnsProvider:         m.CertDNSProvider,
		CertDnsEnv:              m.CertDNSEnv,
	}
}
