package rds

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var structCache sync.Map // 缓存反射信息

type fieldInfo struct {
	RedisTag string
	Index    int
}

func StructToRedisHashOptimized(input interface{}) (map[string]string, error) {
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct")
	}

	t := v.Type()
	cached, _ := structCache.LoadOrStore(t, buildFieldCache(t))
	cache := cached.([]fieldInfo)

	result := make(map[string]string, len(cache))
	for _, info := range cache {
		fValue := v.Field(info.Index)
		if isZero(fValue) {
			continue
		}

		strVal, err := toString(fValue)
		if err != nil {
			return nil, err
		}
		result[info.RedisTag] = strVal
	}
	return result, nil
}

func buildFieldCache(t reflect.Type) []fieldInfo {
	var cache []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		redisTag := field.Tag.Get("redis")
		if redisTag == "" || redisTag == "-" {
			continue
		}
		cache = append(cache, fieldInfo{
			RedisTag: redisTag,
			Index:    i,
		})
	}
	return cache
}

// 零值检测增强
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	// case reflect.Bool:
	// 	return !v.Bool()
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 	return v.Int() == 0
	// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	// 	return v.Uint() == 0
	// case reflect.Float32, reflect.Float64:
	// 	return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.IsZero()
		}
		return false // 非时间结构体不视为零值
	default:
		// 不支持的类型（如 complex、unsafe.Pointer 等）
		return false
	}
}

// 统一字符串转换
func toString(v reflect.Value) (string, error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format(time.RFC3339Nano), nil
		}
		// 处理自定义结构体
		if jsonData, err := json.Marshal(v.Interface()); err == nil {
			return string(jsonData), nil
		}
		return "", fmt.Errorf("unsupported struct type: %s", v.Type())
	default:
		return "", fmt.Errorf("unsupported type: %s", v.Type())
	}
}
