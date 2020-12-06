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

// GetSessions implements helloworld.GreeterServer
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
