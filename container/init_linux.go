package container

import (
	"fmt"
	"syscall"
	"os/exec"
	"strings"
	"os"
	"io"
	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

// initProcess 初始化容器进程，为容器进程挂载 /proc 目录
func InitProcess() error {
	// 阻塞等待，直到父进程向管道中写入内容
	cmds := readUserCommand()
	if len(cmds) == 0 {
		return fmt.Errorf("user command is nil")
	}

	defaultMountFlags :=
		syscall.MS_NOEXEC | // 在本文件系统中不允许运行其他程序
			syscall.MS_NOSUID | // 在本系统运行程序的时候，不允许 set-user-ID 或 set-group-ID
			syscall.MS_NODEV // 所有 mount 的系统都会默认设定的参数
		// 设置为私有挂载，否则容器进程挂载 proc 后主机的 proc 会失效，无法执行 ps 等命令
		// 后续：不能设置该 flag，否则会报错：invalid argument
		// 解决：ubuntu 根目录的挂载点的默认类型应为 MS_SHARED，执行 sudo mount --make-rprivate /
		// 将根目录挂载类型改为私有即可
		//syscall.MS_PRIVATE

	// 等同于命令 mount -t proc proc /proc
	// TODO: 第四个参数 [proc] 是什么意思？
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		return fmt.Errorf("mount proc error: %v", err)
	}

	// 从环境变量中搜索命令所在路径，比如传入的是 ls，返回 /bin/ls
	// 这样用户就不用输入全路径了
	p, err := exec.LookPath(cmds[0])
	if err != nil {
		return fmt.Errorf("exec loop path error: %v", err)
	}

	// 试试如果没有下面这些内容会怎么样
	// 执行完成后，没有进入到容器进程，echo $$ 输出依然为之前的 pid 而不是 1
	if err := syscall.Exec(p, cmds, os.Environ()); err != nil {
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
