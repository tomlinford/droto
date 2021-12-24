package viewspb

import (
	context "context"

	"github.com/tomlinford/droto/example/modelspb"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type UserSerializer interface {
	SerializeUser(context.Context, *modelspb.User) (*User, error)
}

type userSerializer struct{}

func (s userSerializer) SerializeUser(ctx context.Context, modelUser *modelspb.User) (*User, error) {
	return UserFromModel(modelUser), nil
}

type UserLoader interface {
	LoadUser(context.Context, *GetUserRequest) (*modelspb.User, error)
	LoadUsers(context.Context, *ListUsersRequest) (
		results []*modelspb.User, nextCursor, prevCursor string, err error)
}

type UserServiceViewset struct {
	UserLoader     UserLoader
	UserSerializer UserSerializer
}

func NewUserServiceViewset() *UserServiceViewset {
	return &UserServiceViewset{UserSerializer: userSerializer{}}
}

func (v *UserServiceViewset) mustEmbedUnimplementedUserServiceServer() {}

var _ UserServiceServer = &UserServiceViewset{}

func (v *UserServiceViewset) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
	if v.UserLoader == nil {
		return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
	}
	userModel, err := v.UserLoader.LoadUser(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return v.UserSerializer.SerializeUser(ctx, userModel)
}

func (v *UserServiceViewset) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	if v.UserLoader == nil {
		return nil, status.Errorf(codes.Unimplemented, "method ListUser not implemented")
	}
	userModels, nextCursor, prevCursor, err := v.UserLoader.LoadUsers(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	results := make([]*User, len(userModels))
	for i, u := range userModels {
		var err error
		results[i], err = v.UserSerializer.SerializeUser(ctx, u)
		if err != nil {
			return nil, err
		}
	}
	next := proto.Clone(req).(*ListUsersRequest)
	next.Cursor = nextCursor
	prev := proto.Clone(req).(*ListUsersRequest)
	prev.Cursor = prevCursor
	return &ListUsersResponse{
		Results: results,
		Next:    next,
		Prev:    prev,
	}, nil
}
