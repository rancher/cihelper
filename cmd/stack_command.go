package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gitlawr/cihelper/model"
	"github.com/gitlawr/cihelper/service"
	"github.com/urfave/cli"
)

func StackCommand() cli.Command {
	stackFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "stackname",
			Usage: "stack name to upgrade",
		},
		cli.StringFlag{
			Name:  "env-file",
			Usage: "env file to use in catalog",
		},
		cli.StringFlag{
			Name:  "compose-file",
			Usage: "docker compose file for stack upgrade",
		},
		cli.StringFlag{
			Name:  "rancher-file",
			Usage: "rancher compose file for stack upgrade",
		},
		cli.StringFlag{
			Name:  "externalId",
			Usage: "externalId for a catalog template",
		},
		cli.BoolFlag{
			Name:  "tolatest",
			Usage: "upgrade stack to latest catalog version",
		},
	}

	return cli.Command{
		Name:   "stack",
		Usage:  "upgrade stack",
		Action: upgradeStack,
		Flags:  stackFlags,
	}
}

func upgradeStack(ctx *cli.Context) error {
	factory := ClientFactory{}
	apiClient, _ := factory.GetClient(ctx)

	var envs map[string]interface{}
	if ctx.String("env-file") != "" {
		envs = parseCustomEnvFile(ctx.String("env-file"))
	}
	dockerCompose := ""
	if ctx.String("compose-file") != "" {
		ddat, err := ioutil.ReadFile(ctx.String("compose-file"))
		check(err)
		dockerCompose = string(ddat)
	}
	rancherCompose := ""
	if ctx.String("rancher-file") != "" {
		rdat, err := ioutil.ReadFile(ctx.String("rancher-file"))
		check(err)
		rancherCompose = string(rdat)
	}
	config := &model.StackUpgrade{
		CattleUrl:       ctx.GlobalString("envurl"),
		AccessKey:       ctx.GlobalString("accesskey"),
		SecretKey:       ctx.GlobalString("secretkey"),
		StackName:       ctx.String("stackname"),
		DockerCompose:   dockerCompose,
		RancherCompose:  rancherCompose,
		Environment:     envs,
		ExternalId:      ctx.String("externalId"),
		ToLatestCatalog: ctx.Bool("tolatest"),
	}

	return service.UpgradeStack(apiClient, config)
}

func parseCustomEnvFile(file string) map[string]interface{} {
	variables := map[string]interface{}{}

	f, err := os.Open(file)
	if err != nil {
		logrus.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		parts := strings.SplitN(t, "=", 2)
		if len(parts) == 1 {
			variables[parts[0]] = ""
		} else {
			variables[parts[0]] = parts[1]
		}
	}

	if scanner.Err() != nil {
		logrus.Fatal(scanner.Err())
	}

	return variables
}
