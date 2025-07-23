package test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

// 广告系列（Campaign）是广告投放的最高层级，通常是根据广告目标或主题进行组织。
//每个广告系列都有一个独立的广告预算，广告系列的目的是为了帮助你达到你的广告目标。

// 广告组（Ad Set）是广告系列的子层级，通常是根据受众、预算和广告投放时间等因素进行组织。
//每个广告组都有一个独立的预算，用于管理广告投放时间、目标受众、竞价等投放细节。

// 广告创意（Ad）是广告组的下一层级，通常包括广告的文案、图片、视频等内容。

//广告创意可以在多个广告组中使用，以测试哪些广告创意最有效，最终优化广告投放效果。

var (
	AppID       = "899413861170606"
	AppSecret   = "2661493cc74a7f23ef4f9b4b189f3e58"
	AccessToken = "EAAMyAxuzfa4BALdZBtiReW3KlfJZB0M5S0TLOOcjFzvmOsHg3dtZA6QOlPmjJPcxfN6HAw7I5Pz4Wzwe2Du2w2Nc0zPyj1kHWvbwGx52eatPGm6W8nkgIcLOKBufDkbc5ZCWbHDONnsRMsSHCOLuQ1IPIoiPZCSRfqj2OgFjJmiFy089VF4Yq"
	AdAccountID = "421176839456533"
)

// var (
// 	AppID       = "5214995511936199"
// 	AppSecret   = "6212b170ae19e3ac676e4b81958bb532"
// 	AccessToken = "EABKHAsLZBhMcBACSYW2ZCZCYeFx4ZBk2iNqJs0DhjrhWofyYwbairnNQV1NCUZCksmEzJHXjqovOsMZCYT8op8BkZAJ5ZBPVXdUT4OAhZAXoRZAX9xNaYXD0t4kgLOrnM2zGo0xL5Xi7us44EiCRj2okf0Pa7iHC6RRDIy1TtAJ9ZCJGG6IzVm8q739ERzO9YSEGx5xwyhHtEG4oAZDZD"
// 	AdAccountID = "421176839456533"
// )

type Logger struct {
}

func (c Logger) Log(keyvals ...interface{}) error {
	log.Println(keyvals...)
	return nil
}

func TestFB(t *testing.T) {
	// fmt.Println("testf2")

	// config.MustLoad("../configs/config.yaml")
	// // minioClient, err := storage.InitMinio(config.Get().StorageSetting.Minio)
	// // if err != nil {
	// // 	panic("failed to init minio")
	// // }

	// fbService, err := v16.New(Logger{}, AccessToken, AppSecret)
	// fmt.Println(fbService, err)
	// fbService.CustomConversions.
}

func imageFB(imagePath string) {
	// 指定本地图片文件路径

	// 创建一个带有访问令牌的HTTP客户端
	accessToken := AccessToken
	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", "https://graph.facebook.com/v16.0/act_421176839456533/adimages", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 创建一个multipart/form-data类型的请求体，用于上传图片文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("filename", imagePath)
	if err != nil {
		fmt.Println("Error creating file writer:", err)
		return
	}
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Println("Error opening image file:", err)
		return
	}
	defer file.Close()
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		fmt.Println("Error copying file data:", err)
		return
	}
	contentType := writer.FormDataContentType()
	writer.Close()

	// 发送请求，并解析返回的JSON数据，获取image_hash
	req.Header.Set("Content-Type", contentType)
	req.Body = ioutil.NopCloser(body)
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println(string(resBody))
}

// c := v16.Campaign{
// 	// Populate struct values
// 	Status:              "PAUSED",
// 	Objective:           "LINK_CLICKS",
// 	Name:                "baidu102",
// 	AccountID:           AdAccountID,
// 	SpecialAdCategories: make([]string, 1),
// }
// id, err := fbService.Campaigns.Create(context.Background(), c)

// log.Println(id, err)

// a := v16.AdCreative{
// 	AccountID: AdAccountID,
// 	Name:      "ad 01",

// 	ObjectStorySpec: &v16.ObjectStorySpec{
// 		PageID: "100172616372468",
// 		LinkData: &v16.AdCreativeLinkData{
// 			// Caption: "Check out my new product!",
// 			Description: "This is the best product ever!",
// 			ImageHash:   "b15cfe3114ebfd0a8faad6302c8904fc",
// 			Link:        "https://www.example.com",
// 			Message:     "Buy now and get 20% off!",
// 			CallToAction: &v16.AdCreativeLinkDataCallToAction{
// 				Type: "SHOP_NOW",
// 				Value: &v16.AdCreativeLinkDataCallToActionValue{
// 					Link: "https://www.example.com",
// 				},
// 			},
// 		},
// 	},
// }

// adID, adID2, err := fbService.AdCreatives.Create(context.Background(), a)

// log.Println(adID, adID2, err)

// objectName := "1680453807123885100-woshiagenting.jpg"

// object, err := minioClient.GetObject(context.Background(), config.Get().StorageSetting.Minio.Bucket, objectName, minio.GetObjectOptions{})
// if err != nil {
// 	log.Fatalln(err)
// }

// file, err := os.Open("../veer-372794457.jpg")
// if err != nil {
// 	fmt.Println("Error opening image file: ", err)
// 	return
// }

