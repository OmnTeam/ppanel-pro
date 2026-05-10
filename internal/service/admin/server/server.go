package server

import (
	"context"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/server/v1"
	serverbiz "github.com/OmnTeam/npanel-pro/internal/biz/admin/server"
	servermodel "github.com/OmnTeam/npanel-pro/internal/model/server"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type ServerService struct {
	v1.UnimplementedServerServiceServer

	serverUc    *serverbiz.ServerUsecase
	nodeUc      *serverbiz.NodeUsecase
	migrationUc *serverbiz.MigrationUsecase
	log         *log.Helper
}

func NewServerService(serverUc *serverbiz.ServerUsecase, nodeUc *serverbiz.NodeUsecase, migrationUc *serverbiz.MigrationUsecase, logger log.Logger) *ServerService {
	return &ServerService{serverUc: serverUc, nodeUc: nodeUc, migrationUc: migrationUc, log: log.NewHelper(logger)}
}

func (s *ServerService) CreateServer(ctx context.Context, req *v1.CreateServerRequest) (*v1.CreateServerReply, error) {
	protocols := protosToModelProtocols(req.Protocols)
	server, err := s.serverUc.CreateServer(ctx, req.Name, req.Country, req.City, req.Address, int64(req.Sort), protocols)
	if err != nil {
		return nil, err
	}
	return &v1.CreateServerReply{Code: responsecode.AdminCreateServerSuccess, Message: responsecode.CodeMessages[responsecode.AdminCreateServerSuccess], Data: &v1.CreateServerData{Server: serverToProto(server)}}, nil
}

func (s *ServerService) UpdateServer(ctx context.Context, req *v1.UpdateServerRequest) (*v1.UpdateServerReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	protocols := protosToModelProtocols(req.Protocols)
	server, err := s.serverUc.UpdateServer(ctx, int(req.Id), req.Name, req.Country, req.City, req.Address, int64(req.Sort), protocols)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateServerReply{Code: responsecode.AdminUpdateServerSuccess, Message: responsecode.CodeMessages[responsecode.AdminUpdateServerSuccess], Data: &v1.UpdateServerData{Server: serverToProto(server)}}, nil
}

func (s *ServerService) DeleteServer(ctx context.Context, req *v1.DeleteServerRequest) (*v1.DeleteServerReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	if err := s.serverUc.DeleteServer(ctx, int(req.Id)); err != nil {
		return nil, err
	}
	return &v1.DeleteServerReply{Code: responsecode.AdminDeleteServerSuccess, Message: responsecode.CodeMessages[responsecode.AdminDeleteServerSuccess], Data: &v1.DeleteServerData{Success: true}}, nil
}

func (s *ServerService) FilterServerList(ctx context.Context, req *v1.FilterServerListRequest) (*v1.FilterServerListReply, error) {
	total, servers, err := s.serverUc.FilterServerList(ctx, int32(req.Page), int32(req.Size), req.Search)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.Server, 0, len(servers))
	for _, server := range servers {
		list = append(list, serverToProto(server))
	}
	return &v1.FilterServerListReply{Code: responsecode.AdminFilterServerListSuccess, Message: responsecode.CodeMessages[responsecode.AdminFilterServerListSuccess], Data: &v1.FilterServerListData{Total: total, List: list}}, nil
}

