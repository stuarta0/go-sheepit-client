package common

import (
    "bufio"
    //"errors"
    "fmt"
    "io/ioutil"
    "log"
    "math"
    "os"
    "os/exec"
    "path"
    "regexp"
    "strconv"
    "strings"

    "github.com/stuarta0/go-sheepit-client/hardware"
)

type Job struct {
    ArchiveMd5 string      `xml:"archive_md5,attr"`
    Id int                 `xml:"id,attr"`
    UseGpu bool            `xml:"use_gpu,attr"`
    Path string            `xml:"path,attr"`
    Frame int              `xml:"frame,attr"`
    SynchronousUpload bool `xml:"synchronous_upload,attr"`
    Extras string          `xml:"extras,attr"`
    Name string            `xml:"name,attr"`
    Password string        `xml:"password,attr"`

    Renderer *Renderer     `xml:"renderer"`
    Script string          `xml:"script"`

    RootPath string
}

// FileArchive interface
func (j Job) GetExpectedHash() string {
    return j.ArchiveMd5
}
func (j Job) GetArchivePath() string {
    return path.Join(j.RootPath, fmt.Sprintf("%s.zip", j.ArchiveMd5))
}
func (j Job) GetContentPath() string {
    return path.Join(j.RootPath, j.ArchiveMd5)
}

func (j *Job) Render(device hardware.Computer, config Configuration) error {
    fmt.Println("Rendering")

    // set cores based on CPU capabilities or override
    useCores := hardware.CpuStat().TotalCores
    if config.UseCores > 0 {
        useCores = config.UseCores
    }

    // set tile size based on optimal size for given hardware, or using override
    tileSize := device.GetOptimalTileSize()
    if config.TileSize > 0 {
        tileSize = config.TileSize
    }

    // String core_script = "import bpy\n" + "bpy.context.user_preferences.system.compute_device_type = \"%s\"\n" + "bpy.context.scene.cycles.device = \"%s\"\n" + "bpy.context.user_preferences.system.compute_device = \"%s\"\n";
    // if using GPU and has GPU: core_script % ("CUDA", "GPU", gpu.CudaName())
    // else: core_script % ("NONE", "CPU", "CPU")
    // core_script += String.format("bpy.context.scene.render.tile_x = %1$d\nbpy.context.scene.render.tile_y = %1$d\n", getTileSize());
    script := fmt.Sprintf(
        "%s\n" +
        "import bpy\n" + 
        "bpy.context.user_preferences.system.compute_device_type = \"%s\"\n" + 
        "bpy.context.scene.cycles.device = \"%s\"\n" +
        "bpy.context.user_preferences.system.compute_device = \"%s\"\n" +
        "bpy.context.scene.render.tile_x = %[5]d\n" + 
        "bpy.context.scene.render.tile_y = %[5]d\n", 
        j.Script, device.GetComputeDeviceType(), device.GetDeviceName(), device.GetComputeDeviceName(), tileSize)

    // minor difference - script added to content path (will be cleaned up when job directory is deleted)
    scriptPath := path.Join(j.GetContentPath(), fmt.Sprintf("%d_script.py", j.Id))
    if err := ioutil.WriteFile(scriptPath, ([]byte)(script), 0755); err != nil {
        return err
    }

    // command = job['renderer.commandline']
    // replace in command string:
        // ".c": "$scenepath -P $scriptpath", where job['script'] has been written to "working directory\script_<randint>" (no extension), defer delete file until render complete (i.e. job.render exits)
        // ".e": "$rendererpath" + "-t $cpucores" if cpucores specified by user (default all cores)
        // ".o": "$workingdir\$job.id_" (i.e. frame render path; blender will add frame number and extension)
        // ".f": "$job.frame"

    r := regexp.MustCompile(`(?:^|\s)(\.[ecof])(?:\s|$)`)
    cmd := r.ReplaceAllStringFunc(j.Renderer.Command, func (match string) string {
        repl := match
        switch strings.TrimSpace(match) {
        case ".e": // blender executable (used with exec.Command)
            if useCores > 0 {
                repl = fmt.Sprintf("-t %d", useCores)
            } else {
                repl = ""
            }
        case ".c": // .blend path and python script
            repl = fmt.Sprintf("%s -P %s", path.Join(j.GetContentPath(), j.Path), scriptPath)
        case ".o": // output image
            repl = path.Join(j.RootPath, fmt.Sprintf("%d_", j.Id))
        case ".f": // frame #
            repl = fmt.Sprintf("%d", j.Frame)
        }
        return fmt.Sprintf(" %s ", repl)
    })

    // set env vars:
        // BLENDER_USER_CONFIG: working directory
        // CORES: config.cpuCores
        // PRIORITY: config.priority
    // process.setCoresUsed(config.cpuCores) - I get the impression limiting the CPU cores has been a problem since it's set everywhere
    // os.exec(process, env vars)
    args := strings.Fields(cmd)
    renderCmd := exec.Command(path.Join(j.Renderer.GetContentPath(), hardware.RendererPath()), args...)
    renderCmd.Env = append(os.Environ(), 
        fmt.Sprintf("BLENDER_USER_CONFIG=%s", j.RootPath),
        fmt.Sprintf("CORES=%d", useCores),
        fmt.Sprintf("PRIORITY=%d", config.Priority))
    
    // read Stdin from process
    // output status, plus read line for blender error (see Job.detectError for all the string variations), returns (and deletes script file) if error
    renderOut, err := renderCmd.StdoutPipe()
    if err != nil {
        return err
    }

    if err := renderCmd.Start(); err != nil {
        return err
    }

    peakMemory := 0
    scanner := bufio.NewScanner(renderOut)
    re := regexp.MustCompile(`peak\s(\d+\.\d+)([bkmgtpe])`)
    for scanner.Scan() {
        res := re.FindAllStringSubmatch(strings.ToLower(scanner.Text()), -1)
        if len(res) > 0 {
            size, _ := strconv.ParseFloat(res[0][1], 64)
            memory := int(math.Pow(1024, float64(strings.Index("bkmgtpe", res[0][2]))) * size)
            inKb := memory / 1024
            if inKb > peakMemory {
                peakMemory = inKb
                log.Printf("Peak Memory: %d KiB", peakMemory)
            }
        }
    }

    if err := renderCmd.Wait(); err != nil {
        return err
    }

    // find "$workingdir\$job.id_$job.frame*", if !exists, look for "$workingdir\$job.path.crash.txt" if present then blender crashed (+delete file)
    // delete scene dir
    // return image file path

    return nil
}

func (j *Job) Cancel() {
    fmt.Println("Job.Cancel()")
    // this.client.getRenderingJob().setServerBlockJob(true);
    // OS.getOS().kill(this.client.getRenderingJob().getProcessRender().getProcess());
    // this.client.getRenderingJob().setAskForRendererKill(true);
}