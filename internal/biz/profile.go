package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type ProfileRepo interface {
	GetProfile(ctx context.Context, uid uint, username string) (*Profile, error)
	GetProfileById(ctx context.Context, uid uint) (rv *Profile, err error)
	FollowUser(ctx context.Context, uid uint, username string) (*Profile, error)
	UnFollowUser(ctx context.Context, uid uint, username string) (*Profile, error)
}

type ProfileUsecase struct {
	repo ProfileRepo
	log  *log.Helper
}

func NewProfileUsecase(repo ProfileRepo, logger log.Logger) *ProfileUsecase {
	return &ProfileUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

type Profile struct {
	ID        uint
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Email     string `json:"email"`
	Following bool   `json:"following"`
}

func (s *ProfileUsecase) GetProfile(ctx context.Context, uid uint, username string) (rv *Profile, err error) {
	return s.repo.GetProfile(ctx, uid, username)
}

func (s *ProfileUsecase) FollowUser(ctx context.Context, uid uint, username string) (rv *Profile, err error) {
	return s.repo.FollowUser(ctx, uid, username)
}

func (s *ProfileUsecase) UnFollowUser(ctx context.Context, uid uint, username string) (rv *Profile, err error) {
	return s.repo.UnFollowUser(ctx, uid, username)
}
