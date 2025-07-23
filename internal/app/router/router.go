package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"rk-api/internal/app/api"
	"rk-api/internal/app/router/route"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
)

type IRouter interface {
	Register(app *gin.Engine) error
	Prefixes() []string
}

type Router struct {
	ActivityAPI     *api.ActivityAPI
	AgentAPI        *api.AgentAPI
	FlowAPI         *api.FlowAPI
	NineAPI         *api.NineAPI
	RechargeAPI     *api.RechargeAPI
	UserAPI         *api.UserAPI
	QuizAPI         *api.QuizAPI
	WingoAPI        *api.WingoAPI
	WithdrawAPI     *api.WithdrawAPI
	R8API           *api.R8API
	ZfAPI           *api.ZfAPI
	AdminAPI        *api.AdminAPI
	GameAPI         *api.GameAPI
	OauthAPI        *api.OauthAPI
	AuthAPI         *api.AuthAPI
	VerifyAPI       *api.VerifyAPI
	ChainAPI        *api.ChainAPI
	JhszAPI         *api.JhszAPI
	WalletAPI       *api.WalletAPI
	NotificationAPI *api.NotificationAPI
	ChatAPI         *api.ChatAPI
	PlatAPI         *api.PlatAPI
	RealAPI         *api.RealAPI
	TransactionAPI  *api.TransactionAPI
	StatsAPI        *api.StatsAPI
	HashGameAPI     *api.HashGameAPI
	SDGameAPI       *api.SDGameAPI
	CrashGameAPI    *api.CrashGameAPI
	MineGameAPI     *api.MineGameAPI
	DiceGameAPI     *api.DiceGameAPI
	LimboGameAPI    *api.LimboGameAPI
}

func (a *Router) Register(app *gin.Engine) error {
	a.RegisterAPI(app)
	return nil
}

func (a *Router) Prefixes() []string {
	return []string{"/api/"}
}

func (a *Router) RegisterAPI(app *gin.Engine) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		if err := utils.RegisterTranslations(v); err != nil {
			logger.ZInfo("RegisterTranslations", zap.Error(err))
		}
	}

	r := app.Group("/api")

	// 分组注册
	route.RegisterAuthRoutes(r, a.AuthAPI)
	route.RegisterVerifyRoutes(r, a.VerifyAPI)
	route.RegisterUserRoutes(r, a.UserAPI)
	route.RegisterWalletRoutes(r, a.WalletAPI)
	route.RegisterChainRoutes(r, a.ChainAPI)
	route.RegisterQuizRoutes(r, a.QuizAPI)

	route.RegisterCrashGameRoutes(r, a.CrashGameAPI)
	route.RegisterMineGameRoutes(r, a.MineGameAPI)
	route.RegisterLimboGameRoutes(r, a.LimboGameAPI)
	route.RegisterDiceGameRoutes(r, a.DiceGameAPI)
	route.RegisterSDGameRoutes(r, a.SDGameAPI)
	route.RegisterHashGameRoutes(r, a.HashGameAPI)

	route.RegisterActivityRoutes(r, a.ActivityAPI)
	route.RegisterAgentRoutes(r, a.AgentAPI)
	route.RegisterAdminRoutes(r, a.AdminAPI)
	route.RegisterWithdrawRoutes(r, a.WithdrawAPI)
	route.RegisterRechargeRoutes(r, a.RechargeAPI)
	route.RegisterStatsRoutes(r, a.StatsAPI)

	route.RegisterR8GameRoutes(r, a.R8API)
	route.RegisterZFGameRoutes(r, a.ZfAPI)
	route.RegisterJHSZGameRoutes(r, a.JhszAPI)
	route.RegisterGameRoutes(r, a.GameAPI)
	route.RegisterChatRoutes(r, a.ChatAPI)
	route.RegisterPlatRoutes(r, a.PlatAPI)
	route.RegisterRealRoutes(r, a.RealAPI)
	route.RegisterTransactionRoutes(r, a.TransactionAPI)
	route.RegisterFlowRoutes(r, a.FlowAPI)

	route.RegisterNineRoutes(r, a.NineAPI)
	route.RegisterWingoRoutes(r, a.WingoAPI)
	route.RegisterOAuthRoutes(r, a.OauthAPI, app)
}
