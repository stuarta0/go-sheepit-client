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

func (cpu *Cpu) IsValid() bool {
    return cpu.Name != "" && cpu.Family != "" && cpu.Model != "" && cpu.Architecture != ""
}