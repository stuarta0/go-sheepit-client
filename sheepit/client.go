package sheepit

import (
	"fmt"
	"errors"

	"github.com/stuarta0/go-sheepit-client/common"
	"github.com/stuarta0/go-sheepit-client/storage"
	"github.com/stuarta0/go-sheepit-client/hardware"
	"github.com/stuarta0/go-sheepit-client/api"
)


type Client struct {
	Configuration common.Configuration
}

func (c *Client) Run() error {
	fmt.Println("Client.Run()")
	fmt.Printf("%+v\n", c.Configuration)

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
	endpoints, err := api.GetEndpoints(c.Configuration)
	if err != nil {
		return err
	}

	// server.start() - server class inherits from Thread, calls run() which calls stayAlive() which loops indefinitely sleeping every minute until keepalive exceeded, then stats are sent and server can request current job be terminated
	// starts anonymous func as Thread to continually check for finished job to send

	//  
	// loop starts here (1 loop = 1 frame rendered for a job)
	//
	// some loop guff that's probably important (checking whether to get next render or hold off)
	// server.requestJob() - 
		// send request to config['request-job'] with some more params for stats (assume this is to choose the right job for hardware)
		// look up error code from jobrequest.prop['status'], if != 0, error (see Errors for full list of server error codes)
		// get stats and ensure all required attributes are present for job/renderer
		// return new Job
	err = api.RequestJob(c.Configuration, endpoints["request-job"].Location)
	if err != nil {
		return err
	}

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


// /server/config.php
//
// <?xml version="1.0" encoding="utf-8" ?>
// <config status="0">
//     <request type="validate-job" path="/server/send_frame.php" />
//     <request type="request-job" path="/server/request_job.php" />
//     <request type="download-archive" path="/server/archive.php" />
//     <request type="error" path="/server/error.php" />
//     <request type="keepmealive" path="/server/keepmealive.php" max-period="800" />
//     <request type="logout" path="/account.php?mode=logout&amp;worker=1" />
//     <request type="last-render-frame" path="/ajax.php?action=webclient_get_last_render_frame_ui&amp;type=raw" />
// </config>


// /server/request_job.php
//
// FAILURE
// <?xml version="1.0" encoding="utf-8"?>
// <jobrequest status="205"/>
//
// SUCCESS
// <?xml version="1.0" encoding="utf-8" ?>
// <jobrequest status="0">
//     <stats credits_session="0" credits_total="619296" frame_remaining="39752" waiting_project="50" connected_machine="391"/>
//     <job id="1" use_gpu="1" archive_md5="fed2b5d02774c785d31c121a7c9ae217" path="compute-method.blend" frame="0340" synchronous_upload="1" extras="" name="computer_check" password="">
//         <renderer md5="fc6ecd3558678b844c8dac88428bf15e" commandline=".e --factory-startup --disable-autoexec -b .c -o .o -f .f -x 1" update_method="remainingtime"/>
//         <script>import bpy

// # if it's a movie clip, switch to png
// fileformat = bpy.context.scene.render.image_settings.file_format
// if fileformat != 'BMP' and fileformat != 'PNG' and fileformat != 'JPEG' and fileformat != 'TARGA' and fileformat != 'TARGA_RAW' :
// 	bpy.context.scene.render.image_settings.file_format = 'PNG'
// 	#bpy.context.scene.render.file_extension = '.png'
// 	bpy.context.scene.render.filepath = ''

// </script>
//     </job>
// </jobrequest>
