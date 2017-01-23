package hardware

import (
	"os"
	"regexp"
	//"log"

	"golang.org/x/sys/windows/registry"
)

type Windows struct {}

func (w Windows) Name() string {
	return "windows"
}

func (w Windows) RendererPath() string {
	return "rend.exe"
}

func (w Windows) Cpu() *Cpu {
	//cpu := Cpu{Family: "Intel", Name: "Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz", Model: "i7-6700", Cores: 8, Architecture: "64bit"}
	cpu := Cpu{Family: "family", Name: "name", Model: "model", TotalCores: 0, Architecture: "arch"}

	// "Intel64 Family 6 Model 94 Stepping 3, GenuineIntel"
	env := os.Getenv("PROCESSOR_IDENTIFIER")
	family := regexp.MustCompile(`Family\s+(?P<family>[^\s]+)`)
	model := regexp.MustCompile(`Model\s+(?P<model>[^\s]+)`)
	if match := family.FindStringSubmatch(env); match != nil {
		cpu.Family = match[1]
	}
	if match := model.FindStringSubmatch(env); match != nil {
		cpu.Model = match[1]
	}

	// HARDWARE\DESCRIPTION\System\CentralProcessor\0 ProcessorNameString
	// "Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz"
	cpu_key := `HARDWARE\DESCRIPTION\System\CentralProcessor`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, cpu_key, registry.READ)
	defer k.Close()
	if err == nil {
		if subkeys, _ := k.ReadSubKeyNames(-1); len(subkeys) > 0 {
			cpu.TotalCores = len(subkeys) // TODO: check this is a valid method of determining core count
			k2, err := registry.OpenKey(registry.LOCAL_MACHINE, cpu_key + `\` + subkeys[0], registry.QUERY_VALUE)
			defer k2.Close()
			if err == nil {
				if s, _, err := k2.GetStringValue("ProcessorNameString"); err == nil {
					cpu.Name = s
				}
			}
		}
	}

	arch := os.Getenv("PROCESSOR_ARCHITEW6432")
	if arch == "" {
		arch = os.Getenv("PROCESSOR_ARCHITECTURE")
	}
	if arch == "AMD64" {
		cpu.Architecture = "64bit"
	} else {
		cpu.Architecture = "32bit"
	}

	return &cpu
}

func (w Windows) Gpu() *Gpu {
	return &Gpu {Model: "GeForce GTX 780", CudaName: "CUDA_0", TotalMemory: 3355443}
}

func (w Windows) TotalMemory() uint64 {
	// TODO: Calculate total memory
	return 16777216
}