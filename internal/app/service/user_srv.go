package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"rk-api/pkg/storage"
	"rk-api/pkg/structure"
	"time"

	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	AvatarsBucketName = "avatars"
)

// var UserServiceSet = wire.NewSet(wire.Struct(new(UserService), "Repo", "AdminSrv"))
var UserServiceSet = wire.NewSet(
	ProvideUserService,
)

type UserService struct {
	Repo         *repository.UserRepository
	minioCli     *minio.Client
	AdminSrv     *AdminService
	StateSrv     *StateService
	UserLocks    *entities.RedisUserLock
	FinancialSrv *FinancialService
	VerifySrv    *VerifyService
	walletSrv    *WalletService
	// UserLocks  *entities.RedisUserLock

}

func ProvideUserService(repo *repository.UserRepository,
	adminSrv *AdminService,
	stateSrv *StateService,
	walletSrv *WalletService,
	verifySrv *VerifyService,
	financialSrv *FinancialService,
	minioCli *minio.Client,
) *UserService {
	service := &UserService{
		Repo:         repo,
		AdminSrv:     adminSrv,
		StateSrv:     stateSrv,
		FinancialSrv: financialSrv,
		VerifySrv:    verifySrv,
		walletSrv:    walletSrv,
		minioCli:     minioCli,
		UserLocks:    entities.NewRedisUserLock(repo.RDS), //分布式锁
		// UserLocks:  entities.NewRedisUserLock(repo.RDS), //分布式锁
	}
	go storage.CreateBucket(minioCli, AvatarsBucketName)
	return service
}

// func (s *UserService) Lock(uid uint) {
// 	s.UserLocks.Lock(uid)
// }

// func (s *UserService) Unlock(uid uint) {
// 	s.UserLocks.Unlock(uid)
// }

func (s *UserService) CreateUser(user *entities.User) error {
	if err := s.Repo.CreateUser(user); err != nil {
		return err
	}
	wallet := entities.UserWallet{
		UID:          user.ID,
		PromoterCode: user.PromoterCode,
	}
	if err := s.walletSrv.CreateWallet(&wallet); err != nil {
		return err
	}
	summary := entities.FinancialSummary{
		UID:          user.ID,
		PromoterCode: user.PromoterCode,
	}
	if err := s.FinancialSrv.CreateFinancialSummary(&summary); err != nil {
		return err
	}
	return nil
}

// 更新登录信息
func (s *UserService) UpdateLoginInfo(uid uint, ip string) error {
	userForUpdate := entities.User{
		LoginTime: time.Now().Unix(),
		LoginIP:   ip,
	}
	userForUpdate.ID = uid
	if err := s.UpdateUser(&userForUpdate); err != nil {
		return err
	}
	logger.ZInfo("update login ",
		zap.Uint("id", uid),
		zap.String("ip", ip),
	)
	return nil
}

func (s *UserService) OnLogin(uid uint, hooks func()) error {
	ac, _ := s.Repo.GetAccessTime(uid)
	now := time.Now().Unix()

	if now-ac > 30 {
		hooks()
		s.Repo.SetAccessTime(uid, now)
	}
	return nil
}

func (s *UserService) VerifyLoginToken(uid uint, token string) error {
	keepToken, _ := s.Repo.GetUserAccessToken(uid)
	if keepToken != token {
		return errors.WithCode(errors.AccountLoginExpire)
	}
	return nil
}

func (s *UserService) ChangeNickname(req *entities.EditNicknameReq) error {

	user, err := s.GetUserByUID(req.UID)
	if err != nil {
		return err
	}

	if err := s.UpdateUserNameAndGender(req.UID, req.Nickname, req.Gender); err != nil {
		return err
	}

	//重置密码无需 修改redis

	logger.ZInfo("ChangeNickname succ",
		zap.Uint("id", req.UID),
		zap.String("username", user.Username),
		zap.String("nickname", req.Nickname),
		zap.Uint8("gender", req.Gender),
	)

	return nil
}

func (s *UserService) ChangeAvatar(req *entities.EditAvatarReq) error {

	_, err := s.GetUserByUID(req.UID)
	if err != nil {
		return err
	}

	userForUpdate := &entities.User{
		Avatar: req.Avatar,
	}
	userForUpdate.ID = req.UID

	if err := s.UpdateUser(userForUpdate); err != nil {
		return err
	}

	// logger.ZInfo("ChangeAvatar succ",
	// 	zap.Uint("id", userForUpdate.ID),
	// 	zap.String("nickname", user.Nickname),
	// )

	return nil
}

func (s *UserService) BindTelegram(req *entities.BindTelegramReq) error {
	_, err := s.GetUserByUID(req.UID)
	if err != nil {
		return err
	}
	userForUpdate := &entities.User{
		Telegram: req.Telegram,
	}
	userForUpdate.ID = req.UID

	if err := s.UpdateUser(userForUpdate); err != nil {
		return err
	}
	return nil
}

