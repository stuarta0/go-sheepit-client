package hardware

import (
    "path"
)

func PlatformName() string {
    return "mac"
}

func RendererPath() string {
    return path.Join("Blender", "blender.app", "Contents", "MacOS", "blender")
}

func CpuStat() *Cpu {
    cpu := Cpu{}

    // TODO: implement mac stats

    return &cpu
}

func GpuStat() *Gpu {
    return &Gpu {}
}

func TotalMemory() uint64 {
    // TODO: Calculate total memory
    return 8350000
}