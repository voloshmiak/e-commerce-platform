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
	"strings"
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

	fmt.Println(user)

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
	mt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("missing metadata in context")
	}

	bearedToken := mt.Get("Authorization")[0]

	tokenString := strings.TrimPrefix(bearedToken, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	claims := token.Claims.(jwt.MapClaims)

	userID, ok := claims["user-id"]
	if !ok {
		return 0, fmt.Errorf("user-id not found in token claims")
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		return 0, fmt.Errorf("user-id is not a valid float64")
	}

	userIDInt := int(userIDFloat)

	return userIDInt, nil
}
