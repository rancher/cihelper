package cmd

import (
	"github.com/urfave/cli"
)

func UpgradeCommand() cli.Command {
	return cli.Command{
		Name:        "upgrade",
		Usage:       "upgrade resources in rancher",
		Description: "\nupgrade resources in rancher. \nUse `global options` to choose api keys for a different environment.\n\nExample:\n\t$ cihelper upgrade <resource type>\n",
		Action:      cli.ShowAppHelp,
		Flags:       []cli.Flag{},
		Subcommands: []cli.Command{
			ServiceCommand(),
			StackCommand(),
			CatalogCommand(),
		},
	}
}
