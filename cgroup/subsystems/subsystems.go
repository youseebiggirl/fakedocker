package subsystems

import (

)

// ResourceConfig 用于记录资源限制配置
type ResourceConfig struct {
	MemoryLimit string // 内存限制
	CPUShare    string // CPU 时间片权重
	CPUSet      string // CPU 核心数
}

type Subsystem interface {
	// 返回 subsystem 的名字，比如 cpu memory
	Name() string

	// 设置某个 cgroup 在这个 subsystem 中的资源限制
	Set(path string, res *ResourceConfig) error

	// 将进程添加到某个 cgroup 中
	Apply(path string, pid int64) error

	// 移除某个 cgroup
	Remove(path string) error
}

var (
	_ Subsystem = &MemorySubSystem{}
)

type MemorySubSystem struct{}

func (m *MemorySubSystem) Name() string {
	return "memory"
}

func (m *MemorySubSystem) Set(path string, res *ResourceConfig) error {
	return nil
}

func (m *MemorySubSystem) Apply(path string, pid int64) error {
	return nil
}

func (m *MemorySubSystem) Remove(path string) error {
	return nil
}
