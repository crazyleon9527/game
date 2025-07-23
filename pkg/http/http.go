package http

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// GetHttpClient 获取请求客户端
func GetHttpClient(proxys ...string) *resty.Client {
	client := resty.New()
	// 如果有代理
	if len(proxys) > 0 {
		proxy := proxys[0]
		client.SetProxy(proxy)
	}
	client.SetTimeout(time.Second * 10)
	return client
}

func SendPost(client *resty.Client, url string, data interface{}, isJson bool) (map[string]interface{}, error) {
	var resp *resty.Response
	var err error

	if isJson {
		resp, err = client.R().
			SetHeader("Content-Type", "application/json; charset=utf-8").
			SetBody(data).
			Post(url)

		// resp, err = client.R().SetHeader("Content-Type", "application/json").SetBody(data).Post(url)
	} else {
		resp, err = client.R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetFormData(data.(map[string]string)).
			Post(url)
	}

	// logger.ZError("send post",
	// 	zap.String("url", url),
	// 	zap.Any("req", data),
	// 	zap.Any("resp", resp),
	// 	zap.Error(err))

	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SendGet 发送GET请求，并返回解析后的JSON
func SendGet(client *resty.Client, endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
	request := client.R()
	for key, val := range params {
		// 根据参数的实际类型设置查询参数
		switch v := val.(type) {
		case string:
			request.SetQueryParam(key, v)
		case int:
			request.SetQueryParam(key, strconv.Itoa(v))
		case float64:
			request.SetQueryParam(key, strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			request.SetQueryParam(key, strconv.FormatBool(v))
		default:
			request.SetQueryParam(key, fmt.Sprintf("%v", v))
		}
	}

	// 发送请求并获取响应
	resp, err := request.Get(endpoint)
	if err != nil {
		return nil, err
	}

	// // // 检查返回的HTTP状态码是否为200
	// if resp.StatusCode() != http.StatusOK {
	// 	// 如果不是200，将服务器的响应状态码和消息返回
	// 	// logger.Error("-----------------------", resp.StatusCode(), http.StatusOK)
	// }

	var result map[string]interface{}
	// 解析服务器返回的JSON数据
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		// JSON解析错误
		return nil, err
	}
	return result, nil
}
