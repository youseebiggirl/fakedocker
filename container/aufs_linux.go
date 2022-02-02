package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

// 通过 aufs 实现镜像和容器的目录分离

// NewWorkSpace 使用 AUFS 创建文件系统
func NewWorkSpace(rootPath, mntPath, volume string) error {
	// 1. 创建只读层（busybox）
	if err := CreateReadOnlyLayer(rootPath); err != nil {
		return err
	}
	// 2. 创建容器读写层（writeLayer）
	if err := CreateWriteLayer(rootPath); err != nil {
		return err
	}
	// 3. 创建挂载点（mnt），并把只读层和读写层挂载到挂载点
	// 4. 将挂载点作为容器的根目录
	if err := CreateMountPoint(rootPath, mntPath); err != nil {
		return err
	}
	// 如果用户指定了 -v
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := MountVolume(mntPath, volumeURLs); err != nil {
				return err
			}
		} else {
			zlog.New().Error(
				"-v input is not correct, usage /host:/container",
				zap.String("-v input", volume),
			)
		}
	}
	return nil
}

// CreateReadOnlyLayer 将 rootPath/busybox.tar 解压到 rootPath/busybox 目录下，
// 作为容器的只读层，需要确保 rootPath/busybox.tar 存在
func CreateReadOnlyLayer(rootPath string) error {
	busyboxPath := filepath.Join(rootPath, "/busybox")
	busyboxTarPath := filepath.Join(rootPath, "busybox.tar")

	if err := CreateOrClear(busyboxPath); err != nil {
		return err
	}

	cmd := exec.Command("tar", "-xf", busyboxTarPath, "-C", busyboxPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"tar busybox.tar error",
			zap.String("source path", busyboxTarPath),
			zap.String("target path", busyboxPath),
			zap.String("command", cmd.String()),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// CreateWriteLayer 创建一个名为 write_layer 的文件夹作为容器唯一的可写层
// 莫名其妙的 bug，在 test 中单独调用该函数可以创建文件夹，但是在 NewWorkSpace 中
// 调用却无法创建（原因：容器执行完成后会删除该文件夹）
func CreateWriteLayer(rootPath string) error {
	writePath := filepath.Join(rootPath, "write_layer")
	if err := CreateOrClear(writePath); err != nil {
		return err
	}
	return nil
}

// CreateMountPoint 创建 mntPath 并作为挂载点，
// 把 write_layer（可写层） 和 busybox（只读层） 挂载到 mntPath（挂载点）
func CreateMountPoint(rootPath, mntPath string) error {
	if err := CreateOrClear(mntPath); err != nil {
		return err
	}

	writeLayerPath := filepath.Join(rootPath, "write_layer")
	busyboxPath := filepath.Join(rootPath, "busybox")

	// 挂载 aufs 命令示例：
	// 把 a 和 b 挂载到 mnt
	// mount -t aufs -o dirs=./a:./b none ./mnt
	dirs := fmt.Sprintf("dirs=%v:%v", writeLayerPath, busyboxPath)

	// 把 write_layer 和 busybox 挂载到 mnt
	// 对于挂载的权限信息，默认行为是：dirs 指定的左边起第一个目录是 read-write 权限，
	// 后续目录都是 read-only 权限
	// 所以下面需要把 writeLayerPath 作为左起第一个
	// 之后容器会将挂载点 mnt 作为自己的根目录，并 chdir 为 /，此时向容器中（即 mnt）写入的
	// 内容，都会拷贝到可写层 writeLayerPath 中（/root/write_layer），可以在容器运行
	// 过程中，新开一个终端，ls 主机的 /root/write_layer 目录进行验证
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"mount write_layer and busybox to mnt error",
			zap.String("write_layer path", writeLayerPath),
			zap.String("busybox path", busyboxPath),
			zap.String("mnt path", mntPath),
			zap.String("cmd", cmd.String()),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// DeleteWorkSpace 在容器退出时删除 AUFS 对应文件，并 unmount，
// 如果不 unmount 则无法 rm，会报错 Device or resource busy
func DeleteWorkSpace(rootPath, mntPath, volume string) error {
	zlog.New().Info("start delete workspace")
	// 只有在 volume 不为空，且解析后长度为 2，且都不为空时，
	// 才调用 DeleteMountPointWithVolume
	// 其他情况依然调用 DeleteMountPoint
	if volume != "" {
		v := volumeUrlExtract(volume)
		l := len(v)
		if l == 2 && v[0] != "" && v[1] != "" {
			if err := DeleteMountPointWithVolume(mntPath, v); err != nil {
				return err
			}
		} else {
			if err := DeleteMountPoint(mntPath); err != nil {
				return err
			}
		}
	} else {
		// 卸载挂载点（mnt）的文件系统，
		// 删除挂载点
		if err := DeleteMountPoint(mntPath); err != nil {
			return err
		}
	}
	// 删除读写层（writeLayer）
	if err := DeleteWriteLayer(rootPath); err != nil {
		return err
	}
	return nil
}

// DeleteMountPoint 卸载并删除挂载点 mnt
func DeleteMountPoint(mntPath string) error {
	// 卸载
	cmd := exec.Command("umount", mntPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"unmount mnt error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return err
	}
	// 删除
	if err := os.RemoveAll(mntPath); err != nil {
		zlog.New().Error(
			"rm -rf mnt error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// DeleteMountPointWithVolume 卸载并删除 mnt 和容器内的 volume
func DeleteMountPointWithVolume(mntPath string, volumePath []string) error {
	if len(volumePath) < 2 {
		zlog.New().Error("volumePath error, len < 2", zap.Strings("volume path", volumePath))
		return fmt.Errorf("volumePath error, len < 2")
	}

	p := filepath.Join(mntPath, volumePath[1])
	cmd := exec.Command("umount", p)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"unmount volume path error",
			zap.String("path", p),
			zap.Error(err),
		)
		return err
	}

	cmd1 := exec.Command("umount", mntPath)
	_, err = cmd1.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"unmount mnt error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return err
	}

	if err := os.RemoveAll(mntPath); err != nil {
		zlog.New().Error(
			"rm -rf mnt error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func DeleteWriteLayer(rootPath string) error {
	p := filepath.Join(rootPath, "write_layer")
	if err := os.RemoveAll(p); err != nil {
		zlog.New().Error(
			"rm -rf write_layer error",
			zap.String("path", p),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// MountVolume 根据传入的 volumes 将宿主机目录挂载到容器的 mnt 目录下，且在容器退出后，
// 数据卷中的内容仍然能够保存在宿主机中
func MountVolume(mntPath string, volumes []string) error {
	hostPath := volumes[0] // 宿主机目录
	if err := CreateIfNotExist(hostPath); err != nil {
		return err
	}

	// 在容器文件系统中创建挂载点，因为容器会将 mntPath 作为启动目录，所以需要拼接
	// mntPath 和 containerPath 作为最终路径
	containerPath := volumes[1]
	containerVolumePath := filepath.Join(mntPath, containerPath)
	if err := CreateOrClear(containerVolumePath); err != nil {
		return err
	}

	dirs := "dirs=" + hostPath
	// localPath 挂载到 containerVolumePath，此时 localPath 是读写层
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumePath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"mount volume error",
			zap.String("host path", hostPath),
			zap.String("container volume path", containerVolumePath),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// 解析 volume url，传入示例：/root/a:/b，表示将宿主机的 /root/a 挂载到容器的 /b
func volumeUrlExtract(volumeUrl string) []string {
	return strings.Split(volumeUrl, ":")
}
