package hardware

import (
    "errors"
    "fmt"
    "os/exec"
    "path"
    "strconv"
    "strings"
)

func PlatformName() string {
    return "mac"
}

func RendererPath() string {
    return path.Join("Blender", "blender.app", "Contents", "MacOS", "blender")
}

func getSysctl(keys ...string) (string, error) {
    cmd := exec.Command("sysctl", keys...)
    out, err := cmd.Output()
    if err != nil {
        return "", err
    }

    key := fmt.Sprintf("%s:", keys[0])
    value := string(out)
    if strings.HasPrefix(value, key) {
        return strings.TrimSpace(strings.Replace(value, key, "", -1)), nil
    } else {
        return strings.TrimSpace(value), errors.New("Could not determine key:value, returned raw output")
    }
}

func CpuStat() *Cpu {
    //cpu := Cpu{Family: "6", Name: "Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz", Model: "i7-6700", Cores: 8, Architecture: "64bit"}
    cpu := Cpu{Family: "", Name: "", Model: "Unknown", TotalCores: 0, Architecture: "32bit"}

    cpu.Family, _ = getSysctl("machdep.cpu.family")
    cpu.Name, _ = getSysctl("machdep.cpu.brand_string")
    
    cmd := exec.Command("getconf", "LONG_BIT")
    if out, err := cmd.Output(); err == nil {
        if strings.TrimSpace(string(out)) == "64" {
            cpu.Architecture = "64bit"
        }
    }

    // http://stackoverflow.com/a/1715612
    if cores, _ := getSysctl("-n", "hw.ncpu"); cores != "" {
        if icores, err := strconv.ParseInt(cores, 10, 32); err == nil {
            cpu.TotalCores = int(icores)
        }
    }

    return &cpu
}

func GpuStat() *Gpu {
    return &Gpu {}
}

func TotalMemory() uint64 {
    memory, err := getSysctl("hw.memsize")
    if err == nil {
        if size, err := strconv.ParseUint(memory, 10, 64); err == nil {
            return size / 1024
        }
    }

    return 0
}