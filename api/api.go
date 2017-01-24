package api

import (
	"fmt"
	"errors"
	//"net/http"
	"net/url"
	"os"
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

func GetServerConfiguration(c common.Configuration) (map[string]Endpoint, error) {
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
	fmt.Println("Requesting:", url)
	// resp, err := http.Get(url)
	// if err != nil {
	// 	fmt.Println("GET error:")
	// 	fmt.Println(err)
	// }
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }

	f, _ := os.Open(`C:\Users\Stuart\.sheepit\config_response.xml`)
	defer f.Close()
    decoder := xml.NewDecoder(f)

    // for small XML, use whole decode
    var xmlC xmlConfig
    if err := decoder.Decode(&xmlC); err == nil {
    	fmt.Printf("%+v\n", xmlC)
    } else {
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

func RequestJob(c common.Configuration, endpoint string) error {
	fmt.Printf("Request job: %s/%s\n", c.Server, endpoint)
	return nil
}




// // Another xml method
// b, _ := ioutil.ReadAll(f)
// var xmlC xmlConfig
// xml.Unmarshal(b, &xmlC)
// fmt.Printf("%+v\n", xmlC)


// // for reading large XML, use streaming method
// // https://www.goinggo.net/2013/06/reading-xml-documents-in-go.html
// for {
// 	t, _ := decoder.Token(); 
// 	if t == nil { break }

// 	switch se := t.(type) {
// 	case xml.StartElement:
// 		if se.Name.Local == "config" {
// 			var c xmlConfig
// 			decoder.DecodeElement(&c, &se)
			// fmt.Printf("%+v\n", c)
// 		}
// 	}
// }