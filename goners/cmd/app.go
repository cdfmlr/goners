package main

import (
	"encoding/json"
	"fmt"

	"github.com/cdfmlr/goners"
	"github.com/urfave/cli/v2"
)

func commandDevices() *cli.Command {
	return &cli.Command{
		Name:  "devices",
		Usage: "Look up network interfaces (i.e. devices)",
		Flags: []cli.Flag{
			flagFormat(),
		},
		Action: func(ctx *cli.Context) error {
			devices, err := goners.LookupDevices()
			if err != nil {
				return cli.Exit(fmt.Sprintf("failed to lookup devices: %v.", err), 11)
			}

			switch ctx.String("format") {
			case "text":
				for _, d := range devices {
					fmt.Println(d.String())
				}
			case "json":
				j, err := json.Marshal(devices)
				if err != nil {
					return cli.Exit(fmt.Sprintf("failed to marshal json: %v.", err), 12)
				}
				fmt.Println(string(j))
			}
			return nil
		},
	}
}

func commandPcap() *cli.Command {
	flagCategoryOutput := `OUTPUT: outputs captured packets. 
	    Default output is STDOUT. (require a tty with 96 chars width for pretty-print text format)
	    (the requirement is satisfied if you can see above sentence in one line.)
	    Use one of --output FILE or --ws ADDR to override it.`

	return &cli.Command{
		Name:  "pcap",
		Usage: "Capture live packets from device. Root privilege is required.",
		// 大名鼎鼎的 urfave/cli 居然不支持位置参数。。难怪斗不过 spf13/cobra。
		ArgsUsage: "DEVICE\n\nARGUMENTS:\n\tDEVICE: name of the device to capture. Use \"goners devices\" to list available devices.",
		Flags: []cli.Flag{
			flagFormat(),
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output caputred packtes into `FILE`.",
				Category: flagCategoryOutput,
			},
			&cli.StringFlag{
				Name:     "ws",
				Usage:    "Output caputred packtes by WebSocket (listen and serve `ADDR`).",
				Category: flagCategoryOutput,
			},
		},
		Action: func(ctx *cli.Context) error {
			packets, err := goners.CaptureLivePackets(
				ctx.Args().First(), "", 16385, true, goners.BlockForever)
			if err != nil {
				cli.Exit(fmt.Sprintf("failed to capture live packets: %v", err), 13)
			}

			var formater goners.PacketsFormater
			switch ctx.String("format") {
			case "text":
				formater = goners.StringPacketsFormater
			case "json":
				formater = goners.JsonPacketsFormater
			}

			var out goners.Outputer
			switch {
			case ctx.String("output") != "":
				f := ctx.String("output")
				if out, err = goners.NewFileOutputer(f); err != nil {
					cli.Exit(fmt.Sprintf("failed to output into %v: %v", f, err), 14)
				}
			case ctx.String("ws") != "":
				// TODO
				panic("not implemented")
			default:
				// for p := range packets {
				// 	fmt.Println(p.String())
				// }
				// return nil
				if out, err = goners.NewFileOutputer("/dev/stdout"); err != nil {
					cli.Exit(fmt.Sprintf("failed to output into /dev/stdout: %v", err), 14)
				}
			}

			out.Output(formater.FormatPackets(packets))

			return nil
		},
	}
}

func flagFormat() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "format",
		Value: "text",
		Usage: "Output `FORMAT`: text | json\n\ttext: our human preferred text.\n\tjson: the JSON format (more readable for machines)\n",
		Action: func(ctx *cli.Context, s string) error {
			available := []string{"text", "json"}
			for _, a := range available {
				if s == a {
					return nil
				}
			}
			return fmt.Errorf("unexpected format %v. Available: %v", s, available)
		},
	}
}

var app = &cli.App{
	Name:  "goners",
	Usage: "goner's oafish network explorer & reliable sniffer",
	Commands: []*cli.Command{
		commandDevices(),
		commandPcap(),
	},
	Action: func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	},
}
