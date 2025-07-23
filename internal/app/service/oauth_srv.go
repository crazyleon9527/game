package service

import (
	"context"
	"encoding/json"
	"fmt"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/oauth"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"time"

	"github.com/google/go-github/github"
	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	resty "rk-api/pkg/http"
	"rk-api/pkg/logger"
)

var OauthServiceSet = wire.NewSet(
	ProvideOauthService,
)

type OauthService struct {
	Repo       *repository.OauthRepository
	UserSrv    *UserService
	StateCache *ecache.Cache
}

func ProvideOauthService(
	repo *repository.OauthRepository,
	userSrv *UserService,
) *OauthService {
	service := &OauthService{
		Repo:    repo,
		UserSrv: userSrv,
	}
	oauth.InitConfig()                                            //初始化oauth配置
	service.StateCache = ecache.NewLRUCache(1, 6, 10*time.Minute) //初始化缓存
	return service
}

func (s *OauthService) OauthStateUrl(oauthState *entities.OauthState) error {

	oauthState.State = utils.GetRandomSalt()

	switch oauthState.Plat {
	case "github":
		oauthState.OauthUrl = oauth.GithubOauthConfig.AuthCodeURL(oauthState.State, oauth2.AccessTypeOnline)
		logger.ZInfo("OauthStateUrl2", zap.String("state", oauthState.State), zap.String("url", oauthState.OauthUrl), zap.String("plat", oauthState.Plat))
	case "facebook":
		oauthState.OauthUrl = oauth.FacebookOauthConfig.AuthCodeURL(oauthState.State, oauth2.AccessTypeOnline)
	case "twister":
		oauthState.OauthUrl = oauth.TwitterOauthConfig.AuthCodeURL(oauthState.State, oauth2.AccessTypeOnline)
	case "google":
		oauthState.OauthUrl = oauth.GoogleOauthConfig.AuthCodeURL(oauthState.State, oauth2.AccessTypeOnline)
	default:
		return errors.WithCode(errors.UnsupportThirdLogin)
	}

	s.StateCache.Put(oauthState.State, oauthState)

	return nil
}

func (s *OauthService) GetOauthState(state string) (*entities.OauthState, error) {

	obj, exisit := s.StateCache.Get(state)

	if !exisit {
		return nil, errors.With("not exist state or state is expire")
	}
	return obj.(*entities.OauthState), nil
}

func (s *OauthService) VerifyGoogleUser(code string) (*oauth.GoogleUser, error) {

	oauthToken, err := oauth.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		msg := fmt.Sprintf("oauthConf.Exchange() failed with '%s'\n", err)
		return nil, errors.With(msg)
	}

	client := resty.GetHttpClient()

	resp, err := client.R().
		SetHeader("Accept", "application/json").
		Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + oauthToken.AccessToken)

	if err != nil {
		msg := fmt.Sprintf("failed getting user info: %s", err)
		return nil, errors.With(msg)
	}

	// 解析响应体
	gu := oauth.GoogleUser{}
	if err := json.Unmarshal(resp.Body(), &gu); err != nil {
		return nil, err
	}

	return &gu, nil
}

func (s *OauthService) VerifyGithubUser(code string) (*github.User, error) {

	oauthToken, err := oauth.GithubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		msg := fmt.Sprintf("oauthConf.Exchange() failed with '%s'\n", err)
		return nil, errors.With(msg)
	}

	oauthClient := oauth.GithubOauthConfig.Client(context.Background(), oauthToken)

	client := github.NewClient(oauthClient)

	oauthUser, _, err := client.Users.Get(context.Background(), "")

	if err != nil {
		msg := fmt.Sprintf("client.Users.Get() faled with '%s'\n", err)
		return nil, errors.With(msg)
	}
	return oauthUser, nil
}
