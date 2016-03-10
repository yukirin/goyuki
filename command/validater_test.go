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

func TestFloatValidater(t *testing.T) {
	testCases := []struct {
		b1     []byte
		b2     []byte
		result bool
	}{
		{[]byte(""), []byte(""), true},
		{[]byte(""), []byte("1.11"), false},
		{[]byte("1.23 2.23\n3.23"), []byte("1.23 2.23\n3.23"), true},
		{[]byte("1e6"), []byte("1000000"), true},
	}

	validater := FloatValidater{}
	for _, testCase := range testCases {
		result := validater.Validate(testCase.b1, testCase.b2)
		if result != testCase.result {
			t.Errorf("Validate(%v, %v) = %v; want %v", testCase.b1, testCase.b2, result, testCase.result)
		}
	}
}

func TestRoundFloatValidater(t *testing.T) {
	testCases := []struct {
		b1     []byte
		b2     []byte
		result bool
	}{
		{[]byte(""), []byte(""), true},
		{[]byte(""), []byte("1.11"), false},
		{[]byte("1.23 2.23\n3.23"), []byte("1.23 2.23\n3.23"), true},
		{[]byte("1e6"), []byte("1000000"), true},
		{[]byte("1.23456"), []byte("1.23459"), true},
		{[]byte("1.23448"), []byte("1.23452222"), true},
		{[]byte("1.23457"), []byte("1.2345222"), false},
	}

	validater := FloatValidater{Place: 4}
	for _, testCase := range testCases {
		result := validater.Validate(testCase.b1, testCase.b2)
		if result != testCase.result {
			t.Errorf("Validate(%v, %v) = %v; want %v", testCase.b1, testCase.b2, result, testCase.result)
		}
	}
}
