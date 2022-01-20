package container

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个隔离的容器进程
func NewParentProcess(tty bool, command string) *exec.Cmd {
	// 调用 init 命令进行初始化（挂载 /proc）
	// 同时调用用户传入的 command
	args := []string{"init", command}
	//log.Println("args: ", args)
	// 自己调用自己，同时指定了上面的 args，达到初始化的效果
	cmd := exec.Command("/proc/self/exe", args...)
	log.Printf("command: %v\n", cmd.String())
	// fork 一个新进程，并且使用 namespace 对资源进行了隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

// RunProcess 运行容器进程
func RunProcess(tty bool, cmd string) {
	log.Printf("run process, command: %v, tty: %v\n", cmd, tty)
	c := NewParentProcess(tty, cmd)
	// 运行并等待 cmd 执行完成
	// 不要使用 Start()，该函数不会等待执行完成（除非之后再调用 Wait()，其实 Run 内部就是这么做的）
	// 否则进程会无限调用自己
	c.Run()

	// 调用 Run() 程序不能正常运行，不知道为什么

	// if err := c.Start(); err != nil {
	// 	log.Println(err)
	// }
	// c.Wait()
	// 执行完成后，退出进程
	os.Exit(-1)
}

// initProcess 初始化容器进程，为容器进程挂载 /proc 目录
func InitProcess(cmd string) error {
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
		return err
	}
	argv := []string{cmd}
	if err := syscall.Exec(cmd, argv, os.Environ()); err != nil {
		return err
	}
	return nil
}
