package functions

import (
	"errors"
	"fmt"
	"strings"
)

var WrongParamsCount = errors.New("wrong params count")

func IsWrongParamsCount(err error) bool {
	return err == WrongParamsCount
}

func NewWrongParamType(index int) error {
	return fmt.Errorf("wrong param type at %d", index)
}

func IsWrongParamType(err error) bool {
	return strings.Contains(err.Error(), "wrong param type at")
}
