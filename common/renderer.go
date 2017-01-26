package common

import (
    //"errors"
    "fmt"
    "path"
    "time"
)

type Renderer struct {
    ArchiveMd5 string   `xml:"md5,attr"`
    Command string      `xml:"commandline,attr"`
    UpdateMethod string `xml:"update_method,attr"`

    ElapsedDuration time.Duration
    RemainingDuration time.Duration

    RootPath string
}

// FileArchive interface
func (r Renderer) GetExpectedHash() string {
	return r.ArchiveMd5
}
func (r Renderer) GetArchivePath() string {
    return path.Join(r.RootPath, fmt.Sprintf("%s.zip", r.ArchiveMd5))
}
func (r Renderer) GetContentPath() string {
    return path.Join(r.RootPath, r.ArchiveMd5)
}