package api

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/stuarta0/go-sheepit-client/common"
	"github.com/stuarta0/go-sheepit-client/hardware"
	su "github.com/stuarta0/go-sheepit-client/stringutils"
)

type endpoint struct {
	Location string
	Timeout int
}

type xmlRequest struct {
	Type string `xml:"type,attr"`
	Path string `xml:"path,attr"`
	MaxPeriod int `xml:"max-period,attr"`
}

type xmlConfig struct {
	Status int `xml:"status,attr"`
	Requests []xmlRequest `xml:"request"`
}


type xmlStats struct {
	CreditsSession int    `xml:"credits_session,attr"`
	CreditsTotal int      `xml:"credits_total,attr"`
	FramesRemaining int   `xml:"frame_remaining,attr"`
	WaitingProjects int   `xml:"waiting_project,attr"`
	ConnectedMachines int `xml:"connected_machine,attr"`
}

type xmlRenderer struct {
	Md5 string          `xml:"md5,attr"`
	Command string      `xml:"commandline,attr"`
	UpdateMethod string `xml:"update_method,attr"`
}

type xmlJob struct {
	Id int                 `xml:"id,attr"`
	UseGpu bool            `xml:"use_gpu,attr"`
	ArchiveMd5 string      `xml:"archive_md5,attr"`
	Path string            `xml:"path,attr"`
	Frame int              `xml:"frame,attr"`
	SynchronousUpload bool `xml:"synchronous_upload,attr"`
	Extras string          `xml:"extras,attr"`
	Name string            `xml:"name,attr"`
	Password string        `xml:"password,attr"`

	Renderer xmlRenderer   `xml:"renderer"`
	Script string          `xml:"script"`
}

type xmlJobRequest struct {
	Status int `xml:"status,attr"`
	Stats xmlStats `xml:"stats"`
	Job xmlJob `xml:"job"`
}

type xmlKeepalive struct {
	Status int `xml:"status,attr"`
}

type Api struct {
	Server string
	client http.Client
	endpoints map[string]endpoint
	//lastRequest time
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
func New(c common.Configuration) (*Api, error) {
	api := Api{Server:c.Server}
	if jar, err := cookiejar.New(nil); err == nil {
		api.client = http.Client{Jar: jar}
	} else {
		return nil, errors.New("GetEndpoints couldn't store cookies")
	}

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

	url := fmt.Sprintf("%s/server/config.php?%s", api.Server, v.Encode())
	resp, err := api.client.Get(url)
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
    m := make(map[string]endpoint)
    for _, r := range xmlC.Requests {
    	req := endpoint{Location:r.Path, Timeout:r.MaxPeriod}
    	m[r.Type] = req
    }
    api.endpoints = m
    return &api, nil
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
func (api *Api) RequestJob(c common.Configuration) (*common.Job, error) {
	v := url.Values{}
	v.Set("computemethod", fmt.Sprintf("%d", c.ComputeMethod))
	if c.UseCores > 0  {
		v.Set("cpu_cores", fmt.Sprintf("%d", c.UseCores))
	} else {
		cpu := hardware.CpuStat()
		v.Set("cpu_cores", fmt.Sprintf("%d", cpu.TotalCores))
	}

	url := fmt.Sprintf("%s/%s?%s", api.Server, api.endpoints["request-job"].Location, v.Encode())
	resp, err := api.client.Get(url)
	if err != nil {
		fmt.Println("Request failed")
		return nil, err
	}
	defer resp.Body.Close()
    decoder := xml.NewDecoder(resp.Body)

    var xmlJ xmlJobRequest
    if err := decoder.Decode(&xmlJ); err != nil {
    	fmt.Println("Decode failed")
    	return nil, err
    }

    if xmlJ.Status != 0 {
    	return nil, errors.New(fmt.Sprintf("SheepIt Server Error Code %d", xmlJ.Status)) // errors.New(common.ErrorAsString(common.ServerCodeToError(xmlJ.Status)))
    }

    // TODO: massage data
    fmt.Printf("%+v\n", xmlJ)
    return nil, errors.New("RequestJob TBA")
}

func (api *Api) SendKeepalive(job *common.Job) error {

	// TODO: get values for job in a thread locking context here

	v := url.Values{}
	v.Set("job", fmt.Sprintf("%d", job.Id))
	v.Set("frame", fmt.Sprintf("%d", job.Frame))
	if !su.IsEmpty("") {
		v.Set("extras", job.Extras)
	}
	// TODO
	// if job.Renderer != nil {
	// 	v.Set("rendertime", job.Renderer.ElapsedDuration)
	// 	v.Set("remainingtime", job.Renderer.RemainingDuration)
	// }

	url := fmt.Sprintf("%s/%s?%s", api.Server, api.endpoints["keepmealive"].Location, v.Encode())
	fmt.Println("Requesting:", url)
	resp, err := api.client.Get(url)
	if err != nil {
		fmt.Println("Request failed")
		return err
	}
	defer resp.Body.Close()
    decoder := xml.NewDecoder(resp.Body)

    var xmlK xmlKeepalive
    if err := decoder.Decode(&xmlK); err != nil {
    	fmt.Println("Decode failed")
    	return err
    }

    if xmlK.Status == common.KEEPMEALIVE_STOP_RENDERING {
    	log.Println("Server::keeepmealive server asked to kill local render process")
    	job.Cancel()
    }

    return nil
}