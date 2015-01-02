package main

import (
	"flag"
	"github.com/arkenio/goarken"
	"github.com/codegangsta/cli"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

const (
	progname = "arkenctl"
	version  = "0.0.1"
)

func main() {

	stopBroadcaster := goarken.NewBroadcaster()

	app := cli.NewApp()
	app.Name = progname
	app.Usage = "inspect the arken cluster"
	app.Version = version
	app.Flags = GetGlobalFlags()
	app.Commands = GetCommands(stopBroadcaster.Listen())

	// Hack to be able to use flag for glog
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.Usage = func() {}
	flag.Parse()

	glog.Infof("%s starting", progname)
	app.Run(os.Args)
}

func handleSignals() {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR1)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR2)

	go func() {
		isProfiling := false

		defer func() {
			if isProfiling {
				pprof.StopCPUProfile()
			}
		}()

		for {
			sig := <-signals
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				//Exit gracefully
				glog.Info("Shutting down...")
				os.Exit(0)
			case syscall.SIGUSR1:
				pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
			case syscall.SIGUSR2:
				if !isProfiling {
					f, err := os.Create("/tmp/arkenctl.profile")
					if err != nil {
						glog.Fatal(err)
					} else {
						pprof.StartCPUProfile(f)
						isProfiling = true
					}
				} else {
					pprof.StopCPUProfile()
					isProfiling = false
				}

			}
		}

	}()

}
