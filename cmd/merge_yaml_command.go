package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/gitlawr/cihelper/service"
	"github.com/urfave/cli"
)

func MergeYamlCommand() cli.Command {
	return cli.Command{
		Name:        "mergeyaml",
		Usage:       "merge two yaml files",
		Description: "\nmerge two yaml files. \n\nExample:\n\t$ cihelper mergeyaml <file1> <file2>\n",
		Action:      merge,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output,o",
				Usage: "output file",
				Value: "merge_output.yml",
			},
		},
	}
}

func merge(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		return errors.New("arguments mismatch")
	}
	formerFile := args.Get(0)
	latterFile := args.Get(1)
	outputFile := ctx.String("output")
	fdat, err := ioutil.ReadFile(formerFile)
	check(err)
	ldat, err := ioutil.ReadFile(latterFile)
	check(err)
	mdat, err := service.MergeYaml(fdat, ldat)
	check(err)
	err = ioutil.WriteFile(outputFile, mdat, 0644)
	check(err)
	return nil
}
