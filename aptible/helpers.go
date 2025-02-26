package aptible

import (
	"fmt"
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

// makes a in32 slice out of a slice of type interface
func makeInt32Slice(interfaceSlice []interface{}) ([]int32, error) {
	int32Slice := make([]int32, len(interfaceSlice))

	for i := 0; i < len(interfaceSlice); i++ {
		kind := reflect.TypeOf(interfaceSlice[i]).Kind()

		if kind == reflect.Int32 {
			int32Slice[i] = interfaceSlice[i].(int32)
		} else if kind == reflect.Int {
			int32Slice[i] = int32(interfaceSlice[i].(int))
		} else if kind == reflect.Int64 {
			int32Slice[i] = int32(interfaceSlice[i].(int64))
		} else {
			return []int32{}, fmt.Errorf("slice contains non-int elements")
		}
	}
	return int32Slice, nil
}
