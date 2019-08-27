package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/otlet/gohardwaredaemon/hardware"
	"github.com/sevlyar/go-daemon"
)

var (
	hw         = hardware.Hardware{}
	fileExport = flag.String("f", "", "Available format: CSV, JSON, std")
	daemonFlag = flag.Bool("d", false, "Run in daemon mode")
	signal     = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file
  REQUIRE -d (Daemon mode)`)

	stop = make(chan struct{})
	done = make(chan struct{})
)

func worker() {
LOOP:
	for {
		hw.Generate(*fileExport)
		time.Sleep(time.Hour)
		select {
		case <-stop:
			break LOOP
		default:
		}
	}
	done <- struct{}{}
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT || sig == syscall.SIGTERM || sig == syscall.SIGKILL {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}

func daemonMode() {
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	cntxt := &daemon.Context{
		PidFileName: "gohardwaredaemon.pid",
		PidFilePerm: 0644,
		LogFileName: "gohardwaredaemon.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"gohardwaredaemon"},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		fmt.Println("Daemon is nil")
		return
	}
	defer cntxt.Release()

	log.Println("- - - - - - - - - - - - - - -")
	log.Println("daemon started")

	go worker()

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println("daemon terminated")
}

// Main function
func main() {
	flag.Parse()
	if *daemonFlag {
		fmt.Println("DAEMON MODE")
		daemonMode()
	} else if *fileExport != "" {
		hw.Generate(*fileExport)
	} else {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}
