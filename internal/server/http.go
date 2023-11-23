package server

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	articleServer "realworld/api/article/v1"
	profileService "realworld/api/profile/v1"
	userServer "realworld/api/user/v1"
	"realworld/internal/conf"
	"realworld/internal/service"
	"realworld/pkg/middleware/auth"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

func NewSkipRoutersMatcher() selector.MatchFunc {

	skipRouters := map[string]struct{}{
		"/user.v1.User/Authentication":     {},
		"/user.v1.User/Registration":       {},
		"/article.v1.Article/GetArticle":   {},
		"/article.v1.Article/ListArticles": {},
		"/article.v1.Article/GetComments":  {},
		"/article.v1.Article/GetTags":      {},
		"/profile.v1.Profile/GetProfile":   {},
	}

	return func(ctx context.Context, operation string) bool {
		if _, ok := skipRouters[operation]; ok {
			if tr, ok := transport.FromServerContext(ctx); ok {
				fmt.Println("JWTJWTJWT")
				if len(tr.RequestHeader().Get("Authorization")) > 0 {
					return true
				}
			}
			return false
		}
		return true
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, user *service.UserService, profile *service.ProfileService, article *service.ArticleService, jwtc *conf.JWT, logger log.Logger) *http.Server {
	var (
		opts = []http.ServerOption{
			http.ErrorEncoder(errorEncoder),
			http.Middleware(
				recovery.Recovery(),
				selector.Server(auth.JWTAuth(jwtc.Secret)).Match(NewSkipRoutersMatcher()).Build(),
				logging.Server(logger),
			),

			http.Filter(
				handlers.CORS(
					handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
					handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"}),
					handlers.AllowedOrigins([]string{"*"}),
				),
			),
		}
	)
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	userServer.RegisterUserHTTPServer(srv, user)
	profileService.RegisterProfileHTTPServer(srv, profile)
	articleServer.RegisterArticleHTTPServer(srv, article)

	return srv
}
