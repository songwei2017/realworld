package service

import (
	"context"
	"fmt"
	v1 "realworld/api/user/v1"
	"realworld/internal/biz"
	"realworld/pkg/middleware/auth"
)

// GreeterService is a greeter service.
type UserService struct {
	v1.UnimplementedUserServer

	uc *biz.UserUsecase
}

// NewUserService new a User service.
func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) Authentication(ctx context.Context, in *v1.AuthenticationRequest) (*v1.AuthenticationReply, error) {
	rv, err := s.uc.Login(ctx, in.User.Email, in.User.Password)
	if err != nil {
		return nil, err
	}
	return &v1.AuthenticationReply{
		User: &v1.AuthenticationReply_User{
			Username: rv.Username,
			Email:    rv.Email,
			Token:    rv.Token,
			Bio:      rv.Bio,
			Image:    rv.Image,
		},
	}, nil

}

func (s *UserService) Registration(ctx context.Context, in *v1.RegistrationRequest) (*v1.RegistrationReply, error) {

	rv, err := s.uc.Registration(ctx, in.User.Email, in.User.Password, in.User.Username)
	if err != nil {
		return nil, err
	}

	return &v1.RegistrationReply{
		User: &v1.RegistrationReply_User{
			Email:    rv.Email,
			Token:    rv.Token,
			Username: rv.Username,
			Bio:      rv.Bio,
			Image:    rv.Image,
		},
	}, nil

}

func (s *UserService) GetCurrentUser(ctx context.Context, in *v1.GetCurrentUserRequest) (*v1.GetCurrentUserReply, error) {
	cu := auth.FromContext(ctx)
	fmt.Println("cu", cu)
	rv, err := s.uc.GetCurrentUser(ctx, cu.UserID)
	if err != nil {
		return nil, err
	}

	return &v1.GetCurrentUserReply{
		User: &v1.GetCurrentUserReply_User{
			Email:    rv.Email,
			Token:    rv.Token,
			Username: rv.Username,
			Bio:      rv.Bio,
			Image:    rv.Image,
		},
	}, nil

}

func (s *UserService) UpdateUser(ctx context.Context, in *v1.UpdateUserRequest) (*v1.UpdateUserReply, error) {
	cu := auth.FromContext(ctx)
	uu := &biz.UpdateUser{
		Email:    in.User.GetEmail(),
		Bio:      in.User.GetBio(),
		Image:    in.User.GetImage(),
		Username: in.User.GetUsername(),
	}

	if len(in.User.GetPassword()) > 0 {
		uu.Password = in.User.GetPassword()
	}

	rv, err := s.uc.UpdateUser(ctx, cu.UserID, uu)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateUserReply{
		User: &v1.UpdateUserReply_User{
			Email:    rv.Email,
			Image:    rv.Image,
			Token:    rv.Token,
			Bio:      rv.Bio,
			Username: rv.Username,
		}}, nil
}
