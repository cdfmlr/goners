package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cdfmlr/goners"
	"github.com/cdfmlr/goners/api"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
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
				log.Fatalf("failed to lookup devices: %v.", err)
			}

			switch ctx.String("format") {
			case "text":
				for _, d := range devices {
					fmt.Println(d.String())
				}
			case "json":
				j, err := json.Marshal(devices)
				if err != nil {
					log.Fatalf("failed to marshal json: %v.", err)
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

	flagCategoryConfig := `CONFIG: configures the pcap.`

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
				Usage:    "Output caputred packtes by WebSocket (listen `ADDR` and serve ws at \"/\").",
				Category: flagCategoryOutput,
			},
			&cli.StringFlag{
				Name:     "filter",
				Usage:    "sets a `BPF` filter for the pcap (syntax reference: https://biot.com/capstats/bpf.html).",
				Category: flagCategoryConfig,
			},
			&cli.IntFlag{
				Name:     "snaplen",
				Aliases:  []string{"s"},
				Value:    262144, // tcpdump default snaplen
				Usage:    "Snarf snaplen `BYTES` of data from each packet. Packets will be truncated because of a limited snapshot",
				Category: flagCategoryConfig,
			},
			&cli.BoolFlag{
				Name:     "promisc",
				Usage:    "whether to put the interface in promiscuous mode",
				Value:    false,
				Category: flagCategoryConfig,
			},
			&cli.Int64Flag{
				Name:        "timeout",
				Usage:       "timeout in `SECONDS` to stop the capturing. <0 means block forever.",
				Value:       int64(goners.BlockForever),
				DefaultText: "BlockForever",
				Category:    flagCategoryConfig,
			},
		},
		Action: func(ctx *cli.Context) error {
			var timeout time.Duration
			if ctx.Int64("timeout") < 0 {
				timeout = goners.BlockForever
			} else {
				timeout = time.Second * time.Duration(ctx.Int64("timeout"))
			}

			packets, err := goners.CaptureLivePackets(
				context.Background(),
				ctx.Args().First(),
				ctx.String("filter"),
				int32(ctx.Int("snaplen")),
				ctx.Bool("promisc"),
				timeout,
			)
			if err != nil {
				log.Fatalf("failed to capture live packets: %v", err)
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
					log.Fatalf("failed to output into %v: %v", f, err)
				}
			case ctx.String("ws") != "":
				addr := ctx.String("ws")

				var ws websocket.Handler
				out, ws = goners.NewWebSocketOutputer()

				go func() {
					mux := http.NewServeMux()
					mux.Handle("/", ws)
					slog.Info("Listen and serve http",
						"addr", addr, "websocket", "/")
					if err := http.ListenAndServe(addr, mux); err != nil {
						log.Fatalf("failed to listen and serve ws: %v", err)
					}
				}()
			default:
				if out, err = goners.NewFileOutputer("/dev/stdout"); err != nil {
					log.Fatalf("failed to output into /dev/stdout: %v", err)
				}
			}

			out.Output(formater.FormatPackets(packets))

			return nil
		},
	}
}

func commandHttp() *cli.Command {
	apiUsage := `
	devicse:
		GET    /devices           lookup devices
	pcap:
		POST   /pcap              start a capturing session
		DELETE /pcap              stop & close a capturing session
		WS     /pcap/{sessionID}  get packets`

	return &cli.Command{
		Name:  "http",
		Usage: "Listen and serve goners api service on HTTP.\n" + apiUsage,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: "localhost:9800",
				Usage: "start HTTP service on `HOST:PORT`",
			},
		},
		Action: func(ctx *cli.Context) error {
			r := api.NewHttp()
			if err := r.Run(ctx.String("addr")); err != nil {
				log.Fatalf("Run HTTP failed with error: %v", err)
			}
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
		commandHttp(),
	},
	Action: func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	},
}
