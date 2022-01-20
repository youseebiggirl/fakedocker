package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// FindCgroupMountPoint 通过 /proc/self/mountinfo 找出挂载了某个 subsystem 的 hierarchy cgroup
// 根节点所在的目录
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// cat /proc/self/mountinfo 后从中任选一条作为示例：
		// 40 35 0:34 / /sys/fs/cgroup/cpu,cpuacct rw,nosuid,nodev,noexec,relatime shared:649 - cgroup cgroup rw,cpu,cpuacct
		// 按空格分割后：
		// [1] -> 40
		// [2] -> 35
		// [3] -> 0:34
		// [4] -> /
		// [5] -> /sys/fs/cgroup/cpu,cpuacct
		// [6] -> rw,nosuid,nodev,noexec,relatime
		// [7] -> shared:649
		// [8] -> -
		// [9] -> cgroup
		// [10] -> cgroup
		// [11] -> rw,cpu,cpuacct
		// 其中可以通过最后一条判断 subsystem 的类型，这里先通过 len(s)-1 获取到该字段，再通过 ","
		// 进行分割，将分割后的字段[rw cpu cpuacct] 与参数 subsystem 逐一进行比较，如果有任意一个匹
		// 配，则说明该条记录是我们需要的，同时观察发现，第[5]条记录就是该 subsystem 所在的位置，将该
		// 记录返回即可，在上面的例子中就是 /sys/fs/cgroup/cpu,cpuacct (文件名就是 cpu,cpuacct)
		txt := scanner.Text()
		s := strings.Split(txt, " ")

		for _, v := range strings.Split(s[len(s)-1], ",") {
			if v == subsystem {
				return s[4]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	_, err := os.Stat(path.Join(cgroupRoot, cgroupPath))
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
			return "", fmt.Errorf("create cgroup error: %v", err)
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	}
	return "", fmt.Errorf("cgroup path error: %v", err)
}