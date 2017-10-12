package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gitlawr/cihelper/cmd"
	"github.com/urfave/cli"
)

var VERSION = "dev"

var AppHelpTemplate = `{{.Usage}}
Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]
Version: {{.Version}}
{{if .Flags}}
Options:
  {{range .Flags}}{{if .Hidden}}{{else}}{{.}}
  {{end}}{{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var CommandHelpTemplate = `{{.Usage}}
{{if .Description}}{{.Description}}{{end}}
Usage: cihelper [global options] {{.Name}} {{if .Flags}}[OPTIONS] {{end}}{{if ne "None" .ArgsUsage}}{{if ne "" .ArgsUsage}}{{.ArgsUsage}}{{else}}[arg...]{{end}}{{end}}
{{if .Flags}}Options:{{range .Flags}}
	 {{.}}{{end}}{{end}}
`

func main() {
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate

	app := cli.NewApp()
	app.Name = "cihelper"
	app.Usage = "Tool for smoothing ci in rancher"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Version = VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
		},
		cli.StringFlag{
			Name:   "envurl",
			Usage:  "Environment ENDPOINT URL",
			EnvVar: "CATTLE_URL",
		},
		cli.StringFlag{
			Name:   "accesskey",
			Usage:  "Environment ACCESS KEY",
			EnvVar: "CATTLE_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "secretkey",
			Usage:  "Environment SECRET KEY",
			EnvVar: "CATTLE_SECRET_KEY",
		},
	}

	app.Commands = []cli.Command{
		cmd.PushImageCommand(),
		cmd.UpgradeCommand(),
		cmd.MergeYamlCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
