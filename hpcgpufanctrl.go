package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/dereklstinson/gpufans/fanctrl"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	//arduino config defaults are 8 data bits, no parity, one stop bit.
	ctrls := fanctrl.CreateFanControllers("config.json")

	for {
		time.Sleep(time.Millisecond * 50)
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			for i, ctrl := range ctrls {
				if err := ctrl.Close(); err != nil {
					log.Printf("ctrls[%v].Close error:%v\n", i, err)
				}
			}
			fmt.Println("Interupted Program Closed with Elegance")
			stop()

			return

		default:
			var fanrpms string
			for _, ctrl := range ctrls {
				ctrl.AdjustFanSpeed()

				for _, stng := range ctrl.GetFansRPM() {
					fanrpms = fanrpms + " " + stng
				}
			}

		}
		//	if time.Since(ptime).Milliseconds() > 2000 {
		//		ptime = time.Now()
		//			_, err = fmt.Fprint(pipe, fanrpms)
		//			if err != nil {
		//				log.Printf("Error in writing to pipe: %v\n", err)
		//			}
		//	}

	}

}
