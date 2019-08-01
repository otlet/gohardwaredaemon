package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/jedib0t/go-pretty/table"
	"github.com/sevlyar/go-daemon"
)

var (
	tab        = table.NewWriter()
	json       = ""
	fileExport = flag.String("f", "", "Avaliable format: CSV, JSON, std")
	daemonFlag = flag.Bool("d", false, "a bool")
	signal     = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)
)

// Collector for memory information
func memory() {
	memory, err := ghw.Memory()
	if err != nil {
		tab.AppendRow([]interface{}{"MEMORY", "", "ERROR", err})
	} else {
		tab.AppendRow([]interface{}{"MEMORY", "", "INFO", memory.String()})
		tab.AppendRow([]interface{}{"MEMORY", "", "PHYSICAL_BYTES", memory.TotalPhysicalBytes})
		tab.AppendRow([]interface{}{"MEMORY", "", "USABLE_BYTES", memory.TotalUsableBytes})
		tab.AppendRow([]interface{}{"MEMORY", "", "SUPPORTED_PAGE_SIZES", memory.SupportedPageSizes})
	}
}

// Collector for processor information
func cpu() {
	cpu, err := ghw.CPU()
	if err != nil {
		tab.AppendRow([]interface{}{"CPU", "", "ERROR", err})
		return
	}

	tab.AppendRow([]interface{}{"CPU", "", "INFO", cpu})
	tab.AppendRow([]interface{}{"CPU", "", "TOTAL_CORES", cpu.TotalCores})
	tab.AppendRow([]interface{}{"CPU", "", "TOTAL_THREADS", cpu.TotalThreads})

	for index, proc := range cpu.Processors {
		tab.AppendRow([]interface{}{"CPU", fmt.Sprintf("CPU[%v]", index), "VENDOR", proc.Vendor})
		tab.AppendRow([]interface{}{"CPU", fmt.Sprintf("CPU[%v]", index), "MODEL", proc.Model})
		tab.AppendRow([]interface{}{"CPU", fmt.Sprintf("CPU[%v]", index), "THREADS", proc.NumThreads})
	}
}

// Collector for persistent memory information
func blockStrorage() {
	block, err := ghw.Block()
	if err != nil {
		tab.AppendRow([]interface{}{"STORAGE", "", "ERROR", err})
		return
	}

	tab.AppendRow([]interface{}{"STORAGE", "", "INFO", block})

	for index, disk := range block.Disks {
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "VENDOR", disk.Vendor})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "MODEL", disk.Model})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "NAME", disk.Name})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "SERIAL", disk.SerialNumber})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "WWN", disk.WWN})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "STORAGE_CONTROLLER", disk.StorageController})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "SIZE_BYTES", disk.SizeBytes})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "PHYSICAL_BLOCK_SIZE_BYTES", disk.PhysicalBlockSizeBytes})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "NUMA_NODE_ID", disk.NUMANodeID})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "DRIVE_TYPE", disk.DriveType})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "BUS_TYPE", disk.BusType})
		tab.AppendRow([]interface{}{"STORAGE", fmt.Sprintf("DISK[%v]", index), "BUS_PATH", disk.BusPath})
	}
}

// Collector for network information
func network() {
	net, err := ghw.Network()
	if err != nil {
		tab.AppendRow([]interface{}{"NETWORK", "", "ERROR", err})
		return
	}

	tab.AppendRow([]interface{}{"NETWORK", "", "INFO", net})

	for index, nic := range net.NICs {
		tab.AppendRow([]interface{}{"NETWORK", fmt.Sprintf("INTERFACE[%v]", index), "NAME", nic.Name})
		tab.AppendRow([]interface{}{"NETWORK", fmt.Sprintf("INTERFACE[%v]", index), "IS_VIRTUAL", nic.IsVirtual})
		if nic.IsVirtual == false {
			tab.AppendRow([]interface{}{"NETWORK", fmt.Sprintf("INTERFACE[%v]", index), "MAC", nic.MacAddress})
		}
	}
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
		Args:        []string{"goHardwareDaemon"},
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

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func worker() {
LOOP:
	for {
		render()
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
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}

// Render function, where we launch functions with hardware information
// collector.
func render() {
	err := os.Setenv("GHW_DISABLE_WARNINGS", "1")
	tab.SetOutputMirror(os.Stdout)
	tab.AppendHeader(table.Row{"TYPE", "ID", "DESCRIPTION", "VALUE"})

	if *fileExport != "" {

		memory()
		cpu()
		blockStrorage()
		network()

		if err != nil {
			log.Println(err)
			os.Exit(2)
		}

		if *fileExport == "csv" {
			tab.RenderCSV()
		} else if *fileExport == "json" {
			fmt.Println(generateJSON())
		} else if *fileExport == "std" {
			memory()
			cpu()
			blockStrorage()
			network()
			tab.Render()
		} else {
			fmt.Println("Accepted formats: json, csv, std")
		}
	} else {
		log.Println(generateJSON())
	}
}

// Generate JSON file with informations
func generateJSON() string {
	cpu, _ := ghw.CPU()
	memory, _ := ghw.Memory()
	block, _ := ghw.Block()
	net, _ := ghw.Network()
	json = fmt.Sprintf(
		"{\"cpu\":%v,\"memory\":%v,\"disk\":%v,\"network\":%v}",
		cpu.JSONString(false),
		memory.JSONString(false),
		block.JSONString(false),
		net.JSONString(false),
	)
	return json
}

// Main function
func main() {
	flag.Parse()
	if *daemonFlag {
		daemonMode()
	} else if *fileExport != "" {
		render()
	} else {
	}
}
