package aptible

import (
	"fmt"
	"math"
	"reflect"
)

// makes a string slice out of a slice of type interface
func makeStringSlice(interfaceSlice []interface{}) ([]string, error) {
	strSlice := make([]string, len(interfaceSlice))
	for i := 0; i < len(interfaceSlice); i++ {
		if (reflect.TypeOf(interfaceSlice[i]).Kind()) != reflect.String {
			return []string{}, fmt.Errorf("slice contains non-string elements")
		}
		strSlice[i] = interfaceSlice[i].(string)
	}
	return strSlice, nil
}

// makes a int32 slice out of a slice of type interface
func makeInt32Slice(interfaceSlice []interface{}) ([]int32, error) {
	int32Slice := make([]int32, len(interfaceSlice))

	for i := 0; i < len(interfaceSlice); i++ {
		kind := reflect.TypeOf(interfaceSlice[i]).Kind()

		if kind == reflect.Int32 {
			int32Slice[i] = interfaceSlice[i].(int32)
		} else if kind == reflect.Int {
			num := interfaceSlice[i].(int)
			if fitsInt32(num) {
				int32Slice[i] = int32(num)
			} else {
				return []int32{}, fmt.Errorf("slice contains int elements larger than 32 bits: %d", num)
			}
		} else if kind == reflect.Int64 {
			num := interfaceSlice[i].(int64)
			if fitsInt32(num) {
				int32Slice[i] = int32(num)
			} else {
				return []int32{}, fmt.Errorf("slice contains int elements larger than 32 bits: %d", num)
			}
		} else {
			return []int32{}, fmt.Errorf("slice contains non-int elements")
		}
	}
	return int32Slice, nil
}

func fitsInt32[T ~int | ~int64](val T) bool {
	return val >= math.MinInt32 && val <= math.MaxInt32
}
