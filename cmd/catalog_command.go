package cmd

import (
	"io/ioutil"

	"github.com/gitlawr/cihelper/model"
	"github.com/gitlawr/cihelper/service"
	"github.com/urfave/cli"
)

func CatalogCommand() cli.Command {
	catalogFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "repourl",
			Usage: "git url for catalog repo",
		},
		cli.StringFlag{
			Name:  "branch",
			Usage: "catalog repo branch",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "git username",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "git password",
		},
		cli.StringFlag{
			Name:  "cacheroot",
			Usage: "cache directory to store catalog items",
		},
		cli.StringFlag{
			Name:  "foldername",
			Usage: "catalog template folder name",
		},
		cli.BoolFlag{
			Name:  "system",
			Usage: "catalog template type",
		},
		cli.StringFlag{
			Name:  "compose-file",
			Usage: "docker-compose file path",
			Value: "./docker-compose.yml",
		},
		cli.StringFlag{
			Name:  "rancher-file",
			Usage: "rancher-compose file path",
			Value: "./rancher-compose.yml",
		},
		cli.StringFlag{
			Name:  "readme",
			Usage: "readme file path",
		},
	}

	return cli.Command{
		Name:   "catalog",
		Usage:  "upgrade catalog",
		Action: upgradeCatalog,
		Flags:  catalogFlags,
	}
}

func upgradeCatalog(ctx *cli.Context) error {
	factory := ClientFactory{}
	apiClient, _ := factory.GetClient(ctx)
	gitUser := ctx.String("user")
	gitToken, err := service.GetGitToken(apiClient, gitUser)
	if err != nil {
		return err
	}
	composeFile := ctx.String("compose-file")
	rancherFile := ctx.String("rancher-file")
	readmeFile := ctx.String("readme")
	dockerCompose := ""
	rancherCompose := ""
	readme := ""
	if composeFile != "" {
		cdat, err := ioutil.ReadFile(composeFile)
		check(err)
		dockerCompose = string(cdat)
	}
	if rancherFile != "" {
		rdat, err := ioutil.ReadFile(rancherFile)
		check(err)
		rancherCompose = string(rdat)
	}
	if readmeFile != "" {
		rmdat, err := ioutil.ReadFile(readmeFile)
		check(err)
		readme = string(rmdat)
	}
	config := &model.CatalogUpgrade{
		CacheRoot:          ctx.String("cacheroot"),
		GitUrl:             ctx.String("repourl"),
		GitBranch:          ctx.String("branch"),
		TemplateFolderName: ctx.String("foldername"),
		TemplateIsSystem:   ctx.Bool("system"),
		GitUser:            gitUser,
		GitPassword:        gitToken,

		DockerCompose:  dockerCompose,
		RancherCompose: rancherCompose,
		Readme:         readme,
	}
	return service.UpgradeCatalog(config)
}
