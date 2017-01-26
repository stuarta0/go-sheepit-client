package hardware

type Cpu struct {
    Name string
    Family string
    Model string
    Architecture string
    TotalCores int
}

func (cpu *Cpu) IsValid() bool {
    return cpu.Name != "" && cpu.Family != "" && cpu.Model != "" && cpu.Architecture != ""
}

// Computer interface
func (c *Cpu) GetDeviceName() string {
	return "CPU"
}
func (c *Cpu) GetComputeDeviceType() string {
	return "NONE"
}
func (c *Cpu) GetComputeDeviceName() string {
	return "CPU"
}
func (c *Cpu) GetOptimalTileSize() int {
	return 32
}