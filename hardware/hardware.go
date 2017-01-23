package hardware

type Cpu struct {
	Name string
	Family string
	Model string
	Architecture string
	TotalCores int
}

type Gpu struct {
	Model string
	CudaName string
	TotalMemory uint64
}

// A catch-all interface to get details about software and hardware
type Platform interface {

	// Gets the name of the platform, e.g. "windows" or "macos"
	Name() string

	// Gets the path to the Blender executable suitable for the platform
	RendererPath() string

	// Gets the CPU details of this platform
	Cpu() *Cpu

	// Gets the GPU details of this platform
	Gpu() *Gpu

	// Gets the total system memory in MB
	TotalMemory() uint64
}

func (cpu *Cpu) IsValid() bool {
	return cpu.Name != "" && cpu.Family != "" && cpu.Model != "" && cpu.Architecture != ""
}