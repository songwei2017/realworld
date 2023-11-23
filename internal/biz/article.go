package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"realworld/pkg/middleware/auth"
	"regexp"
	"strings"
	"time"
)

type ArticleUsecase struct {
	repo ArticleRepo
	log  *log.Helper
}

func NewArticleUsecase(repo ArticleRepo, logger log.Logger) *ArticleUsecase {
	return &ArticleUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

type ArticleRepo interface {
	List(ctx context.Context, los ListOptions, opts ...DbOption) ([]*Article, int64, error)
	Get(ctx context.Context, slug string) (*Article, error)
	Create(ctx context.Context, a *Article) (*Article, error)
	Update(ctx context.Context, a *Article) (*Article, error)
	Delete(ctx context.Context, a *Article) error
	GetArticle(ctx context.Context, aid uint) (*Article, error)
	CheckFavorited(uid uint, id uint) bool
	Favorite(ctx context.Context, currentUserID uint, aid uint) error
	Unfavorite(ctx context.Context, currentUserID uint, aid uint) error
	GetFavoritesStatus(ctx context.Context, currentUserID uint, as []*Article) (favorited []bool, err error)

	ListTags(ctx context.Context) ([]Tag, error)
}

type CommentRepo interface {
	Create(ctx context.Context, c *Comment) (*Comment, error)
	Get(ctx context.Context, id uint) (*Comment, error)
	List(ctx context.Context, slug string) ([]*Comment, error)
	Delete(ctx context.Context, id uint) error
}

type SocialUsecase struct {
	ar ArticleRepo
	cr CommentRepo
	pr ProfileRepo

	log *log.Helper
}

type Article struct {
	ID             uint
	Slug           string
	Title          string
	Description    string
	Body           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	TagList        []string
	Favorited      bool
	FavoritesCount uint32

	AuthorUserID uint

	Author   *Profile
	Username string
	Email    string
	Image    string
}

type Comment struct {
	ID        uint
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time

	ArticleID uint

	Article  *Article
	AuthorID uint
	Author   *Profile
}

type Tag string

func slugify(title string) string {
	re, _ := regexp.Compile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(title, "-"))
}

func (o *Article) verifyAuthor(id uint) bool {
	return o.AuthorUserID == id
}

func (o *Comment) verifyAuthor(id uint) bool {
	return o.AuthorID == id
}

func NewSocialUsecase(
	ar ArticleRepo,
	pr ProfileRepo,
	cr CommentRepo,
	logger log.Logger) *SocialUsecase {
	return &SocialUsecase{ar: ar, cr: cr, pr: pr, log: log.NewHelper(logger)}
}

func (uc *SocialUsecase) GetArticle(ctx context.Context, slug string) (rv *Article, err error) {
	return uc.ar.Get(ctx, slug)
}

func (uc *SocialUsecase) CreateArticle(ctx context.Context, in *Article) (rv *Article, err error) {
	u := auth.FromContext(ctx)
	in.Slug = slugify(in.Title)
	in.AuthorUserID = u.UserID

	AuthorUser, err := uc.pr.GetProfileById(ctx, u.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	in.Username = AuthorUser.Username
	in.Email = AuthorUser.Email
	in.Image = AuthorUser.Image
	a, err := uc.ar.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return a, err
}

func (uc *SocialUsecase) DeleteArticle(ctx context.Context, slug string) (err error) {
	a, err := uc.ar.Get(ctx, slug)
	if err != nil {
		return err
	}
	if !a.verifyAuthor(auth.FromContext(ctx).UserID) {
		return errors.Unauthorized("user", "verifyAuthor fail")
	}
	return uc.ar.Delete(ctx, a)
}

func (uc *SocialUsecase) AddComment(ctx context.Context, slug string, in *Comment) (rv *Comment, err error) {
	u := auth.FromContext(ctx)
	in.AuthorID = u.UserID
	in.Article = &Article{Slug: slug}
	return uc.cr.Create(ctx, in)
}

func (uc *SocialUsecase) ListComments(ctx context.Context, slug string) (rv []*Comment, err error) {
	return uc.cr.List(ctx, slug)
}

func (uc *SocialUsecase) DeleteComment(ctx context.Context, id uint) (err error) {
	a, err := uc.cr.Get(ctx, id)
	if err != nil {
		return err
	}
	if !a.verifyAuthor(auth.FromContext(ctx).UserID) {
		return errors.Unauthorized("user", "verifyAuthor fail")
	}
	err = uc.cr.Delete(ctx, id)
	return err
}

func (uc *SocialUsecase) FeedArticles(ctx context.Context, opts ...DbOption) (rv []*Article, count int64, err error) {
	rv, count, err = uc.ar.List(ctx, ListOptions{}, opts...)
	if err != nil {
		return nil, 0, err
	}
	return rv, count, nil
}

func (uc *SocialUsecase) ListArticles(ctx context.Context, los ListOptions, opts ...DbOption) (rv []*Article, count int64, err error) {
	uid := auth.GetUserIdOrNotLogin(ctx)
	rv, count, err = uc.ar.List(ctx, los, opts...)

	if err != nil {
		return nil, 0, err
	}

	if uid > 0 {
		for i, article := range rv {
			rv[i].Favorited = uc.ar.CheckFavorited(uid, article.ID)
		}
	}

	return rv, count, nil
}

func (uc *SocialUsecase) UpdateArticle(ctx context.Context, in *Article) (rv *Article, err error) {
	a, err := uc.ar.Get(ctx, in.Slug)
	if err != nil {
		return nil, err
	}
	if !a.verifyAuthor(auth.FromContext(ctx).UserID) {
		return nil, errors.Unauthorized("user", "verifyAuthor fail")
	}
	rv, err = uc.ar.Update(ctx, in)
	return rv, err
}

func (uc *SocialUsecase) GetTags(ctx context.Context) (rv []Tag, err error) {
	return uc.ar.ListTags(ctx)
}

func (uc *SocialUsecase) FavoriteArticle(ctx context.Context, slug string) (rv *Article, err error) {
	a, err := uc.ar.Get(ctx, slug)
	if err != nil {
		return nil, err
	}
	cu := auth.FromContext(ctx)
	err = uc.ar.Favorite(ctx, cu.UserID, a.ID)
	if err != nil {
		return nil, err
	}
	a, err = uc.ar.GetArticle(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	a.Favorited = true
	return a, nil
}

func (uc *SocialUsecase) UnfavoriteArticle(ctx context.Context, slug string) (rv *Article, err error) {
	a, err := uc.ar.Get(ctx, slug)
	if err != nil {
		return nil, err
	}
	cu := auth.FromContext(ctx)
	err = uc.ar.Unfavorite(ctx, cu.UserID, a.ID)
	if err != nil {
		return nil, err
	}

	a.Favorited = false
	return a, nil
}
