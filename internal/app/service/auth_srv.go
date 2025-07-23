package service

import (
	"fmt"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Transform to Uppercase

var AuthServiceSet = wire.NewSet(
	ProvideAuthService,
)

type AuthService struct {
	Repo      *repository.UserRepository
	UserSrv   *UserService
	VerifySrv *VerifyService
	AdminSrv  *AdminService
	StateSrv  *StateService
}

func ProvideAuthService(repo *repository.UserRepository,
	adminSrv *AdminService,
	userSrv *UserService,
	verifySrv *VerifyService,
	stateSrv *StateService,
) *AuthService {

	service := &AuthService{
		Repo:      repo,
		UserSrv:   userSrv,
		VerifySrv: verifySrv,
		AdminSrv:  adminSrv,
		StateSrv:  stateSrv,
	}
	return service
}

// 转换分销码

func (s *AuthService) RegisterUser(registerCredentials *entities.RegisterCredentials) (*entities.User, error) {
	if registerCredentials.Username == "" {
		if registerCredentials.Mobile != "" { //有手机的时候填手机
			registerCredentials.Username = registerCredentials.Mobile
		}
	}
	if registerCredentials.Username == "" {
		if registerCredentials.Email != "" { //有邮箱的时候填邮箱
			registerCredentials.Username = registerCredentials.Email
		}
	}
	if registerCredentials.Username == "" {
		return nil, errors.WithCode(errors.InvalidParam)
	}

	if !registerCredentials.IsOAuth && registerCredentials.Password == "" {
		return nil, errors.WithCode(errors.InvalidParam)
	}

	PromoterCode, err := registerCredentials.ConvertPC() //转换分销码
	if err != nil {
		return nil, err
	}

	if !s.StateSrv.GetBoolState(constant.StateSMSVerificationDisabled) { //开启短信验证
		if err = s.VerifySrv.CheckVerifyCode(registerCredentials.Mobile, registerCredentials.VerCode); err != nil { //判断验证码 todo redis
			return nil, err
		}
	}

	var invite *entities.User
	var inviter string                        //邀请者手机
	if registerCredentials.InviteCode != "" { //特殊
		if len(registerCredentials.InviteCode) < 6 || len(registerCredentials.InviteCode) > 8 {
			return nil, errors.WithCode(errors.InvalidInviteCode)
		}
		invite, err = s.Repo.GetUserByInviteCode(registerCredentials.InviteCode) //邀请人
		if err != nil {
			return nil, err
		}
		if invite == nil {
			return nil, errors.WithCode(errors.InvalidInviteCode)
		}

		PromoterCode = invite.PromoterCode // promotion code 赋予
		inviter = invite.Mobile            //邀请者手机号
	}

	// if PromoterCode == 0 || PromoterCode == 1 { //判断是否是分销商
	// 	return nil, errors.WithCode(errors.InvalidPromotionCode) // todo 解析
	// }

	// 判断用户是否存在
	user, err := s.Repo.GetUserByUsername(registerCredentials.Username)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, errors.WithCode(errors.UserNameExist)
	}

	var hashedPassword []byte
	if registerCredentials.Password != "" { //判断手机号是否存在
		// 创建新用户
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(registerCredentials.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
	}
	user = &entities.User{
		Username:     registerCredentials.Username,
		Password:     string(hashedPassword),
		Plat:         registerCredentials.Plat,
		Channel:      registerCredentials.Channel,
		Mobile:       registerCredentials.Mobile,
		Email:        registerCredentials.Email,
		Inviter:      inviter,
		PromoterCode: PromoterCode,
		Nickname:     registerCredentials.Username,
		IP:           registerCredentials.LoginIP,
		IsRobot:      registerCredentials.IsRobot,
		Status:       registerCredentials.Status,
		Telegram:     registerCredentials.Telegram,
		Device:       registerCredentials.Device,
	}
	if user.Status == 0 {
		user.Status = constant.USER_STATE_NORMAL
	}

	sysUser, err := s.AdminSrv.GetSysUser(&entities.SysUser{UID: PromoterCode}) //用户的分销人处理
	if err != nil {
		return nil, errors.WithCode(errors.InvalidPromotionCode) //
	}
	user.Promoter = sysUser.Username //

	// 将用户保存到数据库
	if err := s.UserSrv.CreateUser(user); err != nil {
		return nil, err
	}

	logger.ZInfo("register succ",
		zap.Uint("id", user.ID),
		zap.String("mobile", user.Mobile),
		zap.String("ip", user.LoginIP),
		zap.String("invite_code", registerCredentials.InviteCode),
		zap.String("promotor", user.Promoter),
	)

	if invite != nil { //处理邀请关系
		inviteForUpdate := new(entities.User)
		inviteForUpdate.ID = invite.ID
		inviteForUpdate.InviteCount = invite.InviteCount + 1 //邀请的一级增加一个

		if err := s.UserSrv.UpdateUser(inviteForUpdate); err != nil {
			return nil, err
		}

		relation := &entities.HallInviteRelation{
			UID:    user.ID, //注册完成后才能拿到
			Mobile: user.Mobile,
			PID:    invite.ID,
			Level:  1, //1级
		}

		inviteRelationQueue, _ := handle.NewInviteRelationQueue(relation) //消息队列去处理 代理关系
		if _, err := mq.MClient.Enqueue(inviteRelationQueue); err != nil {
			logger.ZError("inviteRelationQueue", zap.Any("relation", relation), zap.Error(err))
		}

		// pinduo := &entities.HallInvitePinduo{
		// 	InviteID:   invite.ID,
		// 	InviteName: invite.Username,
		// 	UID:        user.ID,
		// }

		// invitePinduoQueue, _ := handle.NewInvitePinduoQueue(pinduo) //消息队列去处理 代理关系
		// if _, err := mq.MClient.Enqueue(invitePinduoQueue); err != nil {
		// 	logger.ZError("invitePinduoQueue", zap.Any("pinduo", pinduo), zap.Error(err))
		// }
	}

	return user, nil

}

func (s *AuthService) VerifyCredentials(credentials *entities.VerifyCredentials) error {

	if credentials.Username == "" || credentials.Password == "" {
		return errors.WithCode(errors.InvalidParam)
	}

	user, err := s.Repo.GetUserByUsername(credentials.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.WithCode(errors.AccountNotExist)
	}
	if user.Status == constant.USER_STATE_BLOCKED {
		return errors.WithCode(errors.AccountBlocked)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		return errors.WithCode(errors.InvalidPassword)
	}

	return nil
}

func (s *AuthService) MobileLogin(loginCredentials *entities.MobileLoginCredentials) (*entities.User, error) {
	user, err := s.Repo.GetUserByMobile(loginCredentials.Mobile)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.WithCode(errors.AccountNotExist)
	}
	if user.Status == constant.USER_STATE_BLOCKED {
		return nil, errors.WithCode(errors.AccountBlocked)
	}

	if s.StateSrv.GetBoolState(constant.StateSMSVerificationDisabled) { //开启短信验证
		return nil, errors.WithCode(errors.SMSVerificationDisabled)
	}

	if err = s.VerifySrv.CheckVerifyCode(loginCredentials.Mobile, loginCredentials.VerCode); err != nil { //判断验证码 todo redis
		return nil, err
	}

	userForUpdate := &entities.User{
		LoginTime: time.Now().Unix(),
		LoginIP:   loginCredentials.LoginIP,
	}
	userForUpdate.ID = user.ID

	if err := s.Repo.UpdateUser(userForUpdate); err != nil {
		return nil, err
	}
	// redis
	s.UserSrv.UpdateUserTTL(user.ID) //用户信息设置新的过期时间
	logger.ZInfo("login",
		zap.Uint("id", user.ID),
		zap.String("mobile", user.Mobile),
	)
	return user, nil
}

