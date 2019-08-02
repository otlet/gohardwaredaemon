package hardware

import (
	"fmt"
	"log"
	"os"

	"github.com/jaypipes/ghw"
	"github.com/jedib0t/go-pretty/table"
)

// Hardware is struct to access hardware informations
type Hardware struct{}

var (
	tab  = table.NewWriter()
	json = ""
)

// Generate function
func (hw Hardware) Generate(fileExport string) {
	err := os.Setenv("GHW_DISABLE_WARNINGS", "1")
	tab.SetOutputMirror(os.Stdout)
	tab.AppendHeader(table.Row{"TYPE", "ID", "DESCRIPTION", "VALUE"})

	if fileExport != "" {
		memory()
		cpu()
		blockStrorage()
		network()

		if err != nil {
			log.Println(err)
			os.Exit(2)
		}

		switch fileExport {
		case "json":
			fmt.Println(generateJSON())
		case "csv":
			tab.RenderCSV()
		case "std":
			tab.Render()
		default:
			fmt.Println("Accepted formats: json, csv, std")
		}
	}
	log.Println(generateJSON())
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
