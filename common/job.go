package common

import (
    //"errors"
    "fmt"
    "path"
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


func (j *Job) Render() {
    fmt.Println("Job.Render()")
}

func (j *Job) Cancel() {
    fmt.Println("Job.Cancel()")
    // this.client.getRenderingJob().setServerBlockJob(true);
    // OS.getOS().kill(this.client.getRenderingJob().getProcessRender().getProcess());
    // this.client.getRenderingJob().setAskForRendererKill(true);
}