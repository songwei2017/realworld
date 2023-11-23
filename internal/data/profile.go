package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"realworld/internal/biz"
)

type ProfileRepo struct {
	data *Data
	log  *log.Helper
}

func NewProfileRepo(data *Data, logger log.Logger) biz.ProfileRepo {
	return &ProfileRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

type Follow struct {
	gorm.Model
	UserId   uint
	FollowId uint
}

func (r *ProfileRepo) GetProfileById(ctx context.Context, uid uint) (rv *biz.Profile, err error) {
	u := new(User)
	res := r.data.db.Where("id = ? ", uid).First(u)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by username")
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return &biz.Profile{
		Username: u.Username,
		Bio:      u.Bio,
		Image:    u.Image,
		Email:    u.Email,
	}, nil
}

func (r *ProfileRepo) GetProfile(ctx context.Context, uid uint, username string) (rv *biz.Profile, err error) {
	u := new(User)
	res := r.data.db.Where("username = ? ", username).First(u)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by username")
	}
	if res.Error != nil {
		return nil, res.Error
	}

	following := false
	if uid > 0 {
		err = r.data.db.Where("user_id = ? ", uid).Where("follow_id", u.ID).First(&Follow{}).Error
		if err == nil {
			following = true
		}
	}
	return &biz.Profile{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: following,
	}, nil
}

func (r *ProfileRepo) FollowUser(ctx context.Context, uid uint, username string) (rv *biz.Profile, err error) {
	u := new(User)
	res := r.data.db.Where("username = ? ", username).First(u)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by username")
	}
	if res.Error != nil {
		return nil, res.Error
	}

	r.data.db.Save(&Follow{
		UserId:   uid,
		FollowId: u.ID,
	})

	return &biz.Profile{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: true,
	}, nil
}

func (r *ProfileRepo) UnFollowUser(ctx context.Context, uid uint, username string) (rv *biz.Profile, err error) {
	u := new(User)
	res := r.data.db.Where("username = ? ", username).First(u)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by username")
	}
	if res.Error != nil {
		return nil, res.Error
	}

	r.data.db.Where("user_id", uid).Where("follow_id", u.ID).Delete(&Follow{})

	return &biz.Profile{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: false,
	}, nil
}
