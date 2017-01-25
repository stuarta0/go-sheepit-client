package common

import (
	//"errors"
	//"fmt"
	"time"
)

type Renderer struct {
	Md5 string          `xml:"md5,attr"`
	Command string      `xml:"commandline,attr"`
	UpdateMethod string `xml:"update_method,attr"`

	ElapsedDuration time.Time
	RemainingDuration time.Time
}
