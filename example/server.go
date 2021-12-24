package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/tomlinford/droto/example/modelspb"
	"github.com/tomlinford/droto/example/viewspb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type memoryUserLoader struct{}

func (l memoryUserLoader) LoadUser(ctx context.Context,
	req *viewspb.GetUserRequest) (*modelspb.User, error) {
	return &modelspb.User{
		Id:       req.Id,
		Username: fmt.Sprintf("test%d", req.Id),
		AboutMe:  &wrapperspb.StringValue{Value: "I'm a user."},
	}, nil
}

type userSerializer struct{ base viewspb.UserSerializer }

func (s *userSerializer) SerializeUser(ctx context.Context, modelUser *modelspb.User) (*viewspb.User, error) {
	user, err := s.base.SerializeUser(ctx, modelUser)
	if err != nil {
		return nil, err
	}
	user.Username = user.Username[:3] + "***"
	return user, nil
}

func (l memoryUserLoader) LoadUsers(ctx context.Context,
	req *viewspb.ListUsersRequest) (
	results []*modelspb.User, nextCursor, prevCursor string, err error) {
	u, _ := l.LoadUser(ctx, &viewspb.GetUserRequest{Id: 1})
	return []*modelspb.User{u}, "", "", nil
}

func main() {
	server := viewspb.NewUserServiceViewset()
	server.UserLoader = memoryUserLoader{}
	server.UserSerializer = &userSerializer{server.UserSerializer}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 4242))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go runClient()
	s := grpc.NewServer()
	viewspb.RegisterUserServiceServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runClient() {
	conn, err := grpc.Dial("localhost:4242", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := viewspb.NewUserServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetUser(ctx, &viewspb.GetUserRequest{Id: 44})
	if err != nil {
		log.Fatalf("could not get user: %v", err)
	}
	log.Printf("Response: %s", r)
}
