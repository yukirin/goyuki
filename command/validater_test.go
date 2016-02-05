package command

import "testing"

func TestDiffValidater(t *testing.T) {
	testCases := []struct {
		b1     []byte
		b2     []byte
		result bool
	}{
		{[]byte(""), []byte(""), true},
		{[]byte(""), []byte("foo\nbar"), false},
		{[]byte("foo\nbar"), []byte("foo\nbar\n"), true},
		{[]byte("foo\nbar\nfuga\n"), []byte("foo\nbar\nfuga\n"), true},
	}

	validater := DiffValidater{}
	for _, testCase := range testCases {
		result := validater.Validate(testCase.b1, testCase.b2)
		if result != testCase.result {
			t.Errorf("Validate(%v, %v) = %v; want %v", testCase.b1, testCase.b2, result, testCase.result)
		}
	}
}