func (s *AuthService) Login(loginCredentials *entities.LoginCredentials) (*entities.User, error) {
	user, err := s.Repo.GetUserByUsername(loginCredentials.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.WithCode(errors.AccountNotExist)
	}
	if user.Status == constant.USER_STATE_BLOCKED {
		return nil, errors.WithCode(errors.AccountBlocked)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginCredentials.Password)); err != nil {
		return nil, errors.WithCode(errors.InvalidPassword)
	}

	userForUpdate := &entities.User{
		LoginTime: time.Now().Unix(),
		LoginIP:   loginCredentials.LoginIP,
	}
	userForUpdate.ID = user.ID

	if err := s.Repo.UpdateUser(userForUpdate); err != nil {
		return nil, err
	}
	// redis
	s.UserSrv.UpdateUserTTL(user.ID) //用户信息设置新的过期时间
	logger.ZInfo("login",
		zap.Uint("id", user.ID),
		zap.String("mobile", user.Mobile),
	)
	return user, nil
}

func (s *AuthService) VerifyLoginToken(uid uint, token string) error {
	keepToken, _ := s.Repo.GetUserAccessToken(uid)
	if keepToken != token {
		return errors.WithCode(errors.AccountLoginExpire)
	}
	return nil
}

func (s *AuthService) ChangePassword(changePasswordCredentials *entities.ChangePasswordCredentials) error {

	loginCredentials := entities.LoginCredentials{Username: changePasswordCredentials.Username, Password: changePasswordCredentials.Password}

	user, err := s.Login(&loginCredentials)
	if err != nil {
		return err
	}
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(changePasswordCredentials.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if string(hashedNewPassword) == user.Password {
		return errors.WithCode(errors.DuplicatePassword)
	}

	// user.Password = string(hashedNewPassword)
	return s.Repo.UpdatePassword(user, string(hashedNewPassword))
}

func (s *AuthService) ResetPassword(resetPasswordCredentials *entities.ResetPasswordCredentials) error {

	//判断验证码 todo redis
	verifyCode := new(entities.VerifyCode)
	verifyCode.Code = resetPasswordCredentials.VerCode
	verifyCode.Target = resetPasswordCredentials.Mobile
	verifyCode, err := s.Repo.GetVerifyCode(verifyCode)
	if err != nil {
		return errors.WithCode(errors.VerifiedCodeNotMatch) // todo 解析
	}

	if time.Now().Unix()-verifyCode.UpdatedAt > constant.MAX_TOKEN_EXIPRY_TIME {
		return errors.WithCode(errors.VerifiedCodeExpire)
	}

	user, err := s.Repo.GetUserByUsername(resetPasswordCredentials.Mobile) //
	if err != nil {
		return errors.WithCode(errors.AccountNotExist)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(resetPasswordCredentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if string(hashedNewPassword) == user.Password {
		return errors.WithCode(errors.DuplicatePassword)
	}

	user.Password = string(hashedNewPassword)
	user.LoginTime = time.Now().Unix()
	user.LoginIP = resetPasswordCredentials.LoginIP

	if err = s.Repo.UpdateUser(user); err != nil {
		return err
	}
	//重置密码无需 修改redis

	logger.ZInfo("reset password succ",
		zap.Uint("id", user.ID),
		zap.String("ip", user.LoginIP),
		zap.String("mobile", user.Mobile),
	)

	return nil
}

func (s *AuthService) CreateJWTAccessRefreshToken(sso *entities.OAuthToken) error {

	accessExpireTime := time.Duration(config.Get().ServiceSettings.TokenExpireTime*24) * time.Hour

	// 过期的时间
	sso.ExpiresIn = accessExpireTime

	accessToken, err := utils.GenerateJWT(config.Get().ServiceSettings.JwtSignKey, accessExpireTime, fmt.Sprintf("%d", sso.UID))
	if err != nil {
		return err
	}

	sso.AccessToken = accessToken

	refreshExpireTime := time.Duration(config.Get().ServiceSettings.TokenRefreshTime*24) * time.Hour

	refreshToken, err := utils.GenerateJWT(config.Get().ServiceSettings.JwtSignKey, refreshExpireTime, fmt.Sprintf("%d", sso.UID))
	if err != nil {
		return err
	}
	sso.RefreshToken = refreshToken

	// refreshExpiration := time.Duration(*conf.JWTSettings.RefreshTime) * time.Hour
	// err = models.CreateAccessRefreshToken(sso.AccessToken, sso.RefreshToken, sso.ExpiresIn, refreshExpiration)

	if err := s.Repo.SetUserAccessToken(sso.UID, sso.AccessToken, sso.ExpiresIn); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Logout(uid uint) error {
	if err := s.Repo.ExpireUserAccessToken(uid); err != nil {
		return err
	}
	return nil
}

func (r *AuthService) ReigsterAuthUser(req *entities.AddUserReq) (err error) {
	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_EDIT_USER_INFO,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(req),
			Remark:   "修改用户信息",
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "添加用户失败"
			// 打印错误，使用你的日志库方法来记录
			logger.ZError("AddUser", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "添加用户成功"
			logger.ZError("AddUser", zap.Any("req", req))
		}
		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()

	// user := new(entities.User)
	credentials := entities.RegisterCredentials{
		Username: req.Username,
		Password: req.Password,
		Mobile:   req.Username,
		VerCode:  config.Get().ServiceSettings.TrustedUserCode,
		LoginIP:  req.IP,
		Status:   req.Status,
		IsRobot:  1,
	}
	credentials.PromotionCode = fmt.Sprintf("%d", req.OptionID)

	if _, err = r.RegisterUser(&credentials); err != nil {
		return err
	}
	return nil
}
