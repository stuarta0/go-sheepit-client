package common

import (
    //"errors"
    "fmt"
    "io/ioutil"
    "path"

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

func (j *Job) Render(device hardware.Computer) error {
    fmt.Println("Job.Render()")

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
        "bpy.context.scene.render.tile_x = %[4]d\n" + 
        "bpy.context.scene.render.tile_y = %[4]d\n", 
        j.Script, device.GetComputeDeviceType(), device.GetDeviceName(), device.GetComputeDeviceName(), device.GetOptimalTileSize())

    if err := ioutil.WriteFile(path.Join(j.GetContentPath(), "script.py"), ([]byte)(script), 0755); err != nil {
        return err
    }

    // command = job['renderer.commandline']
    // replace in command string:
        // ".c": "$scenepath -P $scriptpath", where job['script'] has been written to "working directory\script_<randint>" (no extension), defer delete file until render complete (i.e. job.render exits)
        // ".e": "$rendererpath" + "-t $cpucores" if cpucores specified by user (default all cores)
        // ".o": "$workingdir\$job.id_" (i.e. frame render path; blender will add frame number and extension)
        // ".f": "$job.frame"
    // set env vars:
        // BLENDER_USER_CONFIG: working directory
        // CORES: config.cpuCores
        // PRIORITY: config.priority
    // process.setCoresUsed(config.cpuCores) - I get the impression limiting the CPU cores has been a problem since it's set everywhere
    // os.exec(process, env vars)
    // read Stdin from process
    // output status, plus read line for blender error (see Job.detectError for all the string variations), returns (and deletes script file) if error
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