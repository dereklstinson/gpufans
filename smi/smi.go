package smi

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func QueryAllGpuTemps() (temps []uint) {
	var out bytes.Buffer

	cmd := exec.Command("nvidia-smi", "--query-gpu="+"temperature.gpu", "--format=csv,noheader,nounits")
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatalf("nvsmi exec error: %v\n", err)

	}

	templist := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	temps = make([]uint, len(templist))
	for i := range templist {
		tint, err := strconv.Atoi(templist[i])
		if err != nil {
			return nil

		}
		temps[i] = uint(tint)

	}
	return temps
}
