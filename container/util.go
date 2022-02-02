package container

import (
	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
	"os"
)

// PathIsExist 返回 p 是否存在，如果存在返回 true，否则返回 false
func PathIsExist(p string) (exist bool, err error) {
	_, err = os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateOrClear 如果 path 不存在则创建，如果 path 存在则清空
func CreateOrClear(path string) error {
	exist, err := PathIsExist(path)
	if err != nil {
		zlog.New().Panic("get path states error", zap.String("path", path), zap.Error(err))
		return err
	}

	if !exist {
		zlog.New().Info("path is not exist, ready to create it.", zap.String("path", path))
		if err := os.Mkdir(path, 0777); err != nil {
			zlog.New().Panic("mkdir path error", zap.String("path", path), zap.Error(err))
			return err
		}
	} else {
		zlog.New().Info("path is exist, ready to clear it.", zap.String("path", path))
		if err := os.RemoveAll(path); err != nil {
			zlog.New().Panic("clear path error", zap.String("path", path), zap.Error(err))
			return err
		}
		if err := os.Mkdir(path, 0777); err != nil {
			zlog.New().Panic("clear path error", zap.String("path", path), zap.Error(err))
			return err
		}
	}

	return nil
}

// CreateIfNotExist 如果 path 不存在则创建，存在则不作任何操作
func CreateIfNotExist(path string) error {
	exist, err := PathIsExist(path)
	if err != nil {
		zlog.New().Panic("get path states error", zap.String("path", path), zap.Error(err))
		return err
	}
	if !exist {
		zlog.New().Info("path is not exist, ready to create it.", zap.String("path", path))
		if err := os.Mkdir(path, 0777); err != nil {
			zlog.New().Panic("mkdir path error", zap.String("path", path), zap.Error(err))
			return err
		}
	}
	return nil
}
