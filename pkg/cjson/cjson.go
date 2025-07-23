package cjson

import jsoniter "github.com/json-iterator/go"

var Cjson = jsoniter.ConfigCompatibleWithStandardLibrary

func Stringify(v interface{}) (string, error) {
	return Cjson.MarshalToString(v)
}

func Parse(str string, v interface{}) error {
	return Cjson.UnmarshalFromString(str, v)
}

func StringifyIgnore(v interface{}) string {
	str, err := Cjson.MarshalToString(v)
	if err != nil {
		return ""
	}
	return str
}
