package main

import (
	"fmt"

	"github.com/YOUSEEBIGGIRL/fakedocke/cgroup/subsystems"
	"github.com/YOUSEEBIGGIRL/fakedocke/container"
	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"github.com/urfave/cli/v2"
)

var run = &cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			fakedocker run -it [process name], such as: fakedocker run -it /bin/bash`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it", // 该命令会分配一个伪终端，将本机的 stdio 与容器的 stdio 相关联
			Usage: "enable tty",
		},
		&cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		&cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		//&cli.StringFlag{
		//	Name:  "v",
		//	Usage: "volume",
		//},
	},
	Action: func(c *cli.Context) error {
		// 判断参数是否包含 command
		if c.Args().Len() < 1 {
			return fmt.Errorf("missing container command")
		}
		// 获取用户指定的 command
		// 只能执行一个命令
		// cmd := c.Args().Get(0)

		var cmds []string
		for _, v := range c.Args().Slice() {
			cmds = append(cmds, v)
		}

		tty := c.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: c.String("mem"),
			CPUShare:    c.String("cpushare"),
			CPUSet:      c.String("cpuset"),
		}
		volume := c.String("v")
		// 调用 RunProcess 启动容器进程
		container.RunProcess(tty, cmds, volume, resConf)
		return nil
	},
}

var init_ = &cli.Command{
	Name: "init",
	Usage: `init container process run user's process in container. 
				Do not call it outside`,
	Action: func(c *cli.Context) error {
		zlog.New().Info("init come on")
		return container.InitProcess()
	},
}
