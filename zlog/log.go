package zlog

import (
	"fmt"

	"go.uber.org/zap"
)

var l *zap.Logger

func init() {
	z, err := zap.NewDevelopment() 
	if err != nil {
		panic(fmt.Sprintf("init log error: ", err))
	}
	l = z
}

func New() *zap.Logger {
	return l
}





