package steamerrors

import (
	"errors"
	"strconv"
)

var (
	ErrNotSupported   = errors.New("Steamworks API is not supported for this platform.")
	ErrNotInitialized = errors.New("Init function must be called first.")
)

type dllError struct {
	methodName string
	err        error
}

func NewDLLError(methodName string, err error) *dllError {
	return &dllError{
		methodName: methodName,
		err:        err,
	}
}

func (err *dllError) Error() string {
	return err.methodName + ": " + err.err.Error()
}

type dllBadReturnCodeError struct {
	methodName string
	returnCode int64
}

func NewDLLBadReturnCodeError(methodName string, returnCode uintptr) *dllBadReturnCodeError {
	return &dllBadReturnCodeError{
		methodName: methodName,
		returnCode: int64(returnCode),
	}
}

func (err *dllBadReturnCodeError) Error() string {
	return err.methodName + ": Bad return code " + strconv.FormatInt(err.returnCode, 10)
}
