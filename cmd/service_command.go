package cmd

import (
	"errors"
	"strings"

	"github.com/rancher/cihelper/model"
	"github.com/rancher/cihelper/service"
	"github.com/urfave/cli"
)

func ServiceCommand() cli.Command {
	serviceFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "image",
			Usage: "image to use",
		},
		cli.StringSliceFlag{
			Name:  "selector",
			Usage: "service selector labels",
		},
		cli.IntFlag{
			Name:  "batchsize",
			Usage: "batch size",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "interval",
			Usage: "batch interval in seconds",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "startfirst",
			Usage: "start before stopping",
		},
	}

	return cli.Command{
		Name:   "service",
		Usage:  "upgrade services",
		Action: upgrade,
		Flags:  serviceFlags,
	}
}

func upgrade(ctx *cli.Context) error {
	factory := ClientFactory{}
	apiClient, _ := factory.GetClient(ctx)
	selectors := ctx.StringSlice("selector")
	svcSelectors, err := envVarstoMap(selectors)
	if err != nil {
		return err
	}
	batchSize := ctx.Int64("batchsize")
	interval := ctx.Int64("interval")
	startFirst := ctx.Bool("startfirst")
	image := ctx.String("image")

	config := &model.ServiceUpgrade{
		ServiceSelector: svcSelectors,
		BatchSize:       batchSize,
		IntervalMillis:  interval,
		StartFirst:      startFirst,
	}
	service.UpgradeServices(apiClient, config, image)
	return nil
}

func envVarstoMap(vars []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, s := range vars {
		splits := strings.Split(s, "=")
		if len(splits) != 2 {
			return nil, errors.New("Parse selector '" + s + "' fail, needs the form 'FOO=BAR'")
		}
		key := splits[0]
		val := splits[1]
		m[key] = val
	}
	return m, nil
}
