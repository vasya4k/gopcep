package grpcapi

import (
	"context"
	"gopcep/controller"
	pb "gopcep/proto"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// GRPCAPI represents api
type GRPCAPI struct {
	sync.RWMutex
	ctr *controller.Controller
	pb.PCEServer
}

// GetSessions aa
func (g *GRPCAPI) GetSessions(ctx context.Context, in *pb.SessionsRequest) (*pb.SessionsReply, error) {
	g.RLock()
	pbSessions := make([]*pb.Session, 0)
	for addr, session := range g.ctr.PCEPSessions {
		pbSessions = append(pbSessions, &pb.Session{
			Address:   addr,
			ID:        string(session.ID),
			MsgCount:  session.MsgCount,
			State:     int32(session.State),
			Keepalive: uint32(session.Keepalive),
			DeadTimer: uint32(session.DeadTimer),
		})
	}
	return &pb.SessionsReply{Sessions: pbSessions}, nil
}

// GetLSPs implements helloworld.GreeterServer
func (g *GRPCAPI) GetLSPs(ctx context.Context, in *pb.LSPRequest) (*pb.LSPReply, error) {
	g.RLock()
	pbLSPs := make([]*pb.LSP, 0)
	session := g.ctr.PCEPSessions[in.PccName]

	for _, lsp := range session.LSPs {
		pbLSPs = append(pbLSPs, &pb.LSP{
			Delegate:     lsp.Delegate,
			Sync:         lsp.Sync,
			Remove:       lsp.Remove,
			Admin:        lsp.Admin,
			Oper:         uint32(lsp.Oper),
			Name:         lsp.Name,
			Src:          lsp.Src,
			Dst:          lsp.Dst,
			SetupPrio:    uint32(lsp.SetupPrio),
			HoldPrio:     uint32(lsp.HoldPrio),
			LocalProtect: lsp.LocalProtect,
			BW:           lsp.BW,
			PLSPID:       lsp.PLSPID,
			LSPID:        uint32(lsp.LSPID),
			SRPID:        lsp.SRPID,
			ExcludeAny:   lsp.ExcludeAny,
			IncludeAny:   lsp.IncludeAny,
			IncludeAll:   lsp.IncludeAll,
		})
	}
	return &pb.LSPReply{LSPs: pbLSPs}, nil
}

// Config represents GRPC Config
type Config struct {
	Address string
	Port    string
}

// Start aaa
func Start(cfg *Config, controller *controller.Controller) *GRPCAPI {
	api := GRPCAPI{
		ctr: controller,
	}

	go func() {
		lis, err := net.Listen("tcp", cfg.Address+":"+cfg.Port)
		if err != nil {
			log.Fatal(err)
		}
		s := grpc.NewServer()
		pb.RegisterPCEServer(s, &api)
		err = s.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return &api
}
