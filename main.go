package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shynuu/go-trunks/trunks"

	"github.com/urfave/cli/v2"
)

func main() {
	var config string
	var flush bool = false
	var acm bool = false

	app := &cli.App{
		Name:  "trunks",
		Usage: "a simple DVB-S2/DVB-RCS2 simulator",
		Authors: []*cli.Author{
			{Name: "Youssouf Drif"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Load configuration from `FILE`",
				Destination: &config,
				Required:    true,
				DefaultText: "not set",
			},
			&cli.BoolFlag{
				Name:        "flush",
				Usage:       "Flush IPTABLES table mangle and clear all TC rules",
				Destination: &flush,
			},
			&cli.BoolFlag{
				Name:        "acm",
				Usage:       "Activate the ACM simulation",
				Destination: &acm,
				DefaultText: "not activated",
			},
		},
		Action: func(c *cli.Context) error {
			err := trunks.InitTrunks(config)
			if err != nil {
				fmt.Println("Init error, exiting...")
				os.Exit(1)
			}

			if flush {
				err = trunks.FlushTables()
				if err != nil {
					fmt.Println("Impossible to flush tables, exiting...")
					os.Exit(1)
				}
			}

			trunks.Run(acm)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
