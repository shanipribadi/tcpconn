package main

import (
	"context"
	"flag"
	"log/slog"
	"strings"
	"tcpconn"
	"time"

	"golang.org/x/sync/errgroup"
)

var targetFlag = flag.String("target", "", "")
var bindsFlag = flag.String("binds", "", "comma separated list of binds")
var connsFlag = flag.Int("conns", 1, "number of conns")

func main() {
	log := slog.Default()
	flag.Parse()
	binds := strings.Split(*bindsFlag, ",")

	log.Info("flag.Parse()", "target", *targetFlag, "binds", binds)
	t := tcpconn.New(binds)

	g, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < *connsFlag; i++ {
		g.Go(func() error {
			return t.Run(ctx, *targetFlag, time.Second*60)
		})
	}

  log.Info("call t.StartRead()")
  t.StartRead()
  time.Sleep(time.Second * 1)
  log.Info("call t.StartWrite()")
  t.StartWrite()
  time.Sleep(time.Second * 1)
  log.Info("call t.StartClose()")
  t.StartClose()

	if err := g.Wait(); err != nil {
		log.Info("g.Wait()", "error", err)
	}
}