func (s *ServerService) GetServerProtocols(ctx context.Context, req *v1.GetServerProtocolsRequest) (*v1.GetServerProtocolsReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	protocols, err := s.serverUc.GetServerProtocols(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &v1.GetServerProtocolsReply{Code: responsecode.AdminGetServerProtocolsSuccess, Message: responsecode.CodeMessages[responsecode.AdminGetServerProtocolsSuccess], Data: &v1.GetServerProtocolsData{Protocols: modelProtocolsToProtos(protocols)}}, nil
}

func (s *ServerService) CreateNode(ctx context.Context, req *v1.CreateNodeRequest) (*v1.CreateNodeReply, error) {
	if req.ServerId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	node, err := s.nodeUc.CreateNode(ctx, req.Name, req.Tags, uint16(req.Port), req.Address, req.ServerId, req.Protocol, req.Enabled, req.NodeType, req.IsHidden, req.NodeGroupIds)
	if err != nil {
		return nil, err
	}
	return &v1.CreateNodeReply{Code: responsecode.AdminCreateNodeSuccess, Message: responsecode.CodeMessages[responsecode.AdminCreateNodeSuccess], Data: &v1.CreateNodeData{Node: nodeToProto(node)}}, nil
}

func (s *ServerService) UpdateNode(ctx context.Context, req *v1.UpdateNodeRequest) (*v1.UpdateNodeReply, error) {
	if req.Id <= 0 || req.ServerId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	node, err := s.nodeUc.UpdateNode(ctx, int(req.Id), req.Name, req.Tags, uint16(req.Port), req.Address, req.ServerId, req.Protocol, req.Enabled, req.NodeType, req.IsHidden, req.NodeGroupIds)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNodeReply{Code: responsecode.AdminUpdateNodeSuccess, Message: responsecode.CodeMessages[responsecode.AdminUpdateNodeSuccess], Data: &v1.UpdateNodeData{Node: nodeToProto(node)}}, nil
}

func (s *ServerService) DeleteNode(ctx context.Context, req *v1.DeleteNodeRequest) (*v1.DeleteNodeReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	if err := s.nodeUc.DeleteNode(ctx, int(req.Id)); err != nil {
		return nil, err
	}
	return &v1.DeleteNodeReply{Code: responsecode.AdminDeleteNodeSuccess, Message: responsecode.CodeMessages[responsecode.AdminDeleteNodeSuccess], Data: &v1.DeleteNodeData{Success: true}}, nil
}

func (s *ServerService) FilterNodeList(ctx context.Context, req *v1.FilterNodeListRequest) (*v1.FilterNodeListReply, error) {
	total, nodes, err := s.nodeUc.FilterNodeList(ctx, int32(req.Page), int32(req.Size), req.Search, req.NodeGroupId)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.Node, 0, len(nodes))
	for _, node := range nodes {
		list = append(list, nodeToProto(node))
	}
	return &v1.FilterNodeListReply{Code: responsecode.AdminFilterNodeListSuccess, Message: responsecode.CodeMessages[responsecode.AdminFilterNodeListSuccess], Data: &v1.FilterNodeListData{Total: total, List: list}}, nil
}

func (s *ServerService) ToggleNodeStatus(ctx context.Context, req *v1.ToggleNodeStatusRequest) (*v1.ToggleNodeStatusReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	node, err := s.nodeUc.ToggleNodeStatus(ctx, int(req.Id), req.Enable)
	if err != nil {
		return nil, err
	}
	return &v1.ToggleNodeStatusReply{Code: responsecode.AdminToggleNodeStatusSuccess, Message: responsecode.CodeMessages[responsecode.AdminToggleNodeStatusSuccess], Data: &v1.ToggleNodeStatusData{Node: nodeToProto(node)}}, nil
}

