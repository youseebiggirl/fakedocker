package subsystems

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

type MemorySubSystem struct{}

func (m *MemorySubSystem) Name() string {
	return subMem
}

// Set 限制内存使用
func (m *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	if res.MemoryLimit == "" {
		return nil
	}

	zlog.New().Info(
		"write limit to file...", 
		zap.String("path", path.Join(subPath, "memory.limit_in_bytes")),
	)

	// 通过将值写入到 memory.limit_in_bytes 文件中来达到限制的效果
	if err := ioutil.WriteFile(
		path.Join(subPath, "memory.limit_in_bytes"),
		[]byte(res.MemoryLimit),
		0644,
	); err != nil {
		return fmt.Errorf("set cgroup memory error: %v", err)
	}

	return nil
}

// Apply 将进程添加到 cgroup 中
func (m *MemorySubSystem) Apply(cgroupPath string, pid int64) error {
	return apply(m.Name(), cgroupPath, int(pid))
}

// Remove 删除 cgroup
func (m *MemorySubSystem) Remove(cgroupPath string) error {
	return remove(m.Name(), cgroupPath)
}
