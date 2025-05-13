package testtool

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

type AssertCase struct {
	A           any
	B           any
	ErrorExpect error
	ErrorActual error
	Description string
}

type AssertFunc func(t assert.TestingT, a any, b any, msgAndArgs ...any) bool

func (s AssertCase) Assert(t assert.TestingT, assertFunc AssertFunc) {
	assert.ErrorIs(t, s.ErrorActual, s.ErrorExpect, fmt.Sprintf("Error-%s", s.Description))
	if assertFunc != nil {
		assertFunc(t, s.A, s.B, s.Description)
	}
}
