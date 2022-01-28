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

// NewPipe 创建一个匿名管道用于父子进程间通信
func NewPipe() (rPipe, wPipe *os.File, err error) {
	rPipe, wPipe, err = os.Pipe()
	if err != nil {
		//zlog.New().Error("create pipe error: ", zap.Error(err))
		return nil, nil, err
	}
	return
}

// NewParentProcess 创建一个隔离的容器进程，但并不运行，同时创建一个管道用于
// 进程通信，返回管道的写端，子进程拥有管道的读端，父进程通过写端向管道写入用户
// 传入的参数，子进程通过读端来获取参数
// tty 表示是否开启一个伪终端
//（疑问：是不是叫 NewChildProcess 更合适？）
func NewParentProcess(tty bool) (cmd *exec.Cmd, wp *os.File) {
	// 自己调用自己，同时调用 init 命令（init 会调用 InitProcess）进行初始化（挂载 /proc）
	// cmd 可以理解为一个子进程，但是还没有启动，后续调用 Run 或 Start 启动
	cmd = exec.Command("/proc/self/exe", "init")
	zlog.New().Info("exec this command: ", zap.String("cmd: ", cmd.String()))
	//log.Printf("command: %v\n", cmd.String())

	rp, wp, err := NewPipe()
	if err != nil {
		zlog.New().Error("create pipe error: ", zap.Error(err))
		return nil, nil
	}

	// ExtraFiles 用于给新进程继承父进程中打开的文件
	// 这里将管道的读端 readPipe 传递给子进程，这样子进程就可以发送数据到管道中了
	// 一个进程默认有 3 个文件描述符，stdin，stdout 和 stderr
	cmd.ExtraFiles = []*os.File{rp}

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
func RunProcess(tty bool, cmds []string, resConf *subsystems.ResourceConfig) {
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

	p, wp := NewParentProcess(tty)
	// 运行并等待 cmd 执行完成
	p.Start()

	sendInitCommand(cmds, wp)

	cg := cgroup.NewCgroupManager("cgroup-test", resConf)
	defer func ()  {
		if err := cg.RemoveAll(); err != nil {
			zlog.New().Error("remove cgroup error", zap.Error(err))
			return
		}
		zlog.New().Info("remove cgroup success.")
	}()
	
	if err := cg.SetAll(); err != nil {
		zlog.New().Error("set cgroup error", zap.Error(err))
		return
	}

	if err := cg.ApplyAll(int64(p.Process.Pid)); err != nil {
		zlog.New().Error("set cgroup error", zap.Error(err))
		return
	}

	p.Wait()

	// 执行完成后，退出进程
	os.Exit(-1)
}