func (s *UserService) BindEmail(req *entities.BindEmailReq) error {
	_, err := s.GetUserByUID(req.UID)
	if err != nil {
		return err
	}

	if err = s.VerifySrv.CheckVerifyCode(req.Email, req.Code); err != nil { //判断验证码 todo redis
		return err
	}

	userForUpdate := &entities.User{
		Telegram: req.Email,
	}
	userForUpdate.ID = req.UID

	if err := s.UpdateUser(userForUpdate); err != nil {
		return err
	}
	return nil
}

func (s *UserService) SearchUser(req *entities.SearchUserReq) (*entities.User, error) {
	user := new(entities.User)
	structure.Copy(req, user)
	return s.Repo.GetUser(user)
}

func (s *UserService) GetUserProfile(uid uint) (*entities.UserProfile, error) {
	// 使用通道传递结果和错误，避免数据竞争
	type userResult struct {
		user *entities.User
		err  error
	}
	userCh := make(chan userResult, 1)
	go func() {
		user, err := s.Repo.GetUserByID(uid)
		userCh <- userResult{user, err}
	}()

	type walletResult struct {
		wallet *entities.UserWallet
		err    error
	}
	walletCh := make(chan walletResult, 1)
	go func() {
		wallet, err := s.walletSrv.GetWallet(uid)
		walletCh <- walletResult{wallet, err}
	}()

	type summaryResult struct {
		summary *entities.FinancialSummary
		err     error
	}
	summaryCh := make(chan summaryResult, 1)
	go func() {
		summary, err := s.FinancialSrv.GetSummary(uid)
		summaryCh <- summaryResult{summary, err}
	}()
	userRes := <-userCh
	walletRes := <-walletCh
	summaryRes := <-summaryCh
	// 检查错误，如果有任意一个错误则返回
	var errs []error
	if userRes.err != nil {
		errs = append(errs, fmt.Errorf("failed to get user: %w", userRes.err))
	}
	if walletRes.err != nil {
		errs = append(errs, fmt.Errorf("failed to get wallet: %w", walletRes.err))
	}
	if summaryRes.err != nil {
		errs = append(errs, fmt.Errorf("failed to get summary: %w", summaryRes.err))
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("errors occurred: %v", errs)
	}
	// 确保指针非nil（假设底层方法在无错误时返回有效对象）
	var user = userRes.user
	if user == nil {
		return nil, errors.With("unexpected nil data without error")
	}

	if walletRes.wallet == nil {

		walletRes.wallet = &entities.UserWallet{UID: user.ID, PromoterCode: user.PromoterCode}
		s.walletSrv.CreateWallet(walletRes.wallet)
	}

	if summaryRes.summary == nil {
		summaryRes.summary = &entities.FinancialSummary{UID: user.ID, PromoterCode: user.PromoterCode}
		s.FinancialSrv.CreateFinancialSummary(summaryRes.summary)
	}

	profile := &entities.UserProfile{
		User:    userRes.user,
		Summary: summaryRes.summary,
		Wallet:  walletRes.wallet,
	}

	// 组合数据

	return profile, nil
}

func (s *UserService) GetUserByUID(UID uint) (user *entities.User, err error) {
	if UID == 0 {
		return nil, errors.WithCode(errors.AccountNotExist)
	}
	user, err = s.Repo.GetUserByID(UID)
	if err != nil {
		return
	}

	if user.Status == constant.USER_STATE_BLOCKED {
		err = errors.WithCode(errors.AccountBlocked)
		return
	}

	return user, nil
}

func (s *UserService) GetUserPC(UID uint) (int, error) {
	return s.Repo.GetUserPCByID(UID)
}

func (s *UserService) GetCustomer(uid uint) (*entities.Customer, error) {
	user, err := s.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}
	var customer entities.Customer
	admin, err := s.AdminSrv.GetSysUserAdmin(uint(user.PromoterCode))
	if err != nil {
		return nil, err
	}
	if admin != nil {
		customer.TelegramTeam = admin.TelegramTeam
		customer.TelegramCustomer = admin.TelegramCustomer
	}
	return &customer, nil
}

func (s *UserService) ClearUserCache(uid uint) error {
	s.walletSrv.ClearWalletCache(uid)
	return s.Repo.ClearUserCache(uid)
}

func (s *UserService) GetUserByname(username string) (*entities.User, error) {
	return s.Repo.GetUserByUsername(username)
}

func (s *UserService) GetUserByMobile(mobile string) (*entities.User, error) {
	return s.Repo.GetUserByMobile(mobile)
}

// 因为有redis 保存,所以回滚要注意
func (s *UserService) UpdateUser(user *entities.User) error {
	return s.Repo.UpdateUser(user)
}

func (s *UserService) UpdateUserNameAndGender(uid uint, nickname string, gender uint8) error {
	return s.Repo.UpdateUserNameAndGender(uid, nickname, gender)
}

func (s *UserService) UpdateUserTTL(uid uint) error {
	var expiration = time.Duration(constant.REDIS_USER_EXPIRE_TIME) * time.Second

	if err := s.Repo.UpdateUserTTL(uid, expiration); err != nil {
		return err
	}
	if err := s.walletSrv.UpdateWalletTTL(uid, expiration); err != nil {
		return err
	}
	return nil
}

