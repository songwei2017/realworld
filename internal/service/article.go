package service

import (
	"context"
	"fmt"
	"realworld/pkg/middleware/auth"
	"strconv"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "realworld/api/article/v1"
	"realworld/internal/biz"
)

type ArticleService struct {
	pb.UnimplementedArticleServer
	uc *biz.SocialUsecase
}

func NewArticleService(uc *biz.SocialUsecase) *ArticleService {
	return &ArticleService{uc: uc}
}

func convertArticle(do *biz.Article) *pb.Articles {
	return &pb.Articles{
		Slug:           strconv.Itoa(int(do.ID)),
		Title:          do.Title,
		Description:    do.Description,
		Body:           do.Body,
		TagList:        do.TagList,
		CreatedAt:      timestamppb.New(do.CreatedAt),
		UpdatedAt:      timestamppb.New(do.UpdatedAt),
		Favorited:      do.Favorited,
		FavoritesCount: do.FavoritesCount,
		Author:         convertProfile(do.Author),
	}
}

func convertComment(do *biz.Comment) *pb.Comment {
	return &pb.Comment{
		Id:        uint32(do.ID),
		CreatedAt: timestamppb.New(do.CreatedAt),
		UpdatedAt: timestamppb.New(do.UpdatedAt),
		Body:      do.Body,
		Author:    convertProfile(do.Author),
	}
}

func convertProfile(do *biz.Profile) *pb.Profile {
	return &pb.Profile{
		Username:  do.Username,
		Bio:       do.Bio,
		Image:     do.Image,
		Following: do.Following,
	}
}

func (s *ArticleService) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (reply *pb.SingleArticleReply, err error) {
	rv, err := s.uc.GetArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &pb.SingleArticleReply{
		Article: convertArticle(rv),
	}, nil
}

func (s *ArticleService) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (reply *pb.SingleArticleReply, err error) {
	rv, err := s.uc.CreateArticle(ctx, &biz.Article{
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		TagList:     req.Article.TagList,
	})
	if err != nil {
		return nil, err
	}
	return &pb.SingleArticleReply{
		Article: convertArticle(rv),
	}, nil
}

func (s *ArticleService) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (reply *pb.SingleArticleReply, err error) {
	rv, err := s.uc.UpdateArticle(ctx, &biz.Article{
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		TagList:     req.Article.TagList,
		Slug:        req.Slug,
	})
	if err != nil {
		return nil, err
	}
	return &pb.SingleArticleReply{
		Article: convertArticle(rv),
	}, nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (reply *pb.SingleArticleReply, err error) {
	err = s.uc.DeleteArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &pb.SingleArticleReply{
		Article: &pb.Articles{
			Slug: req.Slug,
		},
	}, nil
}

func (s *ArticleService) AddComment(ctx context.Context, req *pb.AddCommentRequest) (reply *pb.SingleCommentReply, err error) {
	rv, err := s.uc.AddComment(ctx, req.Slug, &biz.Comment{
		Body: req.Comment.Body,
	})
	if err != nil {
		return nil, err
	}
	return &pb.SingleCommentReply{
		Comment: convertComment(rv),
	}, nil
}

func (s *ArticleService) GetComments(ctx context.Context, req *pb.AddCommentRequest) (reply *pb.MultipleCommentsReply, err error) {
	rv, err := s.uc.ListComments(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	comments := make([]*pb.Comment, 0)
	for _, x := range rv {
		comments = append(comments, convertComment(x))
	}
	return &pb.MultipleCommentsReply{Comments: comments}, nil
}

func (s *ArticleService) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (reply *pb.SingleCommentReply, err error) {
	err = s.uc.DeleteComment(ctx, uint(req.Id))
	return &pb.SingleCommentReply{
		Comment: &pb.Comment{
			Id: uint32(req.Id),
		},
	}, err
}

func (s *ArticleService) FeedArticles(ctx context.Context, req *pb.FeedArticlesRequest) (reply *pb.MultipleArticlesReply, err error) {
	uid := auth.GetUserIdOrNotLogin(ctx)
	rv, count, err := s.uc.ListArticles(ctx,
		biz.ListOptions{},
		biz.DbLimit(req.Limit),
		biz.DbOffset(req.Offset),
		biz.DbAuthorId(uid),
	)
	if err != nil {
		return nil, err
	}
	articles := make([]*pb.Articles, 0)
	for _, x := range rv {
		articles = append(articles, convertArticle(x))
	}
	return &pb.MultipleArticlesReply{Articles: articles, ArticlesCount: uint64(count)}, nil
}

func (s *ArticleService) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (reply *pb.MultipleArticlesReply, err error) {
	fmt.Println("ListArticles", req)
	los := biz.ListOptions{
		Favorited: req.Favorited,
		Tag:       req.Tag,
	}
	rv, count, err := s.uc.ListArticles(ctx,
		los,
		biz.DbLimit(req.Limit),
		biz.DbOffset(req.Offset),
		biz.DbAuthor(req.Author),
	)
	if err != nil {
		return nil, err
	}
	articles := make([]*pb.Articles, 0)
	for _, x := range rv {
		articles = append(articles, convertArticle(x))
	}
	return &pb.MultipleArticlesReply{Articles: articles, ArticlesCount: uint64(count)}, nil
}

func (s *ArticleService) GetTags(ctx context.Context, req *pb.GetTagsRequest) (reply *pb.TagListReply, err error) {
	rv, err := s.uc.GetTags(ctx)
	if err != nil {
		return nil, err
	}
	tags := make([]string, len(rv))
	for i, x := range rv {
		tags[i] = string(x)
	}
	reply = &pb.TagListReply{Tags: tags}
	return reply, nil
}

func (s *ArticleService) FavoriteArticle(ctx context.Context, req *pb.FavoriteArticleRequest) (reply *pb.SingleArticleReply, err error) {
	rv, err := s.uc.FavoriteArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &pb.SingleArticleReply{
		Article: convertArticle(rv),
	}, nil
}

func (s *ArticleService) UnfavoriteArticle(ctx context.Context, req *pb.UnfavoriteArticleRequest) (reply *pb.SingleArticleReply, err error) {
	rv, err := s.uc.UnfavoriteArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &pb.SingleArticleReply{
		Article: convertArticle(rv),
	}, nil
}
