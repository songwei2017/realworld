package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/go-kratos/kratos/v2/log"
	"realworld/internal/biz"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

type User struct {
	gorm.Model
	Email        string `gorm:"size:500;unique"`
	Username     string `gorm:"size:500;unique"`
	Bio          string `gorm:"size:1000"`
	Image        string `gorm:"size:1000"`
	PasswordHash string `gorm:"size:500"`
	Following    uint32
}

// NewGreeterRepo .
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) Save(ctx context.Context, g *biz.User) (*biz.User, error) {

	return g, nil
}

func (r *userRepo) Update(ctx context.Context, g *biz.User) (*biz.User, error) {
	return g, nil
}

func (r *userRepo) FindByID(context.Context, int64) (*biz.User, error) {
	return nil, nil
}

func (r *userRepo) ListByHello(context.Context, string) ([]*biz.User, error) {
	return nil, nil
}

func (r *userRepo) ListAll(context.Context) ([]*biz.User, error) {
	return nil, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*biz.User, error) {
	u := new(User)
	res := r.data.db.Where("email = ? ", email).First(u)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by email")
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &biz.User{
		Email:        u.Email,
		Username:     u.Username,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
		Id:           u.ID,
	}, nil
}
func (r *userRepo) GetUserById(ctx context.Context, id uint) (*biz.User, error) {
	u := new(User)
	res := r.data.db.Where("id = ?", id).First(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	return &biz.User{
		Email:        u.Email,
		Username:     u.Username,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
	}, nil
}
func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) error {
	user := User{
		Email:        u.Email,
		Username:     u.Username,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
	}
	rv := r.data.db.Create(&user)
	if rv.Error == nil {
		fmt.Println("ID:", user.ID)
		u.Id = user.ID
	}
	return rv.Error
}

func (r *userRepo) UpdateUser(ctx context.Context, id uint, u *biz.UpdateUser) error {
	user := User{
		Email:        u.Email,
		Username:     u.Username,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
	}
	if err := r.data.db.Where("id = ? ", id).Updates(&user).Error; err != nil {
		return err
	}
	// 更新所有文章的信息
	r.data.db.Where("author_id = ? ", id).Updates(&Article{Email: u.Email, Username: u.Username, Image: u.Image})
	return nil
}
