package container

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

// InitProcess 初始化容器进程，为容器进程挂载 /proc 目录
func InitProcess() error {
	// 阻塞等待，直到父进程向管道中写入内容
	cmds := readUserCommand()
	if len(cmds) == 0 {
		zlog.New().Error("user command is nil")
		return fmt.Errorf("user command is nil")
	}

	if err := setUpMount(); err != nil {
		return err
	}

	// 从环境变量中搜索命令所在路径，比如传入的是 ls，返回 /bin/ls
	// 这样用户就不用输入全路径了
	p, err := exec.LookPath(cmds[0])
	if err != nil {
		zlog.New().Error("look path error", zap.Error(err))
		return err
	}

	// 试试如果没有下面这些内容会怎么样
	// 执行完成后，没有进入到容器进程，echo $$ 输出依然为之前的 pid 而不是 1
	if err := syscall.Exec(p, cmds, os.Environ()); err != nil {
		zlog.New().Error("exec error", zap.Error(err))
		return err
	}
	return nil
}

// readUserCommand 从管道中读取父进程传递的命令参数
func readUserCommand() []string {
	// NewFile 比较迷的一个函数，看注释也看不懂
	pipe := os.NewFile(uintptr(3), "pipe")
	b, err := io.ReadAll(pipe)
	if err != nil {
		zlog.New().Error("read command from pipe error: ", zap.Error(err))
		return nil
	}
	cmds := string(b)

	// 将多个命令参数用空格分开
	return strings.Split(cmds, " ")
}

// sendInitCommand 父进程发送用户命令到管道中
func sendInitCommand(cmds []string, wp *os.File) {
	cmd := strings.Join(cmds, " ")
	wp.WriteString(cmd)
	wp.Close()
}

func pivotRoot(rootPath string) error {
	// 为了使当前 root 的老 root 和新 root 不在同一个文件系统下，我们把 root
	// 重新 mount 一次，bind mount 是把相同的内容换了一个挂载点的挂载方法
	// FIXME 没搞懂啥意思，自己挂载自己？
	// 把这段话去掉似乎也没发现有什么影响？
	if err := syscall.Mount(rootPath, rootPath, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	pivotPath := filepath.Join(rootPath, ".pivot_root")

	// 判断当前目录是否已有该文件夹
	if _, err := os.Stat(pivotPath); err == nil {
		// 存在则删除
		if err := os.Remove(pivotPath); err != nil {
			errMsg := fmt.Sprintf("%v is exist, need remove it, but remove error", pivotPath)
			zlog.New().Error(errMsg, zap.Error(err))
			return err
		}
	}
	if err := os.Mkdir(pivotPath, 0777); err != nil {
		errMsg := fmt.Sprintf("mkdir %v error", pivotPath)
		zlog.New().Error(errMsg, zap.Error(err))
		//return err
	}

	// 将 rootPath 作为新的根目录，将之前的根目录暂时保存在 pivotPaht 中
	// 挂载点目前任然可以在 mount 命令中看到 （嘛意思？）
	// "error": "invalid argument" 
	// 解决：mount --make-rprivate /
	if err := syscall.PivotRoot(rootPath, pivotPath); err != nil {
		errMsg := fmt.Sprintf("pivotRoot %v to %v error", rootPath, pivotPath)
		zlog.New().Error(errMsg, zap.Error(err))
		return err
	}

	// chdir() 用来将当前的工作目录改变成以参数 path 所指的目录.
	if err := syscall.Chdir("/"); err != nil {
		zlog.New().Error("chdir to / error", zap.Error(err))
		return err
	}

	pivotPath = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotPath, syscall.MNT_DETACH); err != nil {
		errMsg := fmt.Sprintf("unmount %v error", pivotPath)
		zlog.New().Error(errMsg, zap.Error(err))
		return err
	}

	// 删除临时文件夹
	// FIXME: error: device or resource busy
	// 暂时无法解决，但是不影响运行
	if err := os.Remove(pivotPath); err != nil {
		errMsg := fmt.Sprintf("remove path %v error", pivotPath)
		zlog.New().Error(errMsg, zap.Error(err))
		//return err
	}
	return nil
}

// setUpMount 初始化容器的挂载点
func setUpMount() error {
	// 首先设置根目录为私有模式，防止影响 pivot_root
	cmd := exec.Command("mount", "--make-rprivate", "/")
	_, err := cmd.CombinedOutput()
	if err != nil {
		zlog.New().Error("set / mount private error", zap.Error(err))
		return err
	}

	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		zlog.New().Error("getwd error", zap.Error(err))
		return err
	}
	zlog.New().Info("current localtion is", zap.String("path", pwd))

	if err := pivotRoot(pwd); err != nil {
		return err
	}

	defaultMountFlags := syscall.MS_NOEXEC | // 在本文件系统中不允许运行其他程序
		syscall.MS_NOSUID | // 在本系统运行程序的时候，不允许 set-user-ID 或 set-group-ID
		syscall.MS_NODEV // 所有 mount 的系统都会默认设定的参数

	// 设置为私有挂载，否则容器进程挂载 proc 后主机的 proc 会失效，无法执行 ps 等命令
	// 后续：不能设置该 syscall.MS_PRIVATE，否则会报错：invalid argument
	// 解决：ubuntu 根目录的挂载点的默认类型应为 MS_SHARED，执行 sudo mount --make-rprivate /
	// 将根目录挂载类型改为私有即可
	// 或者使用下面的方法：


	//if err := syscall.Mount("/", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
	//	zlog.New().Error("set / mount private error", zap.Error(err))
	//	return err
	//}

	// 等同于命令 mount -t proc proc /proc
	// TODO: 第四个参数 [proc] 是什么意思？
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		zlog.New().Error("mount proc error", zap.Error(err))
		return fmt.Errorf("mount proc error: %v", err)
	}

	// 使用 df -hl 查看，发现有一个 tmpfs 也在挂载
	// tmpfs是 Linux/Unix 系统上的一种基于内存的文件系统。
	// tmpfs 可以使用您的内存或 swap 分区来存储文件。由此可见，
	// temfs 主要存储暂存的文件。
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		zlog.New().Error("mount tmpfs error", zap.Error(err))
		return fmt.Errorf("mount tmpfs error: %v", err)
	}

	return nil	
}
