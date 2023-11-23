package biz

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"realworld/internal/conf"
	"realworld/pkg/middleware/auth"

	v1 "realworld/api/user/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

func hashPassword(pwd string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func verifyPassword(hashed, input string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input)); err != nil {
		return false
	}
	return true
}
func (uc *UserUsecase) generateToken(userID uint) string {
	fmt.Println("uc.jwtc.Secret")
	fmt.Println(uc.jwtc.Secret)
	return auth.GenerateToken(uc.jwtc.Secret, userID)
}

type User struct {
	Id           uint
	Email        string `json:"email"`
	Token        string `json:"token"`
	Username     string `json:"username"`
	Bio          string `json:"bio"`
	Image        string `json:"image"`
	PasswordHash string `json:"passwordHash"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

type UserRegistration struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

type UpdateUser struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Bio          string `json:"bio"`
	Image        string `json:"image"`
	Password     string `json:"password"`
	PasswordHash string `json:"passwordHash"`
}

// userRepo is a Greater repo.
type UserRepo interface {
	Save(context.Context, *User) (*User, error)
	Update(context.Context, *User) (*User, error)
	FindByID(context.Context, int64) (*User, error)
	ListByHello(context.Context, string) ([]*User, error)
	ListAll(context.Context) ([]*User, error)
	GetUserByEmail(context.Context, string) (*User, error)
	GetUserById(context.Context, uint) (*User, error)
	CreateUser(context.Context, *User) error
	UpdateUser(context.Context, uint, *UpdateUser) error
}

// GreeterUsecase is a Greeter usecase.
type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
	jwtc *conf.JWT
}

// NewGreeterUsecase new a Greeter usecase.
func NewUserUsecase(repo UserRepo, logger log.Logger, jwtc *conf.JWT) *UserUsecase {
	return &UserUsecase{repo: repo, log: log.NewHelper(logger), jwtc: jwtc}
}

// CreateGreeter creates a Greeter, and returns the new Greeter.
func (uc *UserUsecase) CreateUser(ctx context.Context, g *User) (*User, error) {

	uc.log.WithContext(ctx).Infof("CreateGreeter: %v", g.Username)
	return uc.repo.Save(ctx, g)
}

func (uc *UserUsecase) Login(ctx context.Context, email, password string) (*UserLogin, error) {
	if len(email) == 0 {
		return nil, errors.New(422, "email", "cannot be empty")
	}
	if len(password) == 0 {
		return nil, errors.New(422, "password", "password be empty")
	}
	u, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !verifyPassword(u.PasswordHash, password) {
		return nil, errors.Unauthorized("user", "login failed")
	}

	fmt.Println(u)
	fmt.Println("ID:", u.Id)

	return &UserLogin{
		Email:    u.Email,
		Username: u.Username,
		Image:    u.Image,
		Bio:      u.Bio,
		Token:    uc.generateToken(u.Id),
	}, nil
}

func (uc *UserUsecase) Registration(ctx context.Context, email, password string, username string) (*UserRegistration, error) {
	u := &User{
		Email:        email,
		Username:     username,
		PasswordHash: hashPassword(password),
		Image:        "http://img.wxcha.com/m00/f0/f5/5e3999ad5a8d62188ac5ba8ca32e058f.jpg",
	}

	// 检查用户名
	if _, err := uc.repo.GetUserByEmail(ctx, email); err == nil {
		return nil, errors.New(400, "USER_EXITS", "user exits")
	}

	if err := uc.repo.CreateUser(ctx, u); err != nil {
		return nil, err
	}
	return &UserRegistration{
		Email:    email,
		Username: username,
		Token:    uc.generateToken(u.Id),
	}, nil
}
func (uc *UserUsecase) GetCurrentUser(ctx context.Context, id uint) (*User, error) {
	u, err := uc.repo.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{
		Email:    u.Email,
		Username: u.Username,
		Bio:      u.Bio,
		Image:    u.Image,
		Token:    uc.generateToken(id),
	}, nil
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, id uint, uu *UpdateUser) (*User, error) {
	if len(uu.Password) > 0 {
		uu.PasswordHash = hashPassword(uu.Password)
	}
	if err := uc.repo.UpdateUser(ctx, id, uu); err != nil {
		return nil, err
	}
	return &User{
		Email:    uu.Email,
		Username: uu.Username,
		Image:    uu.Image,
		Bio:      uu.Bio,
		Token:    uc.generateToken(id),
	}, nil
}
