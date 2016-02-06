package command

import (
	"bufio"
	"bytes"
	"strconv"
)

// Validater is the interface that wraps the Validate method
type Validater interface {
	Validate(actual, expected []byte) bool
}

// DiffValidater is verifies the exact match
type DiffValidater struct {
}

// Validate is verifies the exact match
func (d *DiffValidater) Validate(actual, expected []byte) bool {
	asc := bufio.NewScanner(bytes.NewReader(actual))
	esc := bufio.NewScanner(bytes.NewReader(expected))

	next1, next2 := asc.Scan(), esc.Scan()
	for ; next1 && next2; next1, next2 = asc.Scan(), esc.Scan() {
		if asc.Text() != esc.Text() {
			return false
		}
	}

	if asc.Err() == nil && esc.Err() == nil && next1 == next2 {
		return true
	}
	return false
}

// FloatValidater compares converted to float
type FloatValidater struct {
}

// Validate compares converted to float
func (f *FloatValidater) Validate(actual, expected []byte) bool {
	asc := bufio.NewScanner(bytes.NewReader(actual))
	asc.Split(bufio.ScanWords)
	esc := bufio.NewScanner(bytes.NewReader(expected))
	esc.Split(bufio.ScanWords)

	next1, next2 := asc.Scan(), esc.Scan()
	for ; next1 && next2; next1, next2 = asc.Scan(), esc.Scan() {
		f1, err := strconv.ParseFloat(asc.Text(), 64)
		if err != nil {
			return false
		}

		f2, err := strconv.ParseFloat(esc.Text(), 64)
		if err != nil {
			return false
		}

		if f1 != f2 {
			return false
		}
	}

	if asc.Err() == nil && esc.Err() == nil && next1 == next2 {
		return true
	}
	return false
}

// Validaters is map of available validater
var Validaters = map[string]Validater{
	"diff": &DiffValidater{},
}
