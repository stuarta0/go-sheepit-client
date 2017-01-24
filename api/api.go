package api

import (
	"fmt"
	"errors"
	"net/http"
	"net/url"
	"encoding/xml"

	"github.com/stuarta0/go-sheepit-client/common"
	"github.com/stuarta0/go-sheepit-client/hardware"
)


type xmlRequest struct {
	Type string `xml:"type,attr"`
	Path string `xml:"path,attr"`
	MaxPeriod int `xml:"max-period,attr"`
}

type xmlConfig struct {
	Status int `xml:"status,attr"`
	Requests []xmlRequest `xml:"request"`
}

type Endpoint struct {
	Location string
	MaxPeriod int
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
func GetEndpoints(c common.Configuration) (map[string]Endpoint, error) {
	cpu := hardware.CpuStat()

	v := url.Values{}
	v.Set("login", c.Login)
	v.Set("password", c.Password)
	v.Set("cpu_family", cpu.Family)
	v.Set("cpu_model", cpu.Model)
	v.Set("cpu_model_name", cpu.Name)
	v.Set("os", hardware.PlatformName())
	v.Set("ram", fmt.Sprintf("%d", hardware.TotalMemory()))
	v.Set("bits", cpu.Architecture)
	v.Set("version", "5.290.2718")
	v.Set("hostname", "stuarta0-skylake")
	v.Set("extras", c.Extras)
	if c.UseCores > 0  {
		v.Set("cpu_cores", fmt.Sprintf("%d", c.UseCores))
	} else {
		v.Set("cpu_cores", fmt.Sprintf("%d", cpu.TotalCores))
	}

	url := fmt.Sprintf("%s/server/config.php?%s", c.Server, v.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
    decoder := xml.NewDecoder(resp.Body)

    var xmlC xmlConfig
    if err := decoder.Decode(&xmlC); err != nil {
    	return nil, err
    }

    if xmlC.Status != 0 {
    	return nil, errors.New(common.ErrorAsString(common.ServerCodeToError(xmlC.Status)))
    }

    // convert XML representation to simpler data structure
    m := make(map[string]Endpoint)
    for _, r := range xmlC.Requests {
    	req := Endpoint{Location:r.Path, MaxPeriod:r.MaxPeriod}
    	m[r.Type] = req
    }
    return m, nil
}

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
func RequestJob(c common.Configuration, endpoint string) error {
	fmt.Printf("Request job: %s/%s\n", c.Server, endpoint)
	return nil
}