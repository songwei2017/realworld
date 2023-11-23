package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"realworld/internal/biz"
)

type Article struct {
	gorm.Model
	Slug           string `gorm:"size:200"`
	Title          string `gorm:"size:200;unique"`
	Description    string `gorm:"size:200"`
	Body           string
	Tags           []Tag `gorm:"many2many:article_tags;"`
	AuthorID       uint
	FavoritesCount uint32
	Username       string `gorm:"size:200"`
	Email          string `gorm:"size:200"`
	Image          string `gorm:"size:200"`
}

type Tag struct {
	gorm.Model
	Name     string    `gorm:"size:200;uniqueIndex"`
	Articles []Article `gorm:"many2many:article_tags;"`
}

type ArticleFavorite struct {
	gorm.Model
	UserID    uint
	ArticleID uint
}

type articleRepo struct {
	data *Data
	log  *log.Helper
}

func convertArticle(x Article) *biz.Article {

	tag := make([]string, 0, len(x.Tags))
	for _, t := range x.Tags {
		tag = append(tag, t.Name)
	}
	return &biz.Article{
		ID:             x.ID,
		Slug:           x.Slug,
		Title:          x.Title,
		Description:    x.Description,
		AuthorUserID:   x.AuthorID,
		Body:           x.Body,
		CreatedAt:      x.CreatedAt,
		UpdatedAt:      x.UpdatedAt,
		FavoritesCount: x.FavoritesCount,
		TagList:        tag,
		Author: &biz.Profile{
			Username: x.Username,
			Email:    x.Email,
			Image:    x.Image,
		},
	}
}

func NewArticleRepo(data *Data, logger log.Logger) biz.ArticleRepo {
	return &articleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *articleRepo) List(ctx context.Context, los biz.ListOptions, opts ...biz.DbOption) (rv []*biz.Article, count int64, err error) {
	var articles []Article
	user := &User{}
	tag := &Tag{}
	if len(los.Favorited) > 0 {
		err := r.data.db.Model(&User{}).Where("username", los.Favorited).First(&user).Error
		if err != nil {
			return nil, 0, err
		}
	}
	if len(los.Tag) > 0 {
		err := r.data.db.Model(&Tag{}).Where("name", los.Tag).First(&tag).Error
		if err != nil {
			return nil, 0, err
		}
	}

	db := r.data.db
	db2 := r.data.db
	for _, opt := range opts {
		db = db.Scopes(opt)
	}
	db = db.Order("id desc")
	if user.ID > 0 {
		db = db.Where("id in (?)", db2.Table("article_favorites").Where("deleted_at is null").Where("user_id", user.ID).Select("article_id"))
	}
	if tag.ID > 0 {
		db = db.Where("id in (?)", db2.Table("article_tags").Where("deleted_at is null").Where("tag_id", tag.ID).Select("article_id"))
	}

	result := db.Model(Article{}).Debug().Find(&articles)

	db.Count(&count)
	if result.Error != nil {
		return nil, count, result.Error
	}
	rv = make([]*biz.Article, len(articles))
	for i, x := range articles {
		rv[i] = convertArticle(x)
	}
	return rv, count, nil
}
func (r *articleRepo) CheckFavorited(uid uint, id uint) bool {
	if err := r.data.db.Where(&ArticleFavorite{UserID: uid, ArticleID: id}).First(&ArticleFavorite{}).Error; err == nil {
		return true
	}
	return false
}

func (r *articleRepo) Get(ctx context.Context, slug string) (rv *biz.Article, err error) {
	x := Article{}
	err = r.data.db.Where("id = ?", slug).Preload("Tags").First(&x).Error
	if err != nil {
		return nil, err
	}
	fmt.Println("asdasdadad")
	fmt.Printf("%+#v", x)
	var fc int64
	rv = convertArticle(x)
	err = r.data.db.Model(&ArticleFavorite{}).Where("article_id = ?", x.ID).Count(&fc).Error
	rv.FavoritesCount = uint32(fc)
	return rv, nil
}

