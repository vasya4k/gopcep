package grpcapi

import (
	"context"
	"gopcep/pcep"
	pb "gopcep/proto"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// GRPCAPI represents api
type GRPCAPI struct {
	sync.RWMutex
	PSessions map[string]*pcep.Session
	pb.PCEServer
}

// LoadPSessions aa
func (g *GRPCAPI) LoadPSessions(key string) (*pcep.Session, bool) {
	g.RLock()
	result, ok := g.PSessions[key]
	g.RUnlock()
	return result, ok
}

// DeletePSessions aa
func (g *GRPCAPI) DeletePSessions(key string) {
	g.Lock()
	delete(g.PSessions, key)
	g.Unlock()
}

// StorePSessions aa
func (g *GRPCAPI) StorePSessions(key string, value *pcep.Session) *pcep.Session {
	g.Lock()
	g.PSessions[key] = value
	g.Unlock()
	return value
}

// GetSessions implements helloworld.GreeterServer
func (g *GRPCAPI) GetSessions(ctx context.Context, in *pb.SessionsRequest) (*pb.SessionsReply, error) {
	g.RLock()
	pbSessions := make([]*pb.Session, 0)
	for addr, session := range g.PSessions {
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
func Start(cfg *Config) *GRPCAPI {
	api := GRPCAPI{
		PSessions: make(map[string]*pcep.Session),
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
