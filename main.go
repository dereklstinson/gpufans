package main

import (
	"fmt"
	"log"

	"github.com/dereklstinson/gpu-monitoring-tools/bindings/go/nvml"
)

func devicelistnvidia() []string {
	return nil
}

func main() {

}
func test1() {
	err := nvml.Init()
	defer nvml.Shutdown()
	if err != nil {
		log.Fatalf("Unable to initialize NVML: %v", err)
	}
	ndevs, err := nvml.GetDeviceCount()
	if err != nil {
		log.Fatalf("Unable to Get Numver of Devices: %v", err)
	}
	for i := (uint)(0); i < ndevs; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.Fatalf("Device %v not found. Error Returned: %v", i, err)
		}
		status, err := device.Status()
		if err != nil {
			log.Fatalf("Device %v Status Error. Error Returned: %v", i, err)
		}

		fmt.Printf("GPU temperature is %v!\n", *status.Temperature)
	}
}
