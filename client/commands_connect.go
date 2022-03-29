/*
 * @Author: F1
 * @Date: 2022-03-29 16:20:34
 * @LastEditTime: 2022-03-29 16:43:24
 * @FilePath: /boundary/client/commands_connect.go
 * @Description:
 *
 * Copyright (c) 2022 by splashtop.com, All Rights Reserved.
 */
package client

import (
	"github.com/hashicorp/boundary/internal/cmd/base"
	"github.com/hashicorp/boundary/internal/cmd/commands/connect"
	"github.com/mitchellh/cli"
)

// Commands is the mapping of all the available commands.
var Commands map[string]cli.CommandFactory

func initCommands(ui, serverCmdUi cli.Ui, runOpts *RunOptions) {
	Commands = map[string]cli.CommandFactory{
		"connect": func() (cli.Command, error) {
			return &connect.Command{
				Command: base.NewCommand(ui),
				Func:    "connect",
			}, nil
		},
		"http": func() (cli.Command, error) {
			return &connect.Command{
				Command: base.NewCommand(ui),
				Func:    "http",
			}, nil
		},
		"kube": func() (cli.Command, error) {
			return &connect.Command{
				Command: base.NewCommand(ui),
				Func:    "kube",
			}, nil
		},
		"postgres": func() (cli.Command, error) {
			return &connect.Command{
				Command: base.NewCommand(ui),
				Func:    "postgres",
			}, nil
		},
		"rdp": func() (cli.Command, error) {
			return &connect.Command{
				Command: base.NewCommand(ui),
				Func:    "rdp",
			}, nil
		},
		"ssh": func() (cli.Command, error) {
			return &connect.Command{
				Command: base.NewCommand(ui),
				Func:    "ssh",
			}, nil
		},
	}
}