func (s *ServerService) QueryNodeTag(ctx context.Context, req *v1.QueryNodeTagRequest) (*v1.QueryNodeTagReply, error) {
	tags, err := s.nodeUc.QueryNodeTags(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.QueryNodeTagReply{Code: responsecode.AdminQueryNodeTagSuccess, Message: responsecode.CodeMessages[responsecode.AdminQueryNodeTagSuccess], Data: &v1.QueryNodeTagData{Tags: tags}}, nil
}

func (s *ServerService) HasMigrateServerNode(ctx context.Context, req *v1.HasMigrateServerNodeRequest) (*v1.HasMigrateServerNodeReply, error) {
	hasMigrate, err := s.migrationUc.HasMigrateServerNode(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.HasMigrateServerNodeReply{Code: responsecode.AdminHasMigrateServerNodeSuccess, Message: responsecode.CodeMessages[responsecode.AdminHasMigrateServerNodeSuccess], Data: &v1.HasMigrateServerNodeData{HasMigrate: hasMigrate}}, nil
}

func (s *ServerService) MigrateServerNode(ctx context.Context, req *v1.MigrateServerNodeRequest) (*v1.MigrateServerNodeReply, error) {
	success, fail, _, err := s.migrationUc.MigrateServerNode(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.MigrateServerNodeReply{Code: responsecode.AdminMigrateServerNodeSuccess, Message: responsecode.CodeMessages[responsecode.AdminMigrateServerNodeSuccess], Data: &v1.MigrateServerNodeData{Success: success, Fail: fail}}, nil
}

func (s *ServerService) ResetSortWithServer(ctx context.Context, req *v1.ResetSortRequest) (*v1.ResetSortReply, error) {
	sortItems := make([]*serverbiz.SortItem, 0, len(req.Sort))
	for _, item := range req.Sort {
		if item.Id <= 0 {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		sortItems = append(sortItems, &serverbiz.SortItem{ID: item.Id, Sort: int(item.Sort)})
	}
	if err := s.serverUc.ResetServerSort(ctx, sortItems); err != nil {
		return nil, err
	}
	return &v1.ResetSortReply{Code: responsecode.AdminResetSortWithServerSuccess, Message: responsecode.CodeMessages[responsecode.AdminResetSortWithServerSuccess], Data: &v1.ResetSortData{Success: true}}, nil
}

func (s *ServerService) ResetSortWithNode(ctx context.Context, req *v1.ResetSortRequest) (*v1.ResetSortReply, error) {
	sortItems := make([]*serverbiz.SortItem, 0, len(req.Sort))
	for _, item := range req.Sort {
		if item.Id <= 0 {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		sortItems = append(sortItems, &serverbiz.SortItem{ID: item.Id, Sort: int(item.Sort)})
	}
	if err := s.nodeUc.ResetNodeSort(ctx, sortItems); err != nil {
		return nil, err
	}
	return &v1.ResetSortReply{Code: responsecode.AdminResetSortWithNodeSuccess, Message: responsecode.CodeMessages[responsecode.AdminResetSortWithNodeSuccess], Data: &v1.ResetSortData{Success: true}}, nil
}

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
				ips = append(ips, &v1.ServerOnlineIP{Ip: ip.IP, Protocol: ip.Protocol})
			}
			onlineUsers = append(onlineUsers, &v1.ServerOnlineUser{Ip: ips, UserId: user.UserID, Subscribe: user.Subscribe, SubscribeId: user.SubscribeID, Traffic: user.Traffic, ExpiredAt: user.ExpiredAt})
		}
		status = &v1.ServerStatus{Cpu: s.Status.Cpu, Mem: s.Status.Mem, Disk: s.Status.Disk, Online: onlineUsers, Status: s.Status.Status}
	}
	return &v1.Server{Id: s.ID, Name: s.Name, Country: s.Country, City: s.City, Address: s.Address, Sort: int32(s.Sort), Protocols: modelProtocolsToProtos(s.Protocols), LastReportedAt: s.LastReportedAt, Status: status, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func nodeToProto(n *serverbiz.Node) *v1.Node {
	if n == nil {
		return nil
	}
	return &v1.Node{Id: n.ID, Name: n.Name, Tags: n.Tags, Port: uint32(n.Port), Address: n.Address, ServerId: n.ServerID, Protocol: n.Protocol, Enabled: n.Enabled, Sort: int32(n.Sort), NodeGroupId: n.NodeGroupID, NodeGroupIds: n.NodeGroupIDs, NodeType: n.NodeType, IsHidden: n.IsHidden, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt}
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
	return &servermodel.Protocol{Type: p.Type, Port: p.Port, Enable: p.Enable, Security: p.Security, SNI: p.Sni, AllowInsecure: p.AllowInsecure, Fingerprint: p.Fingerprint, RealityServerAddr: p.RealityServerAddr, RealityServerPort: p.RealityServerPort, RealityPrivateKey: p.RealityPrivateKey, RealityPublicKey: p.RealityPublicKey, RealityShortId: p.RealityShortId, Transport: p.Transport, Host: p.Host, Path: p.Path, ServiceName: p.ServiceName, Cipher: p.Cipher, ServerKey: p.ServerKey, Flow: p.Flow, HopPorts: p.HopPorts, HopInterval: p.HopInterval, ObfsPassword: p.ObfsPassword, DisableSNI: p.DisableSni, ReduceRtt: p.ReduceRtt, UDPRelayMode: p.UdpRelayMode, CongestionController: p.CongestionController, Multiplex: p.Multiplex, PaddingScheme: p.PaddingScheme, UpMbps: p.UpMbps, DownMbps: p.DownMbps, Obfs: p.Obfs, ObfsHost: p.ObfsHost, ObfsPath: p.ObfsPath, XhttpMode: p.XhttpMode, XhttpExtra: p.XhttpExtra, Encryption: p.Encryption, EncryptionMode: p.EncryptionMode, EncryptionRtt: p.EncryptionRtt, EncryptionTicket: p.EncryptionTicket, EncryptionServerPadding: p.EncryptionServerPadding, EncryptionPrivateKey: p.EncryptionPrivateKey, EncryptionClientPadding: p.EncryptionClientPadding, EncryptionPassword: p.EncryptionPassword, Ratio: p.Ratio, CertMode: p.CertMode, CertDNSProvider: p.CertDnsProvider, CertDNSEnv: p.CertDnsEnv, SimnetPsk: p.SimnetPsk, SimnetKeyID: p.SimnetKeyId, SimnetTicketID: p.SimnetTicketId, SimnetPath: p.SimnetPath, SimnetCarrier: p.SimnetCarrier, SimnetAfEnabled: p.SimnetAfEnabled, SimnetAfPathMode: p.SimnetAfPathMode, SimnetAfPathPrefix: p.SimnetAfPathPrefix, SimnetAfPathSuffix: p.SimnetAfPathSuffix, SimnetAfMagicMode: p.SimnetAfMagicMode, SimnetAfResponseJitterMs: p.SimnetAfResponseJitterMs, SimnetAfHandshakePolymorphism: p.SimnetAfHandshakePolymorphism, SimnetAfSettingsJitter: p.SimnetAfSettingsJitter, SimnetAfFakeHeaderInjection: p.SimnetAfFakeHeaderInjection, SimnetReverseEnabled: p.SimnetReverseEnabled, SimnetReverseListenAddr: p.SimnetReverseListenAddr, SimnetReverseListenPort: p.SimnetReverseListenPort, SimnetReverseTargetHost: p.SimnetReverseTargetHost, SimnetReverseTargetPort: p.SimnetReverseTargetPort}
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
	return &v1.Protocol{Type: m.Type, Port: m.Port, Enable: m.Enable, Security: m.Security, Sni: m.SNI, AllowInsecure: m.AllowInsecure, Fingerprint: m.Fingerprint, RealityServerAddr: m.RealityServerAddr, RealityServerPort: m.RealityServerPort, RealityPrivateKey: m.RealityPrivateKey, RealityPublicKey: m.RealityPublicKey, RealityShortId: m.RealityShortId, Transport: m.Transport, Host: m.Host, Path: m.Path, ServiceName: m.ServiceName, Cipher: m.Cipher, ServerKey: m.ServerKey, Flow: m.Flow, HopPorts: m.HopPorts, HopInterval: m.HopInterval, ObfsPassword: m.ObfsPassword, DisableSni: m.DisableSNI, ReduceRtt: m.ReduceRtt, UdpRelayMode: m.UDPRelayMode, CongestionController: m.CongestionController, Multiplex: m.Multiplex, PaddingScheme: m.PaddingScheme, UpMbps: m.UpMbps, DownMbps: m.DownMbps, Obfs: m.Obfs, ObfsHost: m.ObfsHost, ObfsPath: m.ObfsPath, XhttpMode: m.XhttpMode, XhttpExtra: m.XhttpExtra, Encryption: m.Encryption, EncryptionMode: m.EncryptionMode, EncryptionRtt: m.EncryptionRtt, EncryptionTicket: m.EncryptionTicket, EncryptionServerPadding: m.EncryptionServerPadding, EncryptionPrivateKey: m.EncryptionPrivateKey, EncryptionClientPadding: m.EncryptionClientPadding, EncryptionPassword: m.EncryptionPassword, Ratio: m.Ratio, CertMode: m.CertMode, CertDnsProvider: m.CertDNSProvider, CertDnsEnv: m.CertDNSEnv, SimnetPsk: m.SimnetPsk, SimnetKeyId: m.SimnetKeyID, SimnetTicketId: m.SimnetTicketID, SimnetPath: m.SimnetPath, SimnetCarrier: m.SimnetCarrier, SimnetAfEnabled: m.SimnetAfEnabled, SimnetAfPathMode: m.SimnetAfPathMode, SimnetAfPathPrefix: m.SimnetAfPathPrefix, SimnetAfPathSuffix: m.SimnetAfPathSuffix, SimnetAfMagicMode: m.SimnetAfMagicMode, SimnetAfResponseJitterMs: m.SimnetAfResponseJitterMs, SimnetAfHandshakePolymorphism: m.SimnetAfHandshakePolymorphism, SimnetAfSettingsJitter: m.SimnetAfSettingsJitter, SimnetAfFakeHeaderInjection: m.SimnetAfFakeHeaderInjection, SimnetReverseEnabled: m.SimnetReverseEnabled, SimnetReverseListenAddr: m.SimnetReverseListenAddr, SimnetReverseListenPort: m.SimnetReverseListenPort, SimnetReverseTargetHost: m.SimnetReverseTargetHost, SimnetReverseTargetPort: m.SimnetReverseTargetPort}
}
