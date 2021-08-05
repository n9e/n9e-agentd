package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	validateCount int
)

type T1 struct {
	T2  T2
	T2s []T2
}

type T1p struct {
	T2  *T2
	T2s []*T2
}

type T2 struct{}

func (p *T2) Validate() error {
	validateCount++
	return nil
}

func TestValidateFields(t *testing.T) {
	validateCount = 0
	ValidateFields(&T1{T2s: make([]T2, 10)})
	assert.Equal(t, 11, validateCount)

	validateCount = 0
	ValidateFields(&T1p{T2: &T2{}, T2s: make([]*T2, 10)})
	assert.Equal(t, 1, validateCount)

	validateCount = 0
	ValidateFields(&T1p{T2s: []*T2{&T2{}, nil, &T2{}}})
	assert.Equal(t, 2, validateCount)
}
