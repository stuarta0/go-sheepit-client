package client

import (
    "errors"
    //"fmt"
    "log"
    "os"
    //"path"
    "time"

    "github.com/stuarta0/go-sheepit-client/common"
    "github.com/stuarta0/go-sheepit-client/storage"
    "github.com/stuarta0/go-sheepit-client/hardware"
    "github.com/stuarta0/go-sheepit-client/api"
)


type Client struct {
    Configuration common.Configuration
}

func (c *Client) Run() error {
    if e := createDirectories(&c.Configuration); e != nil {
        return e
    }

    // if !OS supported, panic
    // if !CPU supported, panic
    cpu := hardware.CpuStat()
    if !cpu.IsValid() {
        return errors.New("CPU not supported.")
    }

    // clean working/storage directory (delete all directories and any files that are ZIPs with filenames that don't match their MD5 hash)

    // server.getConfiguration() -
        // get hostname
        // <base_url>/server/config.php?... every single value for computer stats, auth, etc in GET params :/
        // http://blog.httpwatch.com/2009/02/20/how-secure-are-query-strings-over-https/ and https://blog.codinghorror.com/youre-probably-storing-passwords-incorrectly/
        // get response which will be content-type: text/xml (see below for structure)
        // store all the key/value pairs, and make keepalive = (int(max-period) - 120) * 1000 // 2mins of safety net apparently; *1000 is probably to convert to milliseconds for a timer
    server, err1 := api.New(c.Configuration)
    if err1 != nil {
        return err1
    }

    // server.start() - server class inherits from Thread, calls run() which calls stayAlive() which loops indefinitely sleeping every minute until keepalive exceeded, then stats are sent and server can request current job be terminated
    // starts anonymous func as Thread to continually check for finished job to send
    var job common.Job
    go func() {
        for {
            timeout, _ := server.ReportProgress(&job)
            // report progress will let us know when it needs to be called again
            time.Sleep(timeout)
        }
    }()

    //  
    // loop starts here (1 loop = 1 frame rendered for a job)
    //
    // some loop guff that's probably important (checking whether to get next render or hold off)
    // server.requestJob() - 
        // send request to config['request-job'] with some more params for stats (assume this is to choose the right job for hardware)
        // look up error code from jobrequest.prop['status'], if != 0, error (see Errors for full list of server error codes)
        // get stats and ensure all required attributes are present for job/renderer
        // return new Job
    newJob, err2 := server.RequestJob(c.Configuration)
    if err2 != nil {
        return err2
    }
    job = *newJob
    job.RootPath = c.Configuration.ProjectDir
    job.Renderer.RootPath = c.Configuration.StorageDir

    // lots of exception handling for various states, if job null then sleep 15 minutes
    // now work(job)
    // download renderer from config['download-archive']?type=binary&job=<job.id> 
        // to storage directory\rendererMD5.zip if ZIP doesn't already exist (+MD5 check after download), extract to working directory\rendererMD5\<os binary path> if rendererMD5 directory doesn't exist (set exec flag on binary)
            // os "windows": "rend.exe"
            // os "mac": "Blender\blender.app\Contents\MacOS\blender"
            // os "linux": "rend.exe"
            // os "freebsd": "rend.exe"
    // download scene from config['download-archive']?type=job&job=<job.id> 
        // to working directory\sceneMD5.zip if ZIP doesn't already exist (+MD5 check after download), extract to working directory\sceneMD5\job['path'] if sceneMD5 directory doesn't exist
    if _, err := os.Stat(job.Renderer.GetArchivePath()); err != nil {
        log.Println("Downloading renderer", job.Renderer.ArchiveMd5)
        server.DownloadArchive(&job, job.Renderer)
    } else {
        log.Println("Reusing cached renderer", job.Renderer.ArchiveMd5)
    }

    if _, err := os.Stat(job.GetArchivePath()); err != nil {
        log.Println("Downloading project", job.ArchiveMd5)
        server.DownloadArchive(&job, job)
    } else {
        log.Println("Reusing cached project", job.ArchiveMd5)
    }



    // job.render() -
        // String core_script = "import bpy\n" + "bpy.context.user_preferences.system.compute_device_type = \"%s\"\n" + "bpy.context.scene.cycles.device = \"%s\"\n" + "bpy.context.user_preferences.system.compute_device = \"%s\"\n";
        // if using GPU and has GPU: core_script % ("CUDA", "GPU", gpu.CudaName())
        // else: core_script % ("NONE", "CPU", "CPU")
        // core_script += String.format("bpy.context.scene.render.tile_x = %1$d\nbpy.context.scene.render.tile_y = %1$d\n", getTileSize());
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
    // if !simultaneous upload, POST with content-type: multipart/form-data;boundary=***232404jkg4220957934FW**
        // write: --***232404jkg4220957934FW**\r\n
        // write: Content-Disposition: form-data; name="file"; filename="$imagepath"\r\n
        // write: \r\n
        // write: image file contents
    // if success, delete file, else retry send every 32s
    // if simulateous, add job to queue (which anonymous Thread at the start will handle)
    // sleep for 4s before next job, then another 2.3s for send frame
    // loop for next job

    return nil
}

func createDirectories(config *common.Configuration) error {
    if err := storage.CreateWorkingDirectory(config.ProjectDir); err != nil { return err }
    return storage.CreateWorkingDirectory(config.StorageDir)
}
