package api

import (
    "crypto/md5"
    "encoding/xml"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "reflect"
    "time"

    "github.com/stuarta0/go-sheepit-client/common"
    "github.com/stuarta0/go-sheepit-client/hardware"
    su "github.com/stuarta0/go-sheepit-client/stringutils"
)

type endpoint struct {
    Location string
    Timeout time.Duration
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

type xmlJobRequest struct {
    Status int `xml:"status,attr"`
    Stats xmlStats `xml:"stats"`
    Job common.Job `xml:"job"`
}

type xmlKeepalive struct {
    Status int `xml:"status,attr"`
}

type Api struct {
    Server string
    client http.Client
    endpoints map[string]endpoint
    lastRequest time.Time
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
    api.lastRequest = time.Now()
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
        req := endpoint{Location:r.Path, Timeout:time.Duration(r.MaxPeriod) * time.Second}
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
//  bpy.context.scene.render.image_settings.file_format = 'PNG'
//  #bpy.context.scene.render.file_extension = '.png'
//  bpy.context.scene.render.filepath = ''

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
    api.lastRequest = time.Now()
    decoder := xml.NewDecoder(resp.Body)

    var xmlJ xmlJobRequest
    if err := decoder.Decode(&xmlJ); err != nil {
        fmt.Println("Decode failed")
        return nil, err
    }

    if xmlJ.Status != 0 {
        return nil, errors.New(fmt.Sprintf("SheepIt Server Error Code %d", xmlJ.Status)) // errors.New(common.ErrorAsString(common.ServerCodeToError(xmlJ.Status)))
    }
    
    return &xmlJ.Job, nil
}

// A function to call periodically with details on current progress. Also doubles as a session keepalive.
// Returns duration until next report is required.
func (api *Api) SendHeartbeat(job *common.Job) (time.Duration, error) {

    // TODO: get values for job in a thread locking context here
    endpoint := api.endpoints["keepmealive"]
    safeTimeout := endpoint.Timeout * 4 / 5 // report within 80% of original timeout

    // if within 25% of timeout expiring, send keepalive
    if time.Since(api.lastRequest) >= endpoint.Timeout * 3 / 4 {
        v := url.Values{}
        v.Set("job", fmt.Sprintf("%d", job.Id))
        v.Set("frame", fmt.Sprintf("%d", job.Frame))
        if !su.IsEmpty("") {
            v.Set("extras", job.Extras)
        }
        if job.Renderer != nil {
            v.Set("rendertime", fmt.Sprintf("%d", job.Renderer.ElapsedDuration.Seconds()))
            v.Set("remainingtime", fmt.Sprintf("%d", job.Renderer.RemainingDuration.Seconds()))
        }

        url := fmt.Sprintf("%s/%s?%s", api.Server, endpoint.Location, v.Encode())
        resp, err := api.client.Get(url)
        if err != nil {
            // report in half the time remaining
            // e.g. if keepalive timeout is 15 minutes, 
            //    the first call will be at 75% or 11:15, 
            //    then on failure, 50% of remaining time after that: 13:07, 14:04, 14:32 until duration < 20sec
            nextReport := (endpoint.Timeout - time.Since(api.lastRequest)) / 2
            if nextReport >= time.Second * 20 {
                return nextReport, err
            } else {
                return safeTimeout, err
            }
        }
        defer resp.Body.Close()
        api.lastRequest = time.Now()
        decoder := xml.NewDecoder(resp.Body)

        var xmlK xmlKeepalive
        if err := decoder.Decode(&xmlK); err != nil {
            return safeTimeout, err
        }

        if xmlK.Status == common.KEEPMEALIVE_STOP_RENDERING {
            log.Println("Server::keeepmealive server asked to kill local render process")
            job.Cancel()
        }
    } 

    return safeTimeout, nil
}

func (api *Api) DownloadArchive(job *common.Job, archive common.FileArchive) error {

    // create target file
    out, err := os.Create(archive.GetArchivePath())
    if err != nil {
        return err
    }
    defer out.Close()

    // type assertion for GET params
    typeId := "job"
    if reflect.TypeOf(archive) == reflect.TypeOf((*common.Renderer)(nil)) { //_, ok := archive.(common.Renderer); ok {
        typeId = "binary"
    }

    v := url.Values{}
    v.Set("job", fmt.Sprintf("%d", job.Id))
    v.Set("type", typeId)
    url := fmt.Sprintf("%s/%s?%s", api.Server, api.endpoints["download-archive"].Location, v.Encode())
    resp, err := api.client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // copy max 32kb at a time
    if _, err = io.Copy(out, resp.Body); err != nil {
        return err
    }

    // verify MD5
    if contents, readErr := ioutil.ReadFile(archive.GetArchivePath()); readErr == nil {
        if hash := fmt.Sprintf("%x", md5.Sum(contents)); hash != archive.GetExpectedHash() {
            os.Remove(archive.GetArchivePath())
            return errors.New(fmt.Sprintf("Downloaded archive for %s %s does not match actual hash %s", typeId, archive.GetExpectedHash(), hash))
        } else {
            log.Printf("Downloaded %s %s successfully", typeId, archive.GetExpectedHash())
        }
    }

    return nil
}

func (api *Api) UploadResult(job *common.Job) error {

    // resp, err := http.PostForm("http://example.com/form",
    // url.Values{"key": {"Value"}, "id": {"123"}})

    return nil
}