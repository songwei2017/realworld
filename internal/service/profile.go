package service

import (
	"context"
	"realworld/internal/biz"
	"realworld/pkg/middleware/auth"

	pb "realworld/api/profile/v1"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	uc *biz.ProfileUsecase
}

func NewProfileService(uc *biz.ProfileUsecase) *ProfileService {
	return &ProfileService{uc: uc}
}

func (s *ProfileService) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.ProfileReply, error) {
	userId := auth.GetUserIdOrNotLogin(ctx)
	reply, err := s.uc.GetProfile(ctx, userId, req.GetUsername())
	if err != nil {
		return nil, err
	}
	return &pb.ProfileReply{
		Profile: &pb.ProfileReply_Profile{
			Username:  reply.Username,
			Bio:       reply.Bio,
			Image:     reply.Image,
			Following: reply.Following,
		},
	}, nil
}
func (s *ProfileService) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.ProfileReply, error) {
	cu := auth.FromContext(ctx)
	reply, err := s.uc.FollowUser(ctx, cu.UserID, req.GetUsername())
	if err != nil {
		return nil, err
	}
	return &pb.ProfileReply{Profile: &pb.ProfileReply_Profile{
		Username:  reply.Username,
		Bio:       reply.Bio,
		Image:     reply.Image,
		Following: true,
	}}, nil
}
func (s *ProfileService) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.ProfileReply, error) {
	cu := auth.FromContext(ctx)
	reply, err := s.uc.UnFollowUser(ctx, cu.UserID, req.GetUsername())
	if err != nil {
		return nil, err
	}
	return &pb.ProfileReply{Profile: &pb.ProfileReply_Profile{
		Username:  reply.Username,
		Bio:       reply.Bio,
		Image:     reply.Image,
		Following: false,
	}}, nil
}
