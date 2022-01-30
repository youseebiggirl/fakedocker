package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

// 通过 aufs 实现镜像和容器的目录分离

func NewWorkSpace(rootPath, mntPath string) {
	CreateReadOnlyLayer(rootPath)
	CreateWriteLayer(rootPath)
	CreateMountPoint(rootPath, mntPath)
}

// CreateReadOnlyLayer 将 rootPath/busybox.tar 解压到 rootPath/busybox 目录下，
// 作为容器的只读层，需要确保 rootPath/busybox.tar 存在
func CreateReadOnlyLayer(rootPath string) {
	busyboxPath := filepath.Join(rootPath, "/busybox")
	busyboxTarPath := filepath.Join(rootPath, "busybox.tar")

	exist, err := pathIsExist(busyboxPath)
	if err != nil {
		zlog.New().Error(
			"get busybox path stat error",
			zap.String("path", busyboxPath),
			zap.Error(err),
		)
		return
	}

	if !exist {
		if err := os.Mkdir(busyboxPath, 0777); err != nil {
			zlog.New().Error(
				"mkdir busybox dir error",
				zap.String("mkdir path", busyboxPath),
				zap.Error(err),
			)
			return
		}
	}

	cmd := exec.Command("tar", "-xf", busyboxTarPath, "-C", busyboxPath)
	_, err = cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"tar busybox.tar error",
			zap.String("source path", busyboxTarPath),
			zap.String("target path", busyboxPath),
			zap.String("command", cmd.String()),
			zap.Error(err),
		)
		return
	}
}

// CreateWriteLayer 创建一个名为 write_layer 的文件夹作为容器唯一的可写层
// 莫名其妙的 bug，在 test 中单独调用该函数可以创建文件夹，但是在 NewWorkSpace 中
// 调用却无法创建（原因：容器执行完成后会删除该文件夹）
func CreateWriteLayer(rootPath string) {
	writePath := filepath.Join(rootPath, "write_layer")
	exist, err := pathIsExist(writePath)
	if err != nil {
		zlog.New().Error(
			"get path stat error",
			zap.String("path", rootPath),
			zap.Error(err),
		)
		return
	}
	//zlog.New().Info("exist", zap.Bool("", exist), zap.String("path", rootPath))

	if exist {
		return
	}
	
	//zlog.New().Info("writePath", zap.String("path", writePath))
	if err := os.Mkdir(writePath, 0777); err != nil {
		zlog.New().Error(
			"mkdir write layer error",
			zap.String("path", writePath),
			zap.Error(err),
		)
		return
	}
}

// CreateMountPoint 创建 mntPath 并作为挂载点，
// 把 write_layer（可写层） 和 busybox（只读层） 挂载到 mntPath（挂载点）
func CreateMountPoint(rootPath, mntPath string) {
	exist, err := pathIsExist(mntPath)
	if err != nil {
		zlog.New().Error(
			"get path stat error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return
	}

	if !exist {
		// 创建 mnt 文件夹作为挂载点
		if err := os.Mkdir(mntPath, 0777); err != nil {
			zlog.New().Error(
				"mkdir mnt dir error",
				zap.String("path", mntPath),
				zap.Error(err),
			)
			return
		}
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
	_, err = cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"mount write_layer and busybox to mnt error",
			zap.String("write_layer path", writeLayerPath),
			zap.String("busybox path", busyboxPath),
			zap.String("mnt path", mntPath),
			zap.String("cmd", cmd.String()),
			zap.Error(err),
		)
		return
	}
}

// pathIsExist 返回 p 是否存在，如果存在返回 true，否则返回 false
func pathIsExist(p string) (exist bool, err error) {
	_, err = os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DeleteWorkSpace(rootPath, mntPath string) {
	DeleteMountPoint(mntPath)
	DeleteWriteLayer(rootPath)
}

func DeleteMountPoint(mntPath string) {
	cmd := exec.Command("umount", mntPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error(
			"unmount mnt error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return
	}
	if err := os.RemoveAll(mntPath); err != nil {
		zlog.New().Error(
			"rm -rf mnt error",
			zap.String("path", mntPath),
			zap.Error(err),
		)
		return
	}
}

func DeleteWriteLayer(rootPath string) {
	p := filepath.Join(rootPath, "write_layer")
	if err := os.RemoveAll(p); err != nil {
		zlog.New().Error(
			"rm -rf write_layer error",
			zap.String("path", p),
			zap.Error(err),
		)
		return
	}
}
