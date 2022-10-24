package actions

import "github.com/go-errors/errors"

func ErrorWithStack(err error) *errors.Error {
	return errors.New(err)
}