func (s *UserService) UpdateAvatar(uid uint, file *multipart.FileHeader) error {

	// 检查文件类型
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return errors.With("Only JPG and PNG files are allowed")
	}

	src, err := file.Open()
	if err != nil {
		return errors.With("Failed to open file")
	}
	defer src.Close()

	// 生成唯一文件名
	objectName := fmt.Sprintf("%d%s", uid, ext)

	// 上传到 MinIO
	_, err = s.minioCli.PutObject(context.Background(), AvatarsBucketName, objectName, src, file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		logger.ZError("UpdateAvatar failed to upload file", zap.Error(err))
		return errors.With("Failed to upload file")
	}

	var url = config.Get().StorageSettings.GetURL()

	// 返回成功信息和文件URL
	fileURL := fmt.Sprintf("%s/%s/%s", url, AvatarsBucketName, objectName)

	logger.ZInfo("UpdateAvatar ", zap.String("url", fileURL))

	err = s.UpdateUser(&entities.User{ID: uid, Avatar: fileURL})
	if err != nil {
		return err
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *UserService) EditUserInfo(req *entities.EditUserInfoReq) (err error) {
	user, err := s.GetUserByUID(req.UID)
	if err != nil && !errors.IsAccountBlocked(err) {
		return err
	}
	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_EDIT_USER_INFO,
			OptionID: req.OptionID,
			IP:       req.IP,
			Data:     cjson.StringifyIgnore(req),
			Remark:   "修改用户信息",
		}

		if err != nil {
			log.Result = "failure"
		} else {
			log.Result = "success"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()

	userForUpdate := entities.User{}
	userForUpdate.ID = user.ID

	selects := make([]string, 0)
	// if req.BetLimit != user.BetAmountLimit {
	// 	userForUpdate.BetAmountLimit = req.BetLimit
	// 	selects = append(selects, "bet_limit")
	// }
	// if req.TimesLimit != user.BetTimesLimit {
	// 	userForUpdate.BetTimesLimit = req.TimesLimit
	// 	selects = append(selects, "times_limit")
	// }
	if req.Color != user.Color {
		userForUpdate.Color = req.Color
		selects = append(selects, "color")
	}
	if req.Password != nil && *req.Password != "" {
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		userForUpdate.Password = string(hashedNewPassword)
		selects = append(selects, "password")
	}

	if req.Status != user.Status {
		userForUpdate.Status = req.Status
		selects = append(selects, "status")
	}

	// if err := userForUpdate.CheckInvalid(); err != nil {
	// 	logger.ZError("EditUserInfo fail", zap.Any("user", userForUpdate), zap.Error(err))
	// 	return err
	// }

	if err := s.Repo.UpdateUserWithSelects(selects, &userForUpdate); err != nil {
		logger.ZError("EditUserInfo fail", zap.Any("user", userForUpdate), zap.Error(err))
		return err
	} else {
		logger.ZInfo("EditUserInfo succ", zap.Any("user", userForUpdate))
	}
	s.Repo.ClearUserCache(userForUpdate.ID) //直接清掉redis 缓存 用户获取信息会重新拉去

	if req.BalanceAdd > 0 || req.BalanceAdd < 0 { //有流水变动

		err = s.walletSrv.HandleWallet(req.UID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
			wallet.SafeAdjustCash(req.BalanceAdd)
			if wallet.Cash < 0 {
				return errors.With("user wallet cash less zero")
			}
			if err := s.walletSrv.UpdateCashWithTx(tx, wallet); err != nil {
				return err
			}

			createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
				UID:          user.ID,
				FlowType:     constant.FLOW_TYPE_GM_CASH,
				Number:       req.BalanceAdd,
				Balance:      wallet.Cash,
				PromoterCode: user.PromoterCode,
			})
			if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
				logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
			}

			return nil
		})
		return err
	}
	return nil
}

func (r *UserService) CheckAndAddTodayBetTimesLmit(uid string, timesLimit int) bool {
	todayBetTimes, err := r.Repo.GetBetTimes(uid)
	if err != nil {
		return true
	}
	if err := r.Repo.IncrementBetTimes(uid); err != nil {
		logger.ZError("CheckAndAddTodayBetTimesLmit IncrementBetTimes", zap.String("uid", uid))
	}
	return int(todayBetTimes) >= timesLimit
}

func (s *UserService) BatchClearUserCacheByPC(pcValue uint) error {
	// 获取所有 pc = 37 的用户 ID 列表
	userIDs, err := s.Repo.GetUserIDsByPC(pcValue)
	if err != nil {
		return err
	}
	// 如果没有找到用户，直接返回
	if len(userIDs) == 0 {
		return nil
	}

	logger.ZInfo("ExpireRDSUsersByPC", zap.Uint("pcValue", pcValue), zap.Int("userIDs", len(userIDs)))
	// 批量删除 Redis 中的用户数据（假设您已经实现了相应的方法）
	return s.Repo.BatchClearUserCache(userIDs)
}
