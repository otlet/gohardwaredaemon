package main

import (
	"fmt"
	"github.com/jaypipes/ghw"
	"github.com/jedib0t/go-pretty/table"
	"os"
)

var (
	tab  = table.NewWriter()
	json = ""
)

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

func main() {
	tab.SetOutputMirror(os.Stdout)
	tab.AppendHeader(table.Row{"TYPE", "ID", "DESCRIPTION", "VALUE"})

	if len(os.Args) > 1 {
		err := os.Setenv("GHW_DISABLE_WARNINGS", "1")

		memory()
		cpu()
		blockStrorage()
		network()

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		if os.Args[1] == "csv" {
			tab.RenderCSV()
		} else if os.Args[1] == "json" {
			collect()
		}
	} else {
		memory()
		cpu()
		blockStrorage()
		network()

		tab.Render()
	}
}

func collect() {
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
	fmt.Println(json)
}
