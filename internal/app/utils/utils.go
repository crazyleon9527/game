package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"regexp"
	"rk-api/pkg/logger"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/text/language"

	"github.com/davecgh/go-spew/spew"
)

func PrintPanicStack(extras ...interface{}) {
	if x := recover(); x != nil {
		logger.Error("Recovered from panic:", x)
		// 直接获取完整的调用堆栈
		stack := debug.Stack()
		logger.Errorf("PANIC STACKTRACE:\n%s\n", stack)

		// 打印额外的信息
		for k := range extras {
			logger.Errorf("EXTRAS#%v DATA:%v\n", k, spew.Sdump(extras[k]))
		}
	}
}

// GetPreferredLanguage 返回用户首选语言的语言标记
// 如果无法解析"Accept-Language"头部字段，将返回默认语言标记
func GetPreferredLanguage(acceptLanguageHeader string) language.Tag {
	acceptLanguageTags, _, err := language.ParseAcceptLanguage(acceptLanguageHeader)
	if err != nil {
		return language.Chinese // 返回默认语言标记
	}
	matcher := language.NewMatcher([]language.Tag{language.English, language.French, language.SimplifiedChinese})
	tag, _, _ := matcher.Match(acceptLanguageTags...)
	return tag
}

// 匹配出域名 ,兼容
func MatchDomain(domain string) (string, error) {
	// 定义正则表达式
	re := regexp.MustCompile(`(?:https?://)?([^/]+)`)
	// 查找所有匹配的结果
	matches := re.FindStringSubmatch(domain)

	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", errors.New("domain format is error")
}

func GetSelfFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return cleanUpFuncName(runtime.FuncForPC(pc).Name())
}

func cleanUpFuncName(funcName string) string {
	end := strings.LastIndex(funcName, ".")
	if end == -1 {
		return ""
	}
	return funcName[end+1:]
}

func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func ReverseSlice(data interface{}) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Slice {
		panic(errors.New("data must be a slice type"))
	}
	valueLen := value.Len()
	for i := 0; i <= int((valueLen-1)/2); i++ {
		reverseIndex := valueLen - 1 - i
		tmp := reflect.ValueOf(value.Index(i).Interface())
		value.Index(i).Set(value.Index(reverseIndex))
		value.Index(reverseIndex).Set(tmp)
	}
}

// EncodeMD5 生成 MD5
func EncodeMD5(value string) string {
	m := md5.New()
	m.Write([]byte(value))
	return hex.EncodeToString(m.Sum(nil))
}

func IsValidIndianMobile(number string) bool {
	// 匹配 "+91" 开头，后跟 10 位数字的模式
	pattern := `^\+91\d{10}$`
	// 匹配 "+91" 开头，后跟以 7、8 或 9 开始的 10 位数字
	// pattern := `^\+91[789]\d{9}$`
	// 编译正则表达式
	regex := regexp.MustCompile(pattern)
	// 使用正则表达式匹配号码
	return regex.MatchString(number)
}

func IsValidMobile(number string) bool {
	// return IsValidIndianMobile(number)
	return true
}

// 判断是否为邮箱
func IsValidEmail(input string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, input)
	return matched
}

func IsValidUsername(username string) bool {
	// 字符串只能包含字母、数字和下划线
	regex := "^[a-zA-Z0-9_]+$"
	re := regexp.MustCompile(regex)
	return re.MatchString(username)
}

// ValidatePassword 验证密码是否符合规则
// func ValidatePassword(password string) error {
// 	// 最小长度
// 	if len(password) < 8 {
// 		return fmt.Errorf("password must be at least 8 characters long")
// 	}

// 	// 最大长度
// 	if len(password) > 64 {
// 		return fmt.Errorf("password must be no more than 64 characters long")
// 	}

// 	// 包含大写字母
// 	upperCaseRegex := regexp.MustCompile(`[A-Z]`)
// 	if !upperCaseRegex.MatchString(password) {
// 		return fmt.Errorf("password must contain at least one uppercase letter")
// 	}

// 	// 包含小写字母
// 	lowerCaseRegex := regexp.MustCompile(`[a-z]`)
// 	if !lowerCaseRegex.MatchString(password) {
// 		return fmt.Errorf("password must contain at least one lowercase letter")
// 	}

// 	// 包含数字
// 	digitRegex := regexp.MustCompile(`[0-9]`)
// 	if !digitRegex.MatchString(password) {
// 		return fmt.Errorf("password must contain at least one digit")
// 	}

// 	return nil
// }

func HmacSHA256(key, message string) string {
	// 使用HmacSHA256签名方法对待签名字符串进行签名
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature
}
