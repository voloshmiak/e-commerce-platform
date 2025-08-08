package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
	"time"
	"user-service/data"
	pb "user-service/protobuf"
)

const (
	secret      = "my-secret-key"
	issuer      = "my-unique-issuer-key"
	expiryHours = 24
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
	user := data.GetUserByEmail(r.GetEmail())

	if user.PasswordHash != r.GetPassword() {
		return nil, errors.New("invalid password")
	}

	claims := jwt.MapClaims{
		"user-id": user.ID,
		"iss":     issuer,
		"exp":     time.Now().Add(expiryHours * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		Token: signedToken,
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, _ *emptypb.Empty) (*pb.GetUserResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	user := data.GetUserByID(int64(userID))
	return &pb.GetUserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
func (s *UserService) UpdateProfile(ctx context.Context, r *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	data.UpdateUser(int64(userID), r.GetEmail(), r.GetFirstName(), r.GetLastName())

	return &emptypb.Empty{}, nil
}

func getUserIDFromContext(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata in context")
	}

	userID := md["user-id"][0]

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user ID: %v", err))
	}

	return userIDInt, nil
}
