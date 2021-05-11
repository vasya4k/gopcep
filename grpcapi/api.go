package grpcapi

import (
	"context"
	"crypto/tls"
	"gopcep/controller"
	pb "gopcep/proto"
	"net"
	"sync"

	auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	ListenAddr string
	ListenPort string
	Tokens     []string
}

// Start gRPC API
func Start(cfg *Config, controller *controller.Controller) error {
	api := GRPCAPI{
		ctr: controller,
	}
	cert, pool, err := GenCerts()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "grpc_api",
			"event": "gent certs error",
		}).Error(err)
		return err
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientCAs:    pool,
	})
	listener, err := net.Listen("tcp", cfg.ListenAddr+":"+cfg.ListenPort)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic":     "grpc_api",
			"event":     "net listen error",
			"addr_port": cfg.ListenAddr + ":" + cfg.ListenPort,
		}).Error(err)
		return err
	}
	token := Auth{
		Tokens: cfg.Tokens,
	}
	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StreamInterceptor(auth.StreamServerInterceptor(token.TokenAuth)),
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(token.TokenAuth)),
	)

	pb.RegisterPCEServer(server, &api)

	go func() {
		err = server.Serve(listener)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic":     "grpc_api",
				"event":     "serve error",
				"addr_port": cfg.ListenAddr + ":" + cfg.ListenPort,
			}).Fatal(err)
		}
	}()

	return nil
}
