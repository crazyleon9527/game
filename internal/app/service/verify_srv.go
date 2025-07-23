package service

import (
	"bytes"
	"fmt"
	"math/rand"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/email"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/sms"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
)

// Transform to Uppercase
const (
	VERIFIEDCODE_TEMPLATE = "Dear user, your verfiy code is %s." //发送验证码模板
)

var VerifyServiceSet = wire.NewSet(
	ProvideVerifyService,
)

type VerifyService struct {
	Repo     *repository.UserRepository
	StateSrv *StateService
	// UserLocks  *entities.RedisUserLock
	smsImplMap  map[string]sms.IMessage
	mailImplMap map[string]email.IProvider
	smsChannel  string
	mailChannel string
}

func ProvideVerifyService(repo *repository.UserRepository,
	adminSrv *AdminService,
	stateSrv *StateService,
) *VerifyService {

	smsImplMap := map[string]sms.IMessage{
		"plant": sms.NewPlantSms("c9Lcj4dy", "4VDwxX32", "http://api.nxcloud.com/api/sms/mtsend"),
		"i51":   sms.NewI51Sms("dn+citpRTz2fUBzPY+8g0w==", "d5b98058c4a44608a7db9a5cad5de002", "https://api.i51sms.com/outVerify/verifCodeSend"),
	}

	mailImplMap := map[string]email.IProvider{
		"mailgun": email.NewMailgunProvider(email.SMTPConfig{
			Server:   "smtp.mailgun.org",                                               // Mailgun SMTP 服务器地址
			Port:     587,                                                              // Mailgun SMTP 端口
			Username: "postmaster@sandbox775c971afa9f488dbf71d47eecde1f4f.mailgun.org", // Mailgun SMTP 用户名
			Password: "b3f7062f8c9a9b4053efb3473a6ba82b-9c3f0c68-34b2534d",             // Mailgun SMTP 密码
			From:     "no-reply@sandbox775c971afa9f488dbf71d47eecde1f4f.mailgun.org",   // 发件人邮箱（必须是有效地址）
			Nickname: "test",                                                           // 发件人昵称
		}),
		"outlook": email.NewOutlookProvider(email.SMTPConfig{
			Server:   "smtp.office365.com",
			Port:     587,
			Username: "181058363@qq.com",
			Password: "rawtnocjcwfnfoau", // 此处为生成的应用密码
			From:     "181058363@qq.com",
			Nickname: "MyApp",
		}),
		"zoho": email.NewZohoProvider(email.SMTPConfig{
			Server:   "smtp.zoho.com",
			Port:     465,
			Username: "yangqing520@zohomail.com",
			Password: "C3XbETYq5f41", // 此处为生成的应用密码
			From:     "yangqing520@zohomail.com",
			Nickname: "yangqing2",
		}),
	}

	service := &VerifyService{
		Repo:        repo,
		StateSrv:    stateSrv,
		smsImplMap:  smsImplMap,
		mailImplMap: mailImplMap,
		smsChannel:  "plant",
		mailChannel: "zoho",
	}
	return service
}

func makeVerificationCode() string {
	// 生成一个 10000 ~ 99999 之间的随机数（5 位数）
	code := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(90000) + 10000)
	return code
}

func (s *VerifyService) CheckVerifyCode(mobile string, code string) (err error) {

	if code != config.Get().ServiceSettings.TrustedUserCode { //判断验证码 todo redis
		verifyCode := new(entities.VerifyCode)
		verifyCode.Code = code
		verifyCode.Target = mobile
		verifyCode, err = s.Repo.GetVerifyCode(verifyCode)
		if err != nil {
			return errors.WithCode(errors.VerifiedCodeNotMatch) // todo 解析
		}
		if verifyCode.VerificationType == constant.VERIFICATION_TYPE_EMAIL {
			if time.Now().Unix()-verifyCode.UpdatedAt > constant.MAX_EMAIL_TOKEN_EXIPRY_TIME {
				return errors.WithCode(errors.VerifiedCodeExpire)
			}
		}

		if verifyCode.VerificationType == constant.VERIFICATION_TYPE_SMS {
			if time.Now().Unix()-verifyCode.UpdatedAt > constant.MAX_TOKEN_EXIPRY_TIME {
				return errors.WithCode(errors.VerifiedCodeExpire)
			}
		}
	}
	return
}

