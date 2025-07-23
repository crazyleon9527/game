package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var (
	config Config
)

type ServiceSettings struct {
	ListenAddress    string `yaml:"ListenAddress"`
	StaticPath       string `yaml:"StaticPath"`
	Timezone         string `yaml:"Timezone"`
	VerificationType int    `yaml:"VerificationType"`
	JwtSignKey       string `yaml:"JwtSignKey"`
	TokenExpireTime  int    `yaml:"TokenExpireTime"`
	TokenRefreshTime int    `yaml:"TokenRefreshTime"`
	Area             string `yaml:"Area"`
	Language         string `yaml:"Language"`
	TrustedUserCode  string `yaml:"TrustedUserCode"`
	Environment      string `yaml:"Environment"` //'development','production'

	OAuth2CallbackHost string `yaml:"OAuth2CallbackHost"` //回调地址

	StorageHost string `yaml:"StorageHost"` //存储服务地址

	EnablePprof   bool `yaml:"EnablePprof"`   //是否注册Pprof
	EnableWingo   bool `yaml:"EnableWingo"`   //是否开启wingo
	EnableNine    bool `yaml:"EnableNine"`    //是否开启nine
	EnableTask    bool `yaml:"EnableTask"`    //是否开启任务列表
	EnableMQ      bool `yaml:"EnableMQ"`      //是否开启消息队列接受
	EnableSwagger bool `yaml:"EnableSwagger"` //是否开启swagger文档
}

func (s ServiceSettings) IsDevelopment() bool {
	return s.Environment == "debug"
}

type StorageSettings struct {
	Endpoint  string `yaml:"Endpoint"` //主库
	AccessKey string `yaml:"AccessKey"`
	SecretKey string `yaml:"SecretKey"`
	UseSSL    bool   `yaml:"UseSSL"`
}

func (s StorageSettings) GetURL() string {
	if s.UseSSL {
		return fmt.Sprintf("https://%s", s.Endpoint)
	} else {
		return fmt.Sprintf("http://%s", s.Endpoint)
	}
}

// https://gorm.io/zh_CN/docs/dbresolver.html 具体设置
type DBSettings struct {
	Driver                   string   `yaml:"Driver"`
	DataSource               string   `yaml:"DataSource"`               //主库
	DataSourceReplicas       []string `yaml:"DataSourceReplicas"`       //从库
	DataSourceSearchReplicas []string `yaml:"DataSourceSearchReplicas"` //读取库
	MaxIdleConns             int      `yaml:"MaxIdleConns"`
	MaxOpenConns             int      `yaml:"MaxOpenConns"`
	EnableAutoMigrate        bool     `yaml:"EnableAutoMigrate"`
	Trace                    bool     `yaml:"Trace"`
	AtRestEncryptKey         string   `yaml:"AtRestEncryptKey"`
	QueryTimeout             *int     `yaml:"QueryTimeout"`
}

type RDBSettings struct {
	UseCluster   bool     `yaml:"UseCluster"`
	ClusterAddrs []string `yaml:"ClusterAddrs"`
	Password     string   `yaml:"Password"`
	DB           int      `yaml:"DB"`
	PoolSize     int      `yaml:"PoolSize"`
	MinIdleConns int      `yaml:"MinIdleConns"`
	MaxConnAge   int      `yaml:"MaxConnAge"`  // In seconds
	PoolTimeout  int      `yaml:"PoolTimeout"` // In seconds
	IdleTimeout  int      `yaml:"IdleTimeout"` // In seconds
}

type CorsSettings struct {
	Enable           bool     `yaml:"Enable"`
	AllowOrigins     []string `yaml:"AllowOrigins"` //["GET", "POST", "PUT", "DELETE", "PATCH"]
	AllowMethods     []string `yaml:"AllowMethods"`
	AllowHeaders     []string `yaml:"AllowHeaders"`
	AllowCredentials bool     `yaml:"AllowCredentials"` //请求是否可以包含cookie，HTTP身份验证或客户端SSL证书等用户凭据
	MaxAge           int      `yaml:"MaxAge"`           //可以缓存预检请求结果的时间（以秒为单位）
}

type TelegramSetting struct {
	ApiToken  string `yaml:"ApiToken"`
	Proxy     string `yaml:"Proxy"`
	ManagerID int64  `yaml:"ManagerID"`
}

type ZfSetting struct {
	ApiUrl     string `yaml:"ApiUrl"`
	AppID      string `yaml:"AppID"`
	AppSecret  string `yaml:"AppSecret"`
	SignSecret string `yaml:"SignSecret"`
}

type JhszSetting struct {
	ApiUrl     string `yaml:"ApiUrl"`
	AppID      string `yaml:"AppID"`
	AppSecret  string `yaml:"AppSecret"`
	SignSecret string `yaml:"SignSecret"`
}

type R8Setting struct {
	ApiUrl string `yaml:"ApiUrl"`
	AppID  string `yaml:"AppID"`
	AppKey string `yaml:"AppKey"`
}
type ChainSetting struct {
	ChainGameHost string `yaml:"ChainGameHost"`
	GameURL       string `yaml:"GameURL"`
}

type QuizSetting struct {
	EventLimit        uint   `yaml:"EventLimit" default:"1"`
	EventStartOffset  uint   `yaml:"EventStartOffset" default:"5"`
	EventEndOffsetMin uint   `yaml:"EventEndOffsetMin" default:"5"`
	EventEndOffsetMax uint   `yaml:"EventEndOffsetMax" default:"10"`
	ClobEndpoint      string `yaml:"ClobEndpoint" default:"https://clob.polymarket.com"`
	GammaEndpoint     string `yaml:"GammaEndpoint" default:"https://gamma-api.polymarket.com"`
}

type Config struct {
	ServiceSettings ServiceSettings `json:"ServiceSettings" yaml:"ServiceSettings,omitempty"`
	DBSettings      DBSettings      `json:"DBSettings" yaml:"DBSettings,omitempty"`
	RDBSettings     RDBSettings     `json:"RDBSettings" yaml:"RDBSettings,omitempty"`
	CorsSettings    CorsSettings    `json:"CorsSettings" yaml:"CorsSettings,omitempty"`
	TelegramSetting TelegramSetting `json:"TelegramSetting" yaml:"TelegramSetting,omitempty"`
	ZfSetting       ZfSetting       `json:"ZfSetting" yaml:"ZfSetting,omitempty"`
	R8Setting       R8Setting       `json:"R8Setting" yaml:"R8Setting,omitempty"`
	JhszSetting     ZfSetting       `json:"JhszSetting" yaml:"JhszSetting,omitempty"`
	ChainSetting    ChainSetting    `json:"ChainSetting" yaml:"ChainSetting,omitempty"`
	StorageSettings StorageSettings `json:"StorageSettings" yaml:"StorageSettings,omitempty"`
	QuizSetting     QuizSetting     `json:"QuizSetting" yaml:"QuizSetting,omitempty"`
}

func Get() Config {
	return config
}

func MustLoad(filePath string) (err error) {
	var data []byte
	if data, err = ioutil.ReadFile(filePath); err != nil {
		err = errors.New(fmt.Sprintf("failed read config file: %s \n", err))
		return
	}
	if err = yaml.Unmarshal(data, &config); err != nil {
		err = errors.New(fmt.Sprintf("failed unmarshal config file: %s \n", err))
	}
	return
}

func PrintWithJSON() string {
	jsonBytes, err := json.Marshal(Get())
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	jsonString := string(jsonBytes)
	return jsonString
}
