package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gopcep/grpcapi"
	pb "gopcep/proto"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Auth struct {
	Token string
}

func (a Auth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + a.Token,
	}, nil
}

func (Auth) RequireTransportSecurity() bool {
	return true
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, pool, err := grpcapi.GenCerts()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "grpc_api",
			"event": "net dial error",
		}).Fatal(err)
	}
	// creds := credentials.NewClientTLSFromCert(pool, "127.0.0.1")
	creds := credentials.NewTLS(&tls.Config{
		// ServerName:         "",
		InsecureSkipVerify: true,
		RootCAs:            pool,
	})
	conn, err := grpc.DialContext(ctx, "127.0.0.1:12345",
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(Auth{
			Token: "boom",
		}),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPCEClient(conn)

	r, err := c.GetSessions(ctx, &pb.SessionsRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	PrintAsJSON(r)
}

//PrintAsJSON Prints anything as JSON
// fmt.Printf("Data: %08b \n", data[:4])
func PrintAsJSON(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}
