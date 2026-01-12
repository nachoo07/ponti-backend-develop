package user

// import (
// 	"context"

// 	pb "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/grpc/pb"
// )

// type Server interface {
// 	GetUserUUID(context.Context, *pb.GetUserRequest) (*pb.GetUserResponse, error)
// }

// type Grpc struct {
// 	pb.UnimplementedUserServiceServer
// 	ucs UseCases
// }

// func NewGrpc(ucs UseCases) Server {
// 	return &Grpc{
// 		ucs: ucs,
// 	}
// }

// func (s *Grpc) GetUserByCrentials(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
// 	userID, err := s.ucs.GetUserByCredentials(ctx, req.Username, req.PasswordHash)
// 	if err != nil {
// 		return &pb.GetUserResponse{}, err
// 	}

// 	_ = userID
// 	UUID := "userID"
// 	return &pb.GetUserResponse{
// 		UUID: UUID,
// 	}, nil
// }
