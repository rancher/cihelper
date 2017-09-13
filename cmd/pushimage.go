package cmd

import (
	"errors"

	"github.com/gitlawr/cihelper/service"
	"github.com/urfave/cli"
)

func PushImageCommand() cli.Command {
	/*
		flags := []cli.Flag{
			cli.StringFlag{
				Name:   "image",
				Usage:  "Environment ENDPOINT URL",
				EnvVar: "CATTLE_URL",
			},
		}
	*/
	return cli.Command{
		Name:        "pushimage",
		Usage:       "push an image using credential from rancher registries configuration",
		Description: "\nPush an image using credential from rancher registries configuration. \nUse `global options` to choose api keys for a different environment.\n\nExample:\n\t$ cihelper pushimage <Image>\n",
		Action:      authAndPush,
		Flags:       []cli.Flag{},
	}
}

func authAndPush(ctx *cli.Context) error {
	factory := ClientFactory{}
	apiClient, _ := factory.GetClient(ctx)
	args := ctx.Args()
	if len(args) != 1 {
		return errors.New("arguments mismatch")
	}
	return service.AuthAndPush(apiClient, args.First())
}
