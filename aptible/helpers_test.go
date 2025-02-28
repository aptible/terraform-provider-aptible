package aptible

import "testing"

func TestMakeStringSlice(t *testing.T) {
	var tests = []struct {
		name           string
		interfaceSlice []interface{}
		expected       []string
		errored        bool
	}{
		{"test_vanilla",
			[]interface{}{"chocolate", "vanilla", "strawberry"},
			[]string{"chocolate", "vanilla", "strawberry"},
			false},
		{"test_invalid_slice",
			[]interface{}{"chocolate", 24, ""},
			[]string{},
			true},
		// Add other test cases if we need in the future
	}

	for _, tc := range tests {
		testName := tc.name
		t.Run(testName, func(t *testing.T) {

			slice, err := makeStringSlice(tc.interfaceSlice)

			if (err != nil) != tc.errored {
				t.Errorf("Input: %v caused error: %v", tc.interfaceSlice, err)
			}

			if !isEqual(slice, tc.expected) {
				t.Errorf("Input: %v should have resulted in slice = %v. It was %v instead.", tc.interfaceSlice, tc.expected, slice)
			}
		})
	}
}

func TestMakeInt32Slice(t *testing.T) {
	var tests = []struct {
		name           string
		interfaceSlice []interface{}
		expected       []int32
		errored        bool
	}{
		{"test_vanilla",
			[]interface{}{13, 27, 2},
			[]int32{13, 27, 2},
			false},
		{"test_invalid_slice",
			[]interface{}{24, "chocolate", 52},
			[]int32{},
			true},
		// Add other test cases if we need in the future
	}

	for _, tc := range tests {
		testName := tc.name
		t.Run(testName, func(t *testing.T) {

			slice, err := makeInt32Slice(tc.interfaceSlice)

			if (err != nil) != tc.errored {
				t.Errorf("Input: %v caused error: %v", tc.interfaceSlice, err)
			}

			if !isEqual(slice, tc.expected) {
				t.Errorf("Input: %v should have resulted in slice = %v. It was %v instead.", tc.interfaceSlice, tc.expected, slice)
			}
		})
	}
}

// Tests if two slices are equal
func isEqual[T comparable](a, b []T) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
