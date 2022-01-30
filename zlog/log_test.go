package zlog

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ll, _ = zap.NewDevelopment(zap.AddStacktrace(zapcore.ErrorLevel))

func f() {
	if err := f1(); err != nil {
		return
	}
}

func f1() error {
	ll.Error("error")
	return fmt.Errorf("error")
}

func z() {
	if err := z1(); err != nil {
		ll.Error("error")
	}
}

func z1() error {
	return fmt.Errorf("error")
}

// 测试输出堆栈信息
func TestStackPrint(t *testing.T) {
	f()
	z()
}