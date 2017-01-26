package hardware

type Gpu struct {
    Model string
    CudaName string
    TotalMemory uint64
}

// Computer interface
func (c *Gpu) GetDeviceName() string {
	return "GPU"
}
func (c *Gpu) GetComputeDeviceType() string {
	return "CUDA"
}
func (c *Gpu) GetComputeDeviceName() string {
	return c.CudaName
}
func (c *Gpu) GetOptimalTileSize() int {
	if c.TotalMemory > 1073741824 {
		return 256
	} else {
		return 128
	}
}