// /短信类型，1为绑定手机号，2为重置用户密码，3重置支付密码，4用户注册，5、解绑定银行卡，6、解绑定支付宝 7、用户登录 8、更换手机号
func (s *VerifyService) SendVerifyCode(verifyCode *entities.VerifyCode) error {

	err := s.Repo.CreatorOrGetVerifyCodeByEmailOrMobile(verifyCode)

	if err != nil {
		return err
	}

	if verifyCode.Code != "" { //已经有记录
		if time.Now().Unix()-verifyCode.UpdatedAt < constant.MAX_TOKEN_RETRY_EXIPRY_TIME {
			return errors.WithCode(errors.RetryFrequenceLimit)
		}

		if time.Unix(verifyCode.UpdatedAt, 0).Day() != time.Now().Day() {
			verifyCode.Count = 0 //重置为0
		}
		if verifyCode.Count > constant.VERIFICATION_CODE_DAY_LIMIT {
			return errors.WithCode(errors.RetryCountLimit)
		}
	}

	verifyCode.Code = makeVerificationCode()

	if verifyCode.VerificationType == constant.VERIFICATION_TYPE_SMS { //使用短信方式
		if !utils.IsValidMobile(verifyCode.Target) { //判断是否为手机号
			return errors.WithCode(errors.InvalidMobile)
		}
		content := fmt.Sprintf(VERIFIEDCODE_TEMPLATE, verifyCode.Code) //根据类型选择不同的模板
		_, err = s.SendOptSms(verifyCode.Target, content)
		if err != nil {
			logger.ZError("SendOptSms", zap.Any("verifyCode", verifyCode), zap.Error(err))
			return errors.With("Verification code failed to send")
		}
	} else if verifyCode.VerificationType == constant.VERIFICATION_TYPE_EMAIL { //使用邮箱方式发送验证码

		content, err := s.RenderEmailTemplate(EmailData{
			Code:         verifyCode.Code,
			BusinessType: "注册",
		})
		if err != nil {
			return err
		}

		err = s.SendOptEmail(verifyCode.Target, "", content) //补充subject
		if err != nil {
			logger.ZError("SendMail", zap.Any("verifyCode", verifyCode), zap.Error(err))
			return errors.With("Verification code failed to send")
		}
	}

	verifyCode.UpdatedAt = time.Now().Unix()
	verifyCode.Count++
	return s.Repo.UpdateVerifyCode(verifyCode)
}

func (s *VerifyService) SendOptSms(phone string, msg string) (string, error) {
	normalizePhone := strings.TrimPrefix(phone, "+") //去掉前面的加号

	sms, ok := s.smsImplMap[s.smsChannel]
	if !ok {
		return "", errors.With(fmt.Sprintf("not init sms impl (%s)", s.smsChannel))
	}

	return sms.Send(normalizePhone, msg)

}

func (s *VerifyService) SwitchSmsChannel(channel string) error {
	_, ok := s.smsImplMap[channel]
	if !ok {
		return errors.With(fmt.Sprintf("not init sms impl (%s)", s.smsChannel))
	}
	s.smsChannel = channel
	return nil
}

func (s *VerifyService) GetSMSVerificationState() bool {
	return s.StateSrv.GetBoolState(constant.StateSMSVerificationDisabled)
}

func (s *VerifyService) ToggleSMSVerificationState(req *entities.ToggleSMSVerificationStateReq) error {

	s.StateSrv.SetState(constant.StateSMSVerificationDisabled, !s.StateSrv.GetBoolState(constant.StateSMSVerificationDisabled))

	logger.ZInfo("ToggleSMSVerificationState", zap.Any("req", req), zap.Bool("state", s.StateSrv.GetBoolState(constant.StateSMSVerificationDisabled)))
	return nil
}

type EmailData struct {
	Code         string // 验证码
	BusinessType string // 业务类型
}

// RenderEmailTemplate 渲染邮件模板
func (s *VerifyService) RenderEmailTemplate(data EmailData) (string, error) {
	tmpl, err := template.ParseFiles("public/template/email_template.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *VerifyService) SendOptEmail(to string, subject string, content string) error {
	mail, ok := s.mailImplMap[s.mailChannel]
	if !ok {
		return errors.With(fmt.Sprintf("not init mail impl (%s)", s.mailChannel))
	}

	return mail.Send([]string{to}, subject, content)
}
