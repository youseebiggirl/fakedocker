package zlog

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var l *zap.Logger

func init() {
	// 输出等级为 error 的日志堆栈
	z, err := zap.NewDevelopment(
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic(fmt.Sprintf("init log error: %v", err))
	}
	l = z
}

func New() *zap.Logger {
	return l
}
