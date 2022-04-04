package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	pipereader, err := os.Open("../rpmfifo")
	if err != nil {
		panic(err)
	}
	defer pipereader.Close()
	for {
		time.Sleep(time.Microsecond * 1000)

		data, err := ioutil.ReadAll(pipereader)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	}
}
