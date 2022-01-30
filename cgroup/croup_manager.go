package cgroup

import (
	"github.com/YOUSEEBIGGIRL/fakedocke/cgroup/subsystems"
	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

var allSubSys = []subsystems.Interface{
	&subsystems.CPUSubSystem{},
	&subsystems.MemorySubSystem{},
}

type CgroupManager struct {
	// 相对于 cgroup 根目录的路径，比如 /sys/fs/cgroup/memory 就是一个
	// 根目录，path 就是在该路径下的一个路径，比如 /sys/fs/cgroup/memory/test1
	// path 就是 test1
	Path           string
	ResourceConfig *subsystems.ResourceConfig
}

func NewCgroupManager(path string, resConf *subsystems.ResourceConfig) *CgroupManager {
	return &CgroupManager{
		Path: path,
		ResourceConfig: resConf,
	}
}

// SetAll 根据 ResourceConfig 设置各个 subsystem 挂载中的 cgroup 资源限制
func (m *CgroupManager) SetAll() (err error) {
	for _, v := range allSubSys {
		err = v.Set(m.Path, m.ResourceConfig)
		if err != nil {
			zlog.New().Error(
				"set cgroup error",
				zap.String("subsystem name", v.Name()),
				zap.String("subsystem path", m.Path),
				zap.Error(err),
			)
			break
		}
	}
	return
}

// ApplyAll 将进程 pid 加入到每个 cgroup 中
func (m *CgroupManager) ApplyAll(pid int64) (err error) {
	for _, v := range allSubSys {
		err = v.Apply(m.Path, pid)
		if err != nil {
			zlog.New().Error(
				"apply process to cgroup error",
				zap.String("subsystem name", v.Name()),
				zap.Int64("pid", pid),
				zap.String("subsystem path", m.Path),
				zap.Error(err),
			)
			break
		}
	}
	return
}

// RemoveAll 释放各个 subsystem 挂载中的 cgroup
func (m *CgroupManager) RemoveAll() (err error) {
	for _, v := range allSubSys {
		err = v.Remove(m.Path)
		if err != nil {
			zlog.New().Error(
				"remove cgroup error",
				zap.String("subsystem name", v.Name()),
				zap.String("subsystem path", m.Path),
				zap.Error(err),
			)
			return
		}
	}
	return
}