package subsystems

import (
	"io/ioutil"
	"fmt"
	"strconv"
	"path"
	"os"
)

const (
	subMem = "memory"
	subCPU = "cpu"
)

// ResourceConfig 用于记录资源限制配置
type ResourceConfig struct {
	MemoryLimit string // 内存限制
	CPUShare    string // CPU 时间片权重
	CPUSet      string // CPU 核心数
}

type Interface interface {
	// 返回 subsystem 的名字，比如 cpu memory
	Name() string

	// 设置某个 cgroup 在这个 subsystem 中的资源限制
	Set(cgroupPath string, res *ResourceConfig) error

	// 将进程添加到某个 cgroup 中
	Apply(cgroupPath string, pid int64) error

	// 移除某个 cgroup
	Remove(cgroupPath string) error
}

var (
	_ Interface = &MemorySubSystem{}
	_ Interface = &CPUSubSystem{}
)

// apply 和 remove 具有通用性，可以复用代码
// 只需独立实现 set() 的逻辑即可

func apply(subSysName, cgroupPath string, pid int) error {
	subPath, err := GetCgroupPath(subSysName, cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}

	if err := ioutil.WriteFile(
		path.Join(subPath, "tasks"),
		[]byte(strconv.Itoa(pid)),
		0644,
	); err != nil {
		return fmt.Errorf("set cgroup proc error: %v", err)
	}

	return nil
}

func remove(subSysName, cgroupPath string) error {
	subPath, err := GetCgroupPath(subSysName, cgroupPath, false)
	if err != nil {
		return fmt.Errorf("remove cgroup %s error: %v", cgroupPath, err)
	}

	if err := os.Remove(subPath); err != nil {
		return err
	}

	return nil
}

