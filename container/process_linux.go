package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/YOUSEEBIGGIRL/fakedocke/cgroup"
	"github.com/YOUSEEBIGGIRL/fakedocke/cgroup/subsystems"
	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"go.uber.org/zap"
)

var (
	mntPath  = "/root/mnt"
	rootPath = "/root/"
)

// NewPipe 创建一个匿名管道用于父子进程间通信
func NewPipe() (rPipe, wPipe *os.File, err error) {
	rPipe, wPipe, err = os.Pipe()
	if err != nil {
		zlog.New().Error("create pipe error: ", zap.Error(err))
		return nil, nil, err
	}
	return
}

// NewParentProcess 创建一个隔离的容器进程，但并不运行，同时创建一个管道用于
// 进程通信，返回管道的写端，子进程拥有管道的读端，父进程通过写端向管道写入用户
// 传入的参数，子进程通过读端来获取参数
// tty 表示是否开启一个伪终端
//（疑问：是不是叫 NewChildProcess 更合适？）
func NewParentProcess(tty bool, volume string) (cmd *exec.Cmd, wp *os.File) {
	// 自己调用自己，同时调用 init 命令（init 会调用 InitProcess）进行初始化（挂载 /proc）
	// cmd 可以理解为一个子进程，但是还没有启动，后续调用 Run 或 Start 启动
	cmd = exec.Command("/proc/self/exe", "init")
	zlog.New().Info("exec this command: ", zap.String("cmd: ", cmd.String()))
	//log.Printf("command: %v\n", cmd.String())

	rp, wp, err := NewPipe()
	if err != nil {
		//zlog.New().Error("create pipe error: ", zap.Error(err))
		return nil, nil
	}

	// ExtraFiles 用于给新进程继承父进程中打开的文件
	// 这里将管道的读端 readPipe 传递给子进程，这样子进程就可以发送数据到管道中了
	// 一个进程默认有 3 个文件描述符，stdin，stdout 和 stderr
	cmd.ExtraFiles = []*os.File{rp}

	// 将只读层和可写层挂载到 mntPath
	NewWorkSpace(rootPath, mntPath, volume)
	// 给创建出来的子进程指定容器初始化后的工作目录
	cmd.Dir = mntPath

	// fork 一个新进程，并且使用 namespace 对资源进行了隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	// 是否开启伪终端
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd, wp
}

// RunProcess 运行容器进程
func RunProcess(
	tty bool,
	cmds []string,
	volume string,
	resConf *subsystems.ResourceConfig,
) {
	zlog.New().Info(
		"run process",
		zap.Strings("all command", cmds),
		zap.Bool("tty open status: ", tty),
	)

	zlog.New().Info(
		"resource config",
		zap.String("memory limit", resConf.MemoryLimit),
		zap.String("cpushare limit", resConf.CPUShare),
		zap.String("cpuset limit", resConf.CPUSet),
	)

	// 因为 NewParentProcess 里面会调用 NewWorkSpace 进行挂载，所以必须在程序结束时
	// 执行 DeleteWorkSpace 取消挂载，不然会有一些文件任然处于挂载状态，产生一些错误，
	// 为了达到目的，使用 defer 进行注册
	defer func() {
		// 容器执行完成后，把容器对应的 write layer 删除
		if err := DeleteWorkSpace(rootPath, mntPath, volume); err != nil {
			os.Exit(-1)
		}
	}()


	p, wp := NewParentProcess(tty, volume)
	// 运行并等待 cmd 执行完成
	if err := p.Start(); err != nil {
		zlog.New().Error("run process error", zap.Error(err))
		os.Exit(-1)
	}

	sendInitCommand(cmds, wp)

	cg := cgroup.NewCgroupManager("cgroup-test", resConf)
	defer func() {
		if err := cg.RemoveAll(); err != nil {
			os.Exit(-1)
		}
	}()

	// 忽略错误判断，函数内部会打印日志，如果遇到错误直接 return，那么后面的 DeleteWorkSpace
	// 会无法执行，会产生各种诡异的问题
	if err := cg.SetAll(); err != nil {
		os.Exit(-1)
	}
	if err := cg.ApplyAll(int64(p.Process.Pid)); err != nil {
		os.Exit(-1)
	}
	if err := p.Wait(); err != nil {
		os.Exit(-1)
	}
	// 执行完成后，退出进程
	os.Exit(-1)
}