func (r *articleRepo) Create(ctx context.Context, a *biz.Article) (*biz.Article, error) {
	tags := make([]Tag, 0)
	for _, x := range a.TagList {
		tags = append(tags, Tag{
			Name: x,
		})
	}
	if len(tags) > 0 {
		err := r.data.db.Clauses(clause.OnConflict{DoNothing: true}).Create(tags).Error
		if err != nil {
			return nil, err
		}
	}

	po := Article{
		Slug:        a.Slug,
		Title:       a.Title,
		Description: a.Description,
		Body:        a.Body,
		AuthorID:    a.AuthorUserID,
		Tags:        tags,
		Username:    a.Username,
		Email:       a.Email,
		Image:       a.Image,
	}
	result := r.data.db.Create(&po)
	if result.Error != nil {
		return nil, result.Error
	}

	return convertArticle(po), nil
}

func (r *articleRepo) Update(ctx context.Context, a *biz.Article) (*biz.Article, error) {
	var po Article
	if result := r.data.db.Where("id = ?", a.Slug).First(&po); result.Error != nil {
		return nil, result.Error
	}
	tags := make([]Tag, 0)
	for _, x := range a.TagList {
		tags = append(tags, Tag{
			Name: x,
		})
	}
	if len(tags) > 0 {
		err := r.data.db.Clauses(clause.OnConflict{DoNothing: true}).Create(tags).Error
		if err != nil {
			return nil, err
		}
	}
	po.Tags = tags
	po.Title = a.Title
	po.Description = a.Description
	po.Body = a.Body

	// 删除全部旧的tag
	r.data.db.Table("article_tags").Where("article_id", a.Slug).Delete(&struct{}{})
	err := r.data.db.Where("id = ?", a.Slug).Session(&gorm.Session{FullSaveAssociations: true}).Updates(&po).Error
	return convertArticle(po), err
}

func (r *articleRepo) Delete(ctx context.Context, a *biz.Article) error {
	rv := r.data.db.Delete(&Article{}, a.ID)
	return rv.Error
}

func (r *articleRepo) Favorite(ctx context.Context, currentUserID uint, aid uint) error {
	af := ArticleFavorite{
		UserID:    currentUserID,
		ArticleID: aid,
	}

	var a Article
	if err := r.data.db.Where("id = ?", aid).First(&a).Error; err != nil {
		return err
	}

	if result := r.data.db.Where(&ArticleFavorite{UserID: currentUserID, ArticleID: aid}).First(&ArticleFavorite{}); result.RowsAffected == 0 {
		err := r.data.db.Create(&af).Error
		if err != nil {
			return err
		}
		a.FavoritesCount += 1
	} else {
		if err := r.data.db.Where(&ArticleFavorite{UserID: currentUserID, ArticleID: aid}).Delete(&ArticleFavorite{}).Error; err != nil {
			return err
		}
		a.FavoritesCount -= 1
	}

	err := r.data.db.Model(&a).UpdateColumn("favorites_count", a.FavoritesCount).Error
	return err
}

func (r *articleRepo) Unfavorite(ctx context.Context, currentUserID uint, aid uint) error {
	po := ArticleFavorite{
		UserID:    currentUserID,
		ArticleID: aid,
	}
	err := r.data.db.Where(&ArticleFavorite{UserID: currentUserID, ArticleID: aid}).Delete(&po).Error
	if err != nil {
		return err
	}
	var a Article
	if err := r.data.db.Where("id = ?", aid).First(&a).Error; err != nil {
		return err
	}

	err = r.data.db.Model(&a).UpdateColumn("favorites_count", a.FavoritesCount-1).Error
	return err
}

func (r *articleRepo) GetFavoritesStatus(ctx context.Context, currentUserID uint, aa []*biz.Article) (favorited []bool, err error) {
	var po ArticleFavorite
	if result := r.data.db.First(&po); result.Error != nil {
		return nil, nil
	}
	return nil, nil
}

func (r *articleRepo) ListTags(ctx context.Context) (rv []biz.Tag, err error) {
	var tags []Tag
	err = r.data.db.Find(&tags).Error
	if err != nil {
		return nil, err
	}
	rv = make([]biz.Tag, len(tags))
	for i, x := range tags {
		rv[i] = biz.Tag(x.Name)
	}
	return rv, nil
}

func (r *articleRepo) GetArticle(ctx context.Context, aid uint) (rv *biz.Article, err error) {
	x := Article{}
	err = r.data.db.Where("id = ?", aid).First(&x).Error
	if err != nil {
		return nil, err
	}
	var fc int64
	rv = convertArticle(x)
	err = r.data.db.Model(&ArticleFavorite{}).Where("article_id = ?", x.ID).Count(&fc).Error
	rv.FavoritesCount = uint32(fc)
	return rv, nil
}
