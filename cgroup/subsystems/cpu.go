package subsystems

import (
	"io/ioutil"
	"fmt"
	"path"
)

type CPUSubSystem struct{}

func (m *CPUSubSystem) Name() string {
	return subCPU
}

// Set 限制内存使用
func (m *CPUSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	if res.CPUShare == "" {
		return nil
	}

	// 通过将值写入到 memory.limit_in_bytes 文件中来达到限制的效果
	if err := ioutil.WriteFile(
		path.Join(subPath, "memory.limit_in_bytes"),
		[]byte(res.MemoryLimit),
		0644,
	); err != nil {
		return fmt.Errorf("set cgroup cpu error: %v", err)
	}

	return nil
}

// Apply 将进程添加到 cgroup 中
func (m *CPUSubSystem) Apply(cgroupPath string, pid int64) error {
	return apply(m.Name(), cgroupPath, int(pid))
}

// Remove 删除 cgroup
func (m *CPUSubSystem) Remove(cgroupPath string) error {
	return remove(m.Name(), cgroupPath)
}
