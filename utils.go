package bhgrpcutils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
)

// NewError error
func NewError(c codes.Code, msg string) error {
	_, file, line, _ := runtime.Caller(1)

	errorLog := fmt.Sprintf("error file : %v ( code : %d) ( line : %d ) \n error info : %v", file, int(c), line, msg)
	logrus.Info(errorLog)

	return status.Error(c, msg)
}
