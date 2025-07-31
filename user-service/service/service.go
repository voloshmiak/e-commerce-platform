package service

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"user-service/data"
	pb "user-service/protobuf"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserService) Register(_ context.Context, r *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID := data.AddUser(r.GetEmail(), r.GetPassword(), r.GetFirstName(), r.GetLastName())

	return &pb.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *UserService) Authenticate(_ context.Context, r *pb.AuthRequest) (*pb.AuthResponse, error) {
	_ = data.GetUserByEmail(r.GetEmail())

	return &pb.AuthResponse{
		Token: "token",
	}, nil
}

func (s *UserService) GetProfile(_ context.Context, r *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user := data.GetUserByID(r.GetUserId())
	return &pb.GetUserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
func (s *UserService) UpdateProfile(_ context.Context, r *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	data.UpdateUser(r.GetUserId(), r.GetEmail(), r.GetFirstName(), r.GetLastName())

	return &emptypb.Empty{}, nil
}
