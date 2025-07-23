package service

import (
	"net/url"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"

	"encoding/base32"

	"github.com/skip2/go-qrcode"

	"github.com/dgryski/dgoogauth"
)

var AdminServiceSet = wire.NewSet(
	ProvideAdminService,
)

type AdminService struct {
	Repo        *repository.AdminRepository
	StateSrv    *StateService
	SecretCache *ecache.Cache
}

func ProvideAdminService(repo *repository.AdminRepository, stateSrv *StateService) *AdminService {
	secretCache := ecache.NewLRUCache(1, 4, 30*time.Minute) //初始化缓存
	return &AdminService{
		Repo:        repo,
		SecretCache: secretCache,
		StateSrv:    stateSrv,
	}
}

// 创建系统操作日志
func (s *AdminService) GetSysUserAreaList() ([]*entities.SysUserArea, error) {
	return s.Repo.GetSysUserAreaList()
}

// 创建系统操作日志
func (s *AdminService) CreateSystemOptionLog(entity *entities.SystemOptionLog) error {
	entity.Time = time.Now().Unix()
	return s.Repo.CreateSystemOptionLog(entity)
}

func (s *AdminService) GetSysUser(entity *entities.SysUser) (*entities.SysUser, error) {
	return s.Repo.GetSysUser(entity)
}

func (s *AdminService) GetSysUserAdmin(pc uint) (*entities.SysUserAdmin, error) {
	return s.Repo.GetSysUserAdmin(pc)
}

func (s *AdminService) CheckGoogleAuthCodeBinded(account string) (bool, error) {
	user, err := s.Repo.GetSysUserByUsername(account)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.With("account not exist")
	}
	if user.Secret != "" {
		return true, nil
	}
	return false, nil
}

func (s *AdminService) GenAuthQRCode(account string) (string, error) {
	user, err := s.Repo.GetSysUserByUsername(account)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.With("account not exist")
	}
	// 生成一个随机秘密密钥
	secret_ := base32.StdEncoding.EncodeToString([]byte(config.Get().ServiceSettings.JwtSignKey))

	secret := strings.TrimRight(secret_, "=") //兼容ios

	issuer := "cheetahs" // 这里是你的应用名称

	// 生成OTPAuth链接
	URL, err := url.Parse("otpauth://totp/")
	if err != nil {
		return "", err
	}

	URL.Path += url.PathEscape(issuer) + ":" + url.PathEscape(account)
	params := url.Values{}
	params.Add("secret", secret)
	params.Add("issuer", issuer)

	URL.RawQuery = params.Encode()

	// 生成二维码
	png, err := qrcode.Encode(URL.String(), qrcode.Medium, 256)
	if err != nil {
		return "", errors.With("can not make qrcode")
	}

	fileName := account + ".png" // 可以基于用户账号或者其他唯一标识来命名文件

	err = utils.WriteFileWithDir(config.Get().ServiceSettings.StaticPath+"qrcode/", fileName, png, 0644) // 0644 是文件的权限
	if err != nil {
		return "", err
	}
	// logger.Error("--------------", s.SecretCache, account, secret)

	s.SecretCache.Put(account, secret_) //放入缓存

	// 记录保存文件的信息，实用的日志记录操作总是好的
	logger.Info("Saved QR code to ", fileName)
	return fileName, nil
}

// // 用户从Google Authenticator应用中获得的6位数字
func (s *AdminService) VerifyGoogleAuthCode(account string, authCode string) error {
	user, err := s.Repo.GetSysUserByUsername(account)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.With("account not exist")
	}

	secret := user.Secret

	if secret == "" { //假设还没有设置，就看看缓存有没有
		val, ok := s.SecretCache.Get(account)
		if !ok {
			return errors.With("bind exipre")
		}
		secret = val.(string)
	}

	otpConfig := &dgoogauth.OTPConfig{
		Secret:      secret,
		WindowSize:  3, // 允许的时间偏差，具体值要看应用需要
		HotpCounter: 0, // 只用于HOTP，对于TOTP这里应始终为0
	}
	// NJ2WY2LFOJAGYYLOMRUW4ZZSGAZDG===

	// logger.ZError("VerifyGoogleAuthCode", zap.String("account", account), zap.String("authCode", authCode), zap.String("secret", secret))
	// 验证输入的验证码是否正确
	valid, err := otpConfig.Authenticate(authCode)
	if err != nil {
		logger.ZError("VerifyGoogleAuthCode", zap.String("account", account), zap.String("authCode", authCode), zap.String("secret", secret), zap.Error(err))
		return err
	}

	if !valid {
		return errors.With("invalid auth code")
		// 用户验证码错误，显示错误或重新验证
	}
	if user.Secret == "" { //没有设置则设置
		user.Secret = secret
		if err := s.Repo.UpdateSysUser(user); err != nil {
			return err
		}
	}
	return nil
}

// 月度备份和清理
func (s *AdminService) CallMonthBackupAndClean(req *entities.MonthBackupAndCleaReq) error {
	logger.ZInfo("CallMonthBackupAndClean", zap.Any("req", req))
	go func() {
		defer utils.PrintPanicStack()
		s.StateSrv.SetState(constant.StateMonthBackupAndClean, true)
		// 确保在函数结束时重新启用返利
		defer func() {
			s.StateSrv.SetState(constant.StateMonthBackupAndClean, false)
		}()
		if err := s.Repo.CallMonthBackupAndCleanV4(req.TableNames); err != nil {
			logger.ZError("CallMonthBackupAndClean", zap.Error(err))
		}
	}()
	return nil
}

// 变更PC号业务员合并
func (s *AdminService) CallChangePC(req *entities.CallChangePCReq) error {
	logger.ZInfo("CallChangePC", zap.Any("req", req))
	go func() {
		defer utils.PrintPanicStack()
		s.StateSrv.SetState(constant.StateChangePC, true)
		// 确保在函数结束时重新启用返利
		defer func() {
			s.StateSrv.SetState(constant.StateChangePC, false)
		}()

		if err := s.Repo.CallChangePC(req.SRC, req.DST); err != nil {
			logger.ZError("CallChangePC", zap.Error(err))
		}
	}()
	return nil
}
