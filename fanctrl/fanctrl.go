package fanctrl

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/dereklstinson/gpufans/smi"
	"github.com/tarm/serial"
)

type Fan struct {
	Index  uint8   `json:"Index,omitempty"`
	GpuIDs []uint8 `json:"GpuIDs,omitempty"`
}
type Controller struct {
	Port     string          `json:"Port,omitempty"`
	Baud     int             `json:"Baud,omitempty"`
	Parity   serial.Parity   `json:"Parity,omitempty"`
	StopBits serial.StopBits `json:"StopBits,omitempty"`
	Fans     []Fan           `json:"Fans,omitempty"`
	Temps    []uint          `json:"Temps,omitempty"`
	Dutys    []uint          `json:"Dutys,omitempty"`
	p        *serial.Port
	index    int
}

const dfaultfancontrol = `[{
	"Port": "/dev/ttyUSB0",
	"Baud": 19200,
	"Fans":[{
			"Index": 0,
			"GpuIDs": [0,1]
			},
			{
			"Index": 1,
			"GpuIDs": [2,3]
			}],
	"Temps":[30,40, 75,90],
	"Dutys":[100,200,500,999]
  }]`

//CreateFanControllers default fan config is config.json
//config *serial.Config is
func CreateFanControllers(configfile string) (c []Controller) {
	data, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Printf("error in ReadFile() %v", err)
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		log.Printf("error in Unmarshal() %v", err)
	}
	for i := range c {
		var smode = &serial.Config{Name: c[i].Port, Baud: c[i].Baud, Parity: c[i].Parity, StopBits: c[i].StopBits}
		c[i].p, err = serial.OpenPort(smode)
		if err != nil {
			log.Printf("error in OpenPort() %v\n", err)
		}
	}

	return c
}
func CloseAll(c []Controller) error {
	var s string

	for i := range c {
		s = s + c[i].Close().Error() + " "
	}
	return errors.New(s)
}
func (c *Controller) AdjustFanSpeed() {
	dlist := c.getdutycurve(smi.QueryAllGpuTemps())
	for i := range dlist {
		fanindx := strconv.Itoa(i)
		duty := strconv.Itoa(int(dlist[i]))
		bees := []byte(fanindx + duty)

		_, err := c.p.Write(bees)
		if err != nil {
			log.Printf("Error: Controller %v -- port.Write:%v\n", i, err)
		}
		err = c.p.Flush()
		if err != nil {
			log.Printf("Error: Controller %v --  port.Flush:%v\n", i, err)
		}
	}

}
func (c *Controller) GetFansRPM() (rpm []string) {
	rpm = make([]string, 2)

	for i := range c.Fans {
		buf := make([]byte, 5)
		_, err := c.p.Write([]byte(strconv.Itoa(i+2) + "000"))
		if err != nil {
			log.Printf("Error in port.Write:%v\n", err)
		}
		n, err := c.p.Read(buf)
		if err != nil {
			log.Printf("Error in Reading Port:%v\n", err)
		}
		rpm[i] = strings.Trim(string(buf[:n]), "\r")
		err = c.p.Flush()
		if err != nil {
			log.Printf("Error in Flush Port:%v\n", err)
		}
	}
	return rpm
}

//Close closes the serial port to controller
func (c *Controller) Close() error {
	return c.p.Close()
}
func (c *Controller) getdutycurve(tlist []uint) (dlist []uint) {
	dlist = make([]uint, len(c.Fans))
	for i := range c.Fans {
		for j := range c.Fans[i].GpuIDs {
			if dlist[i] < tlist[c.Fans[i].GpuIDs[j]] {
				dlist[i] = tlist[c.Fans[i].GpuIDs[j]]
			}
		}
		dlist[i] = c.tempmap(dlist[i])
	}
	return dlist
}

func (c *Controller) tempmap(t uint) (dc uint) {
	if c.Temps[0] > t {
		return c.Dutys[0]
	}
	for i := range c.Temps {
		if c.Temps[i] > t {
			return mapdata(t, c.Temps[i-1], c.Temps[i], c.Dutys[i-1], c.Dutys[i])
		}
	}
	return uint(999)
}

func mapdata(in, inmin, inmax, outmin, outmax uint) (out uint) {
	return (in-inmin)*(outmax-outmin)/(inmax-inmin) + outmin

}

/*func (c *Controller) Update(configfile string) error {
	data, err := ioutil.ReadFile(configfile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	return nil
}*/
