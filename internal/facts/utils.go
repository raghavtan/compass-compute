package facts

import (
	"strconv"
)

func convertToFloat64(value interface{}) *float64 {
	switch v := value.(type) {
	case float64:
		return &v
	case float32:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	case bool:
		if v {
			f := 1.0
			return &f
		}
		f := 0.0
		return &f
	}
	return nil
}

func convertToBool(value interface{}) *bool {
	switch v := value.(type) {
	case bool:
		return &v
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return &b
		}
	case float64:
		b := v != 0
		return &b
	case int:
		b := v != 0
		return &b
	}
	return nil
}
