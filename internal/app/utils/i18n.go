package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/i18n"
)

var (
	T           i18n.TranslateFunc
	translators map[string]string = make(map[string]string)
)

func TranslationsPreInit(dir string) error {
	if err := InitTranslationsWithDir(dir); err != nil {
		return err
	}
	T = TfuncWithFallback("en") //默认
	return nil
}

func InitTranslationsWithDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range files {
		if !entry.IsDir() { //
			filename := entry.Name()
			if filepath.Ext(filename) == ".json" {
				translatorKey := strings.TrimSuffix(filename, filepath.Ext(filename))
				translationFilePath := filepath.Join(dir, filename)

				translators[translatorKey] = translationFilePath
				if err := i18n.LoadTranslationFile(translationFilePath); err != nil {
					return err // handle error when loading a translation file
				}
			}
		}
	}
	return nil
}

func GetUserTranslations(locales ...string) i18n.TranslateFunc {
	for _, locale := range locales {
		if _, found := translators[strings.ToLower(locale)]; found {
			return TfuncWithFallback(locale)
		}
	}
	return T
}

func SetTranslations(locale string) i18n.TranslateFunc {
	translations := TfuncWithFallback(locale)
	return translations
}

func TfuncWithFallback(pref string) i18n.TranslateFunc {
	t, _ := i18n.Tfunc(pref)
	return func(translationID string, args ...interface{}) string {
		if translated := t(translationID, args...); translated != translationID {
			return translated
		}
		// t, _ := i18n.Tfunc(DEFAULT_LOCALE)
		// return t(translationID, args...)
		return translationID
	}
}
