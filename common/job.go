package common

import (
	//"errors"
	"fmt"
)

type Job struct {
	Id int                 `xml:"id,attr"`
	UseGpu bool            `xml:"use_gpu,attr"`
	ArchiveMd5 string      `xml:"archive_md5,attr"`
	Path string            `xml:"path,attr"`
	Frame int              `xml:"frame,attr"`
	SynchronousUpload bool `xml:"synchronous_upload,attr"`
	Extras string          `xml:"extras,attr"`
	Name string            `xml:"name,attr"`
	Password string        `xml:"password,attr"`

	//Renderer xmlRenderer   `xml:"renderer"`
	Script string          `xml:"script"`
}

func (j *Job) Render() {
	fmt.Println("Job.Render()")
}