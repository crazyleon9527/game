package utils

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni            *ut.UniversalTranslator
	once           sync.Once
	supportLocales []string

	customTagTranslations []struct { //自定义tag 的翻译
		tag   string
		trans map[string]string
	}
)

func init() {
	supportLocales = []string{"en", "zh"}
	uni = ut.New(en.New(), en.New(), zh.New()) //默认是英文, 支持额是英文，中文

	customTagTranslations = []struct {
		tag   string
		trans map[string]string
	}{
		{
			tag: "e164",
			trans: map[string]string{
				"zh": "{0}格式不正确，正确的格式应为E.164国际电话号码格式",
				// "en": "{0} is not in the correct format, the correct format should be E.164 international phone number format",
			},
		},
	}

}

func RegisterTranslations(v *validator.Validate) (err error) {
	translator_en, found := uni.GetTranslator("en")
	if found {
		if err := en_translations.RegisterDefaultTranslations(v, translator_en); err != nil {
			fmt.Println("en register translations:", err)
			return err
		}
	}
	translator_zh, found := uni.GetTranslator("zh")
	if found {
		if err := zh_translations.RegisterDefaultTranslations(v, translator_zh); err != nil {
			fmt.Println("zh register translations:", err)
			return err
		}
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return "{{" + name + "}}" //
	})
	//注册自定义
	for _, t := range customTagTranslations {
		for _, locale := range supportLocales {
			if t.trans[locale] != "" {
				RegisterTranslation(v, locale, t.tag, t.trans[locale])
			}
		}
	}

	return nil
}

// 注册自定义翻译
func RegisterTranslation(v *validator.Validate, locale string, tag string, translation string) {
	trans, found := uni.GetTranslator(locale)
	if found {
		v.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, translation, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T(fe.Tag(), fe.Field())
			if err != nil {
				log.Printf("warning: error translating FieldError: %#v", fe)
				return fe.(error).Error()
			}
			return t
		})
	}
}

// 获得一个新的验证器,项目采用 gin 自带的
func GetValidate() *validator.Validate {
	v := validator.New()
	RegisterTranslations(v)
	return v
}

// // 自定义字段名称
// // 参考自 : https://github.com/syssam/go-playground-sample/blob/master/main.go
//
//	// fieldNames := [...]string{"username", "password", "email", "UserId", "verCode", "mobile"}
func RegisterFieldTranslation(fieldNames ...string) {
	for _, locale := range supportLocales {
		trans, _ := uni.GetTranslator(locale)
		localTrans := GetUserTranslations(locale)
		for _, fieldName := range fieldNames {
			fieldTag := fmt.Sprintf("{{%v}}", fieldName)
			trans.Add(fieldTag, localTrans(fieldName), false)
		}
	}
}

// GetTrans 获取翻译
func GetTrans(locale ...string) ut.Translator {
	trans, _ := uni.FindTranslator(locale...)
	return trans
}
