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

	"github.com/tarm/serial"
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
	Port  string `json:"Port,omitempty"`
	Baud  int    `json:"Baud,omitempty"`
	Fans  []Fan  `json:"Fans,omitempty"`
	Temps []uint `json:"Temps,omitempty"`
	Dutys []uint `json:"Dutys,omitempty"`
}

var port []serial.Port
var config []Controller

func init() {
	var err error

	if err != nil {
		log.Printf("error in init() %v", err)
	}
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("error in ReadFile() %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("error in Unmarshal() %v", err)
	}

}
func main() {
	var fan1rpm = "1000"
	var fan2rpm = "1000"
	//arduino config defaults are 8 data bits, no parity, one stop bit.
	var smode = &serial.Config{Name: config[0].Port, Baud: config[0].Baud, Parity: serial.ParityNone, StopBits: serial.Stop1}
	port, err := serial.OpenPort((smode))
	if err != nil {
		log.Printf("error in OpenPort() %v\n", err)
	}
	defer port.Close()
	fc := createfancurve(config[0].Temps, config[0].Dutys)
	ptime := time.Now()
	tlist := make([]uint, 12)
	for {

		if time.Since(ptime).Milliseconds() > 25 {
			fnrpm1, err := strconv.Atoi(fan1rpm)
			if err != nil {
				log.Printf("strconv.Atoi no feedback: %v\n", err)
			}
			fnrpm2, err := strconv.Atoi(fan2rpm)
			if err != nil {
				log.Printf("strconv.Atoi no feedback: %v\n", err)
			}
			if fnrpm1 == 0 || fnrpm2 == 0 {
				fmt.Println(fnrpm1, fnrpm2)
				log.Fatal("One of the Fans RPMS is 0")
			}
			ptime = time.Now()
			tlist = getgputemps(tlist)
			if tlist == nil {
				log.Println("Templist wasn't read will default duty cycle to 50%")
			} else {
				dlist := getdutylist(tlist, fc, &config[0])
				for i := range dlist {

					fanindx := strconv.Itoa(int(i))
					duty := strconv.Itoa(int(dlist[i]))
					bees := []byte(fanindx + duty)

					_, err := port.Write(bees)
					if err != nil {
						log.Printf("Error in port.Write:%v\n", err)
					}
					err = port.Flush()
					if err != nil {
						log.Printf("Error in port.Write:%v\n", err)
					}

				}
				buf := make([]byte, 5)
				bees := []byte("2000")
				_, err := port.Write(bees)
				if err != nil {
					log.Printf("Error in port.Write:%v\n", err)
				}
				n, err := port.Read(buf)
				if err != nil {
					log.Printf("Error in Reading Port:%v\n", err)
				}
				fan1rpm = strings.Trim(string(buf[:n]), "\r")
				err = port.Flush()
				if err != nil {
					log.Printf("Error in port.Flush:%v\n", err)
				}
				bees = []byte("3000")
				n, err = port.Write(bees)
				if err != nil {
					log.Printf("Error in port.Write:%v\n", err)
				}

				n, err = port.Read(buf)
				if err != nil {
					log.Printf("Error in Reading Port:%v\n", err)
				}
				fan2rpm = strings.Trim(string(buf[:n]), "\r")
				err = port.Flush()
				if err != nil {
					log.Printf("Error in port.Flush:%v\n", err)
				}

			}

		}

	}

}

func getdutylist(tlist []uint, fc *fancurve, c *Controller) (dlist []uint) {
	dlist = make([]uint, len(c.Fans))
	for i := range c.Fans {
		for j := range c.Fans[i].GpuIDs {
			if dlist[i] < tlist[c.Fans[i].GpuIDs[j]] {
				dlist[i] = tlist[c.Fans[i].GpuIDs[j]]
			}
		}
		dlist[i] = fc.tempmap(dlist[i])
	}
	return dlist
}

type fancurve struct {
	temp []uint
	duty []uint
}

func createfancurve(temp, duty []uint) (fc *fancurve) {
	fc = new(fancurve)
	if len(temp) != len(duty) {
		return nil
	}
	fc.duty = make([]uint, len(duty))
	fc.temp = make([]uint, len(temp))
	copy(fc.duty, duty)
	copy(fc.temp, temp)
	return fc
}
func (fc *fancurve) update(temp, duty []uint) {
	fc.duty = make([]uint, len(duty))
	fc.temp = make([]uint, len(temp))
	copy(fc.duty, duty)
	copy(fc.temp, temp)
}
func (fc *fancurve) tempmap(t uint) (dc uint) {
	if fc.temp[0] > t {
		return fc.duty[0]
	}
	for i := range fc.temp {
		if fc.temp[i] > t {
			return mapdata(t, fc.temp[i-1], fc.temp[i], fc.duty[i-1], fc.duty[i])
		}
	}
	return uint(999)
}

func mapdata(in, inmin, inmax, outmin, outmax uint) (out uint) {
	return (in-inmin)*(outmax-outmin)/(inmax-inmin) + outmin

}

func getgputemps(temps []uint) []uint {
	templist := nqueryall("temperature.gpu")
	for i := range templist {
		tint, err := strconv.Atoi(templist[i])
		if err != nil {
			return nil

		}
		temps[i] = uint(tint)

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
