package command

import "bytes"

// Validater is the interface that wraps the Validate method
type Validater interface {
	Validate(actual, expected []byte) bool
}

// DiffValidater is verifies the exact match
type DiffValidater struct {
}

// Validate is verifies the exact match
func (d *DiffValidater) Validate(actual, expected []byte) bool {
	return bytes.Equal(actual, expected)
}

// Validaters is map of available validater
var Validaters = map[string]Validater{
	"diff": &DiffValidater{},
}
