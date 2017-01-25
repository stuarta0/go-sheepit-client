// +build

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"

	"github.com/stuarta0/go-sheepit-client/client"
	"github.com/stuarta0/go-sheepit-client/common"
	su "github.com/stuarta0/go-sheepit-client/stringutils"
)

func main() {
	config := common.Configuration{Server:"example"} // always overidden :(
	flag.StringVar(&config.Server, "server", "", "Render-farm server")
	flag.StringVar(&config.Login, "login", "", "User's login")
	flag.StringVar(&config.Password, "password", "", "User's password")
	cacheDirPtr := flag.String("cache-dir", "", "Cache/Working directory. Caution, everything in it not related to the render-farm will be removed")
	flag.IntVar(&config.MaxUploadingJob, "max-uploading-job", -1, "")
	flag.StringVar(&config.Gpu, "gpu", "", "CUDA name of the GPU used for the render, for example CUDA_0")
	computeMethodPtr := flag.String("compute-method", "", "CPU: only use cpu, GPU: only use gpu, CPU_GPU: can use cpu and gpu (not at the same time) if -gpu is not use it will not use the gpu")
	flag.IntVar(&config.UseCores, "cores", -1, "Number of cores/threads to use for the render")
	//flag.StringVar(&config., "request-time", "", "H1:M1-H2:M2,H3:M3-H4:M4 Use the 24h format. For example to request job between 2am-8.30am and 5pm-11pm you should do --request-time 2:00-8:30,17:00-23:00 Caution, it's the requesting job time to get a project not the working time")	
	flag.StringVar(&config.Proxy, "proxy", "", "URL of the proxy")
	flag.StringVar(&config.Extras, "extras", "", "Extras data push on the authentication request")
	//flag.StringVar(&config.UiType, "ui", "text", "Specify the user interface to use, only 'text' allowed.")
	flag.IntVar(&config.Priority, "priority", 19, "Set render process priority (19 lowest to -19 highest)")

	//flag.BoolVar(&config., "verbose", false, "Display log") // --verbose, print_log
	//flag.BoolVar(&config., "version", false, "Display application version") // --version, versionHandler
	//flag.BoolVar(&config., "show-gpu", false, "Print available CUDA devices and exit") // --show-gpu, listGpuParameterHandler
	//flag.BoolVar(&config., "no-systray", false, "Don't use systray, always false")
	configPathPtr := flag.String("config", "", "Specify the configuration file")

	// extra command line args for go-sheepit
	flag.StringVar(&config.ProjectDir, "project-dir", "", "Cache directory for project files. Caution, everything in it not related to the render-farm will be removed")
	flag.StringVar(&config.StorageDir, "storage-dir", "", "Cache directory for renderers. Caution, everything in it not related to the render-farm will be removed")

	flag.Parse()

	// use string to identify compute method
	config.SetComputeMethod(*computeMethodPtr)

	if cacheDirPtr != nil {
		config.ProjectDir = *cacheDirPtr;
		config.StorageDir = *cacheDirPtr;
	}

	// if we have a config file, use it's values for those that weren't provided
	// NOTE: the java client config file is incompatible - it needs to be reformatted to valid TOML (i.e. quoted values for strings)
	if _, err := os.Stat(*configPathPtr); err != nil {
		fmt.Printf("Config file \"%s\" does not exist\n", *configPathPtr)
	} else {
		config2 := common.Configuration{}
		if _, err := toml.DecodeFile(*configPathPtr, &config2); err != nil {
			fmt.Printf("Unable to read config \"%s\": %s\n", *configPathPtr, err)
		} else {

			if !su.IsEmpty(config2.ProjectDir) {
				config2.StorageDir = config2.ProjectDir
			}

			config.Merge(config2)
		}
	}

	config.SetDefaults()
	fmt.Println("Running client with the following configuration (password omitted):")
	Dump(config)

	// manage shutdown/exit calls
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		count := 0
		for {
			sig := <-sigs
			if sig.String() == "interrupt" && count == 0 {
				fmt.Println("Will exit after current frame. Ctrl+C to exit now.")
				count++
			} else if count >= 1 {
				done <- true
			}
		}
	}()

	// run client to manage rendering jobs requested from server
	client := client.Client{Configuration:config}
	if e := client.Run(); e != nil {
		panic(fmt.Sprintf("Client.Run() raised error: %s", e))
	}

	// wait for exit signal
	//<-done
}

func Dump(obj interface{}) {
	if b, err := json.Marshal(obj); err == nil {
		var out bytes.Buffer
		json.Indent(&out, b, "", "\t")
		out.WriteTo(os.Stdout)
		fmt.Println()
	}
}
