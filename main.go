package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	trunks "github.com/shynuu/trunks/runtime"
	"github.com/urfave/cli/v2"
)

// Initialize signals handling
func initSignals() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	func(_ os.Signal) {}(<-cancelChan)
	os.Exit(0)
}

func main() {
	var config string
	var flush bool = false
	var acm bool = false
	var qos bool = false
	var disable_kernel_version_check = false
	var logs string

	app := &cli.App{
		Name:                 "trunks",
		Usage:                "a simple DVB-S2/DVB-RCS2 simulator",
		EnableBashCompletion: true,
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
			&cli.StringFlag{
				Name:        "logs",
				Usage:       "Log path for the log file",
				Destination: &logs,
				Required:    false,
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
			&cli.BoolFlag{
				Name:        "qos",
				Usage:       "Process traffic using QoS",
				Destination: &qos,
				DefaultText: "not activated",
			},
			&cli.BoolFlag{
				Name:        "disable-kernel-version-check",
				Usage:       "Disable check for bugged kernel versions",
				Destination: &disable_kernel_version_check,
				DefaultText: "kernel version check enabled",
			},
		},
		Action: func(c *cli.Context) error {
			trunksConfig, err := trunks.InitTrunks(config, qos, logs, acm, disable_kernel_version_check)
			if err != nil {
				fmt.Println("Init error, exiting...")
				os.Exit(1)
			}

			if flush {
				err = trunksConfig.FlushTables()
				if err != nil {
					fmt.Println("Impossible to flush tables, exiting...")
					os.Exit(1)
				}
			}

			trunksConfig.Run()
			return nil
		},
	}
	go initSignals()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
