package subsystems

import (
	"fmt"
	"testing"
)

func TestFindCgroupMountPoint(t *testing.T) {
	s := FindCgroupMountPoint(subCPU)
	fmt.Printf("cpu: %v\n", s)
	// Output: 
	// cpu: /sys/fs/cgroup/cpu,cpuacct

	memory := FindCgroupMountPoint(subMem)
	fmt.Printf("memory: %v\n", memory)
	// Output: 
	// memory: /sys/fs/cgroup/memory
}

func TestGetCgroupPath(t *testing.T) {
	s, err := GetCgroupPath(subMem, "cgroup-test", true)
	if err != nil {
		t.Fatal(err)
	}
	
	t.Logf("path: %v\n", s)
}