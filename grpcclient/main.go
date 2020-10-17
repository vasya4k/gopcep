package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "gopcep/proto"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:12345", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPCEClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
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
