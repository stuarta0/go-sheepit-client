package common

import (
	"fmt"
	"errors"
	"time"
	"io/ioutil"
	"os"
	"os/user"
	"log"
	"path"

	su "github.com/stuarta0/go-sheepit-client/stringutils"
)

type ComputeType int

// ComputeType enum
// https://golang.org/ref/spec#Iota
const (
	COMPUTE_CPU_GPU = iota 	// 0
	COMPUTE_CPU 			// 1
	COMPUTE_GPU 			// 2
)

func (d *ComputeType) UnmarshalText(text []byte) error {
    var err error
    s := string(text)
    if s == "CPU" {
    	*d = COMPUTE_CPU
    } else if s == "GPU" {
    	*d = COMPUTE_GPU
    } else if s == "CPU_GPU" {
    	*d = COMPUTE_CPU_GPU
    } else {
    	err = errors.New(fmt.Sprintf("Compute type %s not recognised.", s))
    }
    return err
}

type Configuration struct {
	Server string
	Proxy string
	Login string
	Password string `json:"-"`

	// Used for initialisation only
	CacheDir string `toml:"cache-dir"`

	// Directory containing downloaded project files. Default /tmp
	ProjectDir string `toml:"project-dir"`

	// Directory containing downloaded renderers. Default ~/.sheepit/
	StorageDir string `toml:"storage-dir"`

	MaxUploadingJob int
	Gpu string `toml:"compute-gpu"` // TODO: calculate GPU device from interrogation of CUDA lib; OS-specific
	ComputeMethod ComputeType `toml:"compute-method"`

	// Total cores to use when rendering. Default all available cores.
	UseCores int `toml:"cores"`
	
	// Times during which SheepIt will request new jobs
	RequestTime []time.Time

	// ??
	Extras string

	//UiType string // always text

	// Process priority for rendering
	Priority int

	TileSize int `toml:"tile-size"`
}

// Given a string representation of the COMPUTE_ consts, convert to value
// For example, "CPU" -> const COMPUTE_CPU -> 1
func (c *Configuration) SetComputeMethod(method string) {
	c.ComputeMethod.UnmarshalText([]byte(method))
}

// Take a string such as "CUDA_0" and determine actual hardware capabilities from CUDA lib
func (c *Configuration) SetGpuDevice(gpu string) {
	// TODO
}

// After setting up Configuration struct, call SetDefaults() to ensure remaining defaults are correctly configured (e.g. working directories)
func (c *Configuration) SetDefaults() {
	if su.IsEmpty(c.Server) {
		c.Server = "https://client.sheepit-renderfarm.com"
	}

	if !su.IsEmpty(c.CacheDir) {
		if su.IsEmpty(c.ProjectDir) {
			c.ProjectDir = c.CacheDir
		}
		if su.IsEmpty(c.StorageDir) {
			c.StorageDir = c.CacheDir
		}
	}

	if su.IsEmpty(c.ProjectDir) {
		if dir, err := ioutil.TempDir("", "farm_"); err != nil {
			log.Fatal("Could not create temporary directory for Configuration.ProjectDir")
		} else {
			c.ProjectDir = dir
		}
	}
	if su.IsEmpty(c.StorageDir) {
		if usr, err := user.Current(); err != nil {
			// if we can't determine the current user, we must reuse the ProjectDir for the renderers
			c.StorageDir = c.ProjectDir
		} else {
			if _, err := os.Stat(usr.HomeDir); err != nil {
				// if this user doesn't have a home directory, we must reuse the ProjectDir for the renderers
				c.StorageDir = c.ProjectDir
			} else {
				c.StorageDir = path.Join(usr.HomeDir, ".sheepit", "storage")
			}
		}
	}
}

func (c *Configuration) Merge(other Configuration) {
	// Merge will bring in values from [other] when this value is null or default
	if (su.IsEmpty(c.Server) && !su.IsEmpty(other.Server)) { c.Server = other.Server }
	if (su.IsEmpty(c.Login) && !su.IsEmpty(other.Login)) { c.Login = other.Login }
	if (su.IsEmpty(c.Password) && !su.IsEmpty(other.Password)) { c.Password = other.Password }
	if (su.IsEmpty(c.CacheDir) && !su.IsEmpty(other.CacheDir)) { c.CacheDir = other.CacheDir }
	if (su.IsEmpty(c.ProjectDir) && !su.IsEmpty(other.ProjectDir)) { c.ProjectDir = other.ProjectDir }
	if (su.IsEmpty(c.StorageDir) && !su.IsEmpty(other.StorageDir)) { c.StorageDir = other.StorageDir }
	if (c.MaxUploadingJob < 0 && other.MaxUploadingJob > 0) { c.MaxUploadingJob = other.MaxUploadingJob }
	if (su.IsEmpty(c.Gpu) && !su.IsEmpty(other.Gpu)) { c.Gpu = other.Gpu }
	if (c.ComputeMethod != other.ComputeMethod) { c.ComputeMethod = other.ComputeMethod }
	if (c.UseCores < 0 && other.UseCores > 0) { c.UseCores = other.UseCores }
	// requestTime
	if (su.IsEmpty(c.Proxy) && !su.IsEmpty(other.Proxy)) { c.Proxy = other.Proxy }
	if (su.IsEmpty(c.Extras) && !su.IsEmpty(other.Extras)) { c.Extras = other.Extras }
	// uiType
	if (c.Priority == 0 && other.Priority != 0) { c.Priority = other.Priority }
	if (c.TileSize < 0 && other.TileSize > 0) { c.TileSize = other.TileSize }
}