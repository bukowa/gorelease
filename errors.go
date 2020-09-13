package gorelease

import "fmt"

type (
	ErrorEmptyFileName Target
)

func (err ErrorEmptyFileName) Error() string {
	return fmt.Sprintf("target '%#v' has empty file name", err)
}
