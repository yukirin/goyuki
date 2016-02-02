package main

import (
	"github.com/mitchellh/cli"
	"github.com/yukirin/goyuki/command"
)

func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"get": func() (cli.Command, error) {
			return &command.GetCommand{
				Meta: *meta,
			}, nil
		},
		"run": func() (cli.Command, error) {
			return &command.RunCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: GitCommit,
				Name:     Name,
			}, nil
		},
	}
}
