// Copyright (c) 2024, DCSO GmbH
package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mcuadros/go-syslog"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mini-syslog-receiver",
		Usage: "receive and dump syslog data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   "0.0.0.0",
				Usage:   "address to listen on (0.0.0.0 means all interfaces)",
				Action: func(ctx *cli.Context, v string) error {
					if val := net.ParseIP(v); val == nil {
						return fmt.Errorf("IP value %v not parseable", v)
					}
					return nil
				},
			},
			&cli.Uint64Flag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   514,
				Usage:   "port to listen on",
				Action: func(ctx *cli.Context, v uint64) error {
					if v >= 65536 {
						return fmt.Errorf("port value %v out of range[0-65535]", v)
					}
					return nil
				},
			},
			&cli.Uint64Flag{
				Name:    "sample",
				Aliases: []string{"m"},
				Value:   1000,
				Usage:   "sample up to <value> log entries, then exit",
			},
			&cli.BoolFlag{
				Name:    "tcp",
				Aliases: []string{"t"},
				Value:   false,
				Usage:   "use TCP instead of UDP",
			},
			&cli.BoolFlag{
				Name:    "tls",
				Aliases: []string{"s"},
				Value:   false,
				Usage:   "use TLS for TCP server",
			},
			&cli.StringFlag{
				Name:  "tls-key",
				Value: "",
				Usage: "PEM private key file to use for TCP/TLS server",
			},
			&cli.StringFlag{
				Name:  "tls-chain",
				Value: "",
				Usage: "PEM certificate chain file to use for TCP/TLS server",
			},
			&cli.StringFlag{
				Name:    "outfile",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "file to write output to (print to console if empty)",
			},
		},
		Action: func(ctx *cli.Context) error {
			channel := make(syslog.LogPartsChannel)
			handler := syslog.NewChannelHandler(channel)

			var err error
			server := syslog.NewServer()
			server.SetFormat(syslog.Automatic)
			server.SetHandler(handler)

			addr := fmt.Sprintf("%s:%d", ctx.String("listen"), ctx.Uint64("port"))
			if ctx.Bool("tcp") {
				if ctx.Bool("tls") {
					tlsC := ctx.String("tls-chain")
					if len(tlsC) == 0 {
						log.Fatal("TLS chain missing")
					}
					tlsK := ctx.String("tls-key")
					if len(tlsK) == 0 {
						log.Fatal("TLS key missing")
					}
					var cert tls.Certificate
					cert, err = tls.LoadX509KeyPair(tlsC, tlsK)
					if err != nil {
						log.Fatal(err)
					}
					cfg := &tls.Config{
						Certificates: []tls.Certificate{cert},
					}
					log.Println("using TCP/TLS", addr)
					err = server.ListenTCPTLS(addr, cfg)
				} else {
					log.Println("using TCP", addr)
					err = server.ListenTCP(addr)
				}
			} else {
				log.Println("using UDP", addr)
				err = server.ListenUDP(addr)
			}
			if err != nil {
				log.Fatal(err)
			}

			outfile := ctx.String("outfile")
			var logFunc func(val string)
			var f *os.File

			if len(outfile) > 0 {
				log.Println("writing to file", outfile)
				// log to file
				f, err = os.OpenFile(outfile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					log.Fatal(err)
				}
				logFunc = func(val string) {
					if _, err := f.WriteString(val + "\n"); err != nil {
						log.Println(err)
					}
				}
			} else {
				// log to console
				logFunc = func(val string) {
					fmt.Println(val)
				}
			}

			err = server.Boot()
			if err != nil {
				log.Fatal(err)
			}

			go func(channel syslog.LogPartsChannel) {
				var i uint64
				limit := ctx.Uint64("sample")
				for logParts := range channel {
					j, err := json.Marshal(logParts)
					if err != nil {
						log.Println(err)
						continue
					}
					logFunc(string(j))
					i += 1
					if limit > 0 && i >= limit {
						log.Printf("sample limit of %d log entries reached\n", limit)
						if f != nil {
							log.Println("closing output file")
							f.Close()
						}
						os.Exit(0)
					}
				}
			}(channel)

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				if f != nil {
					log.Println("closing output file")
					f.Close()
				}
				os.Exit(0)
			}()

			server.Wait()
			return nil
		},
	}

	app.HideHelpCommand = true
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
