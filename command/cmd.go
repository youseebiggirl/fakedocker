package command

import (
	"fmt"
	"log"

	"github.com/YOUSEEBIGGIRL/fakedocke/container"
	"github.com/urfave/cli/v2"
)

func Run() *cli.Command {
	cmd := &cli.Command{
		Name: "run",
		Usage: `Create a container with namespace and cgroups limit
				fakedocker run -it [process name], such as: fakedocker run -it /bin/bash`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "it",	// 该命令会分配一个伪终端，将本机的 stdio 与容器的 stdio 相关联
				Usage: "enable it",
			},
		},
		Action: func(c *cli.Context) error {
			// 判断参数是否包含 command
			if c.Args().Len() < 1 {
				return fmt.Errorf("missing container command")
			}
			// 获取用户指定的 command
			cmd := c.Args().Get(0)
			tty := c.Bool("it")
			// 调用 RunProcess 启动容器进程
			container.RunProcess(tty, cmd)
			return nil
		},
	}
	return cmd
}

func Init() *cli.Command {
	cmd := &cli.Command{
		Name: "init",
		Usage: `init container process run user's process in container. 
				Do not call it outside`,
		Action: func(c *cli.Context) error {
			log.Println("init come on")
			cmd := c.Args().Get(0)
			log.Println("init args: ", cmd)
			return container.InitProcess(cmd)
		},
	}
	return cmd
}