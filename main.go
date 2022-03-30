package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

const (
	bin       = "nvidia-smi"
	gpuArg    = "--id="
	queryArg  = "--query-gpu="
	formatArg = "--format=csv,noheader,nounits"
)

type Fan struct {
	Index  uint8   `json:"Index,omitempty"`
	GpuIDs []uint8 `json:"GpuIDs,omitempty"`
}
type Controller struct {
	Port  string  `json:"Port,omitempty"`
	Baud  int     `json:"Baud,omitempty"`
	Fans  []Fan   `json:"Fans,omitempty"`
	Temps []uint8 `json:"Temps,omitempty"`
	Dutys []uint8 `json:"Dutys,omitempty"`
}

var ports []string
var config []Controller

func init() {
	//	config.Controllers = make([]Controller, 4)
	var err error
	ports, err = serial.GetPortsList()
	if err != nil {
		log.Printf("error in init() %v", err)
	}
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("error in ReadFile() %v", err)
	}
	//fmt.Println(string(data))
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("error in Unmarshal() %v", err)
	}

}
func main() {
	fmt.Printf("%+v\n", config)
	fmt.Println(ports)
	//arduino config defaults are 8 data bits, no parity, one stop bit.
	var smode serial.Mode
	smode.BaudRate = config[0].Baud
	smode.DataBits = 8
	smode.Parity = serial.NoParity
	smode.StopBits = 1
	fmt.Println(config[0].Port, smode)
	//var sport serial.Port
	port, err := serial.Open(ports[2], &smode)
	fmt.Println(port)
	if err != nil {
		log.Panicf("error in Unmarshal() %v", err)

	}
	fc := createfancurve(config[0].Temps, config[0].Dutys)

	ptime := time.Now()
	tlist := make([]uint8, 4)
	for {

		if time.Since(ptime).Milliseconds() > 10 {
			ptime = time.Now()
			tlist = getgputemps(tlist)
			if tlist == nil {
				log.Println("Templist wasn't read will default duty cycle to 50%")
			} else {
				dlist := getdutylist(tlist, fc, &config[0])
				for i := range dlist {
					ibyte := uint8(i)
					fmt.Println(i)
					bees := []byte{ibyte, dlist[i], 0, 0}
					_, err := port.Write(bees)
					if err != nil {
						log.Printf("Error in port.Write:%v\n", err)
					}
				}
			}

		}

	}
	//strconv.Atoi()

}

func getdutylist(tlist []uint8, fc *fancurve, c *Controller) (dlist []uint8) {
	dlist = make([]uint8, len(c.Fans))
	for i := range c.Fans {
		for j := range c.Fans[i].GpuIDs {
			if dlist[i] < tlist[c.Fans[i].GpuIDs[j]] {
				dlist[i] = tlist[c.Fans[i].GpuIDs[j]]
			}
		}
	}
	return dlist
}

type fancurve struct {
	temp []uint8
	duty []uint8
}

func createfancurve(temp, duty []uint8) (fc *fancurve) {
	fc = new(fancurve)
	if len(temp) != len(duty) {
		return nil
	}
	fc.duty = make([]uint8, len(duty))
	fc.temp = make([]uint8, len(temp))
	copy(fc.duty, duty)
	copy(fc.temp, temp)
	return fc
}
func (fc *fancurve) update(temp, duty []uint8) {
	fc.duty = make([]uint8, len(duty))
	fc.temp = make([]uint8, len(temp))
	copy(fc.duty, duty)
	copy(fc.temp, temp)
}
func (fc *fancurve) tempmap(t uint8) (dc uint8) {
	if fc.temp[0] > t {
		return fc.duty[0]
	}
	for i := range fc.temp {
		if fc.temp[i] > t {
			return mapdata(t, fc.temp[i-1], fc.temp[i], fc.duty[i-1], fc.duty[i])
		}
	}
	return uint8(100)
}

func mapdata(in, inmin, inmax, outmin, outmax uint8) (out uint8) {
	return (in-inmin)*(outmax-outmin)/(inmax-inmin) + outmin

}

func getgputemps(temps []uint8) []uint8 {
	templist := nqueryall("temperature.gpu")
	for i := range templist {
		tint, err := strconv.Atoi(templist[i])
		if err != nil {
			return nil

		}
		temps[i] = uint8(tint)

	}
	return temps
}

//this will query all gpus and send it back as a slice of strings
func nqueryall(query string) []string {
	var out bytes.Buffer

	cmd := exec.Command(bin, queryArg+query, formatArg)
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatalf("nvsmi exec error: %v\n", err)

	}

	return strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")

}
