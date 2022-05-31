package apps

import (
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func NewFlags() []cli.Flag {
	return []cli.Flag{

		altsrc.NewStringFlag(&cli.StringFlag{Name: "useragent", Aliases: []string{"u"}, Usage: "Append to the http user agent string"}),

		altsrc.NewStringFlag(&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Value: "~/.ssh/authorized_keys", Usage: "Write output to file"}),

		altsrc.NewBoolFlag(&cli.BoolFlag{Name: "color", Aliases: []string{"c"}, Value: true, Hidden: true, Usage: "log color"}),

		altsrc.NewBoolFlag(&cli.BoolFlag{Name: "remove", Aliases: []string{"r"}, Value: false, Usage: "Remove a key from authorized keys file"}),
	}
}
