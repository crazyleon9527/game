package structure

import (
	"rk-api/pkg/cjson"

	"github.com/jinzhu/copier"
)

// Copy 结构体映射
func Copy(s, ts interface{}) error {
	return copier.Copy(ts, s)
}

func CopyIgnoreEmpty(s, ts interface{}) error {
	return copier.CopyWithOption(ts, s, copier.Option{IgnoreEmpty: true, DeepCopy: false})
}

func MapToStruct(m interface{}, s interface{}) error {
	jsonData, err := cjson.Cjson.Marshal(m)
	if err != nil {
		return err
	}
	// 再将JSON字节解码到指定的struct中
	err = cjson.Cjson.Unmarshal(jsonData, &s)
	if err != nil {
		return err
	}
	return nil

	// jsonData, err := json.Marshal(m)
	// if err != nil {
	// 	return err
	// }
	// // 再将JSON字节解码到指定的struct中
	// err = json.Unmarshal(jsonData, &s)
	// if err != nil {
	// 	return err
	// }
	// return nil

	// return mapstructure.Decode(m, s)
}

func StructToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := cjson.Cjson.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = cjson.Cjson.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
