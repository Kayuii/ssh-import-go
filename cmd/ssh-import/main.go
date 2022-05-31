package main

import (
	"log"
	"os"
	"time"

	"github.com/kayuii/ssh-import-go/apps"
	"github.com/kayuii/ssh-import-go/routes"
	"github.com/kayuii/ssh-import-go/version"
	"github.com/urfave/cli/v2"
)

func Action(c *cli.Context) error {
	return routes.New(c).Exec()
}

func main() {
	app := cli.NewApp()
	app.Name = version.APP_NAME
	app.Version = version.APP_VERSION
	app.Description = version.APP_DESC
	app.Usage = version.APP_USAGE
	app.Compiled = time.Now()
	app.Authors = []*cli.Author{
		{
			Name:  version.APP_AUTHOR,
			Email: version.APP_EMAIL,
		},
	}
	app.Flags = apps.NewFlags()
	// app.Before =
	app.Action = Action
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