// // 将对象内容复制到bytes.Buffer中
// bodyBuf := &bytes.Buffer{}
// _, err = io.Copy(bodyBuf, object)
// if err != nil {
// 	fmt.Println(err)
// 	return
// }
// r := bytes.NewReader(bodyBuf.Bytes())

// img, err := fbService.Images.Upload(context.Background(), AdAccountID, "veer-372794457.jpg", r)
// // img, err := fbService.Images.Upload(context.Background(), AdAccountID, "veer-372794457.jpg", file)
// fmt.Println(img, err)

// defer file.Close()

// imageFB("../veer-372794457.jpg")

// list, err := fbService.Campaigns.List(AdAccountID).Do(context.Background())
// log.Println(list, err)

// fbService.AdCreatives.Create(ctx context.Context, a v14.AdCreative)

// Link caption是指链接的标题下方显示的文本，用于描述链接内容。Link caption必须是一个实际的URL，
// 并且应准确反映用户点击链接时访问的URL和相关的广告主或业务。具体来说，Link caption需要描述链接所指向的网页的内容，以便让用户了解链接的含义和目的。

// 需要注意的是，Link caption不适用于Instagram。Instagram不允许在链接上添加额外的描述文本，因此在创建广告时，不需要设置Link caption。

// session := fb.NewSession("<access_token>")
// file, err := os.Open("<path_to_image_file>")
// if err != nil {
// 	fmt.Println("Error opening image file: ", err)
// 	return
// }
// defer file.Close()

// image, err := fb.CreateImage(file, session)
// if err != nil {
// 	fmt.Println("Error creating image: ", err)
// 	return
// }
// fmt.Println("Image Hash: ", image.Hash)

// curl \
//   -F 'name=Sample Creative' \
//   -F 'object_story_spec={
//     "link_data": {
//       "link": "https://www.baidu.com",
//       "message": "try it out"
//     },
//     "page_id": "100172616372468"
//   }' \

//     adCreative := fb.AdCreative{
//         Name: "My Ad Creative",
//         ObjectStorySpec: fb.ObjectStorySpec{
//             PageID: "<page_id>",
//             LinkData: &fb.LinkData{
//                 Caption: "Check out my new product!",
//                 Description: "This is the best product ever!",
//                 ImageHash: "<image_hash>",
//                 Link: "https://www.example.com",
//                 Message: "Buy now and get 20% off!",
//                 CallToAction: &fb.LinkDataCallToAction{
//                     Type: "SHOP_NOW",
//                     Value: &fb.LinkDataCallToActionValue{
//                         Link: "https://www.example.com",
//                     },
//                 },
//             },
//         },
//     }

// func main() {
//     session := fb.NewSession("<access_token>")
//     campaign := fb.Campaign{
//         Name: "My Campaign",
//         Objective: fb.CampaignObjectiveReach,
//     }
//     campaignID, err := fb.CreateCampaign(campaign, fb.AdAccountID("<ad_account_id>"), session)
//     if err != nil {
//         fmt.Println("Error creating campaign: ", err)
//         return
//     }
//     fmt.Println("Campaign ID: ", campaignID)

//     adSet := fb.AdSet{
//         Name: "My Ad Set",
//         CampaignID: campaignID,
//         OptimizationGoal: fb.AdSetOptimizationGoalReach,
//         BillingEvent: fb.AdSetBillingEventImpressions,
//         BidAmount: 100,
//         DailyBudget: 5000,
//         StartTime: "<start_time>",
//         EndTime: "<end_time>",
//         Targeting: fb.Targeting{
//             GeoLocations: &fb.TargetingGeoLocation{
//                 Countries: []string{"US"},
//             },
//             AgeMin: 18,
//             AgeMax: 65,
//             Gender: 1,
//         },
//     }
//     adSetID, err := fb.CreateAdSet(adSet, fb.AdAccountID("<ad_account_id>"), session)
//     if err != nil {
//         fmt.Println("Error creating ad set: ", err)
//         return
//     }
//     fmt.Println("Ad Set ID: ", adSetID)

//     adCreative := fb.AdCreative{
//         Name: "My Ad Creative",
//         ObjectStorySpec: fb.ObjectStorySpec{
//             PageID: "<page_id>",
//             LinkData: &fb.LinkData{
//                 Caption: "Check out my new product!",
//                 Description: "This is the best product ever!",
//                 ImageHash: "<image_hash>",
//                 Link: "https://www.example.com",
//                 Message: "Buy now and get 20% off!",
//                 CallToAction: &fb.LinkDataCallToAction{
//                     Type: "SHOP_NOW",
//                     Value: &fb.LinkDataCallToActionValue{
//                         Link: "https://www.example.com",
//                     },
//                 },
//             },
//         },
//     }
//     adCreativeID, err := fb.CreateAdCreative(adCreative, fb.AdAccountID("<ad_account_id>"), session)
//     if err != nil {
//         fmt.Println("Error creating ad creative: ", err)
//         return
//     }
//     fmt.Println("Ad Creative ID: ", adCreativeID)

//     ad := fb.Ad{
//         Name: "My Ad",
//         AdSetID: adSetID,
//         Creative: fb.AdCreativeData{
//             CreativeID: adCreativeID,
//         },
//         Status: fb.AdStatusActive,
//     }
//     adID, err := fb.CreateAd(ad, fb.AdAccountID("<ad_account_id>"), session)
//     if err != nil {
//         fmt.Println("Error creating ad: ", err)
//         return
//     }
//     fmt.Println("Ad ID: ", adID)
// }
