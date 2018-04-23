package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	url := flag.String("u", "", "The url to test")
	e := flag.String("e", "", "the executable to choose bombardier|wrk")
	d := flag.Int("d", 5, "duration in seconds of each requests")
	l := flag.Bool("l", false, "duration in seconds of each requests")
	flag.Parse()

	if *url == "" || *e == "" {
		fmt.Println("url not defined")
		os.Exit(-1)
	}

	type Exec func(url string, c int, n int, d string) (*perf, error)
	var exec Exec

	if *e == "bombardier" {
		exec = bombardier{}.Execute
	} else if *e == "wrk" {

	} else {
		fmt.Printf("executable %s not yet implemented\r\n", *e)
		os.Exit(-1)
	}

	var result []*perf

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	concconection := 1
	duration := fmt.Sprintf("%ds", *d)
	thread := 1
	baseLine, err := exec(*url, concconection, thread, duration)
	if err != nil {
		fmt.Println(err)
	}

	result = append(result, baseLine)

	var lastNbR = baseLine.reqbysec

	go func() {
		for true {
			time.Sleep(1 * time.Second)
			concconection++
			if thread < 13 {
				thread++
			}
			p, err := exec(*url, concconection, 0, duration)
			if err != nil {
				break
			}
			result = append(result, p)

			// if avg time increase more than 10 times (compqred to the baseline)
			if p.avg > baseLine.avg*10 {
				fmt.Println(stopInformation(concconection, thread, "the average time increase more than 10 times"))
				break
			}

			// if Nb error is more thn 10% of the request
			if p.No2xxStatus()*p.req2xx/100 > 10 {
				fmt.Println(stopInformation(concconection, thread, fmt.Sprintf("too much No 2xx requests %d", p.No2xxStatus())))
				break
			}

			// if we have less requests serve than previously
			if p.reqbysec < lastNbR {
				fmt.Println(stopInformation(concconection, thread, "Increasing the concurrent connection we had less requests serve"))
				break
			}

			// if the increase of the number of request tend to be stable (<3%)
			if (p.reqbysec*100/lastNbR)-100 < 3 {
				fmt.Println(stopInformation(concconection, thread, "the nb of requests are not increasing anymore"))
				break
			}

			lastNbR = p.reqbysec
			//todo : time stop ?
			//todo : c too high ?
		}
		close(c)
	}()

	<-c
	data := graphData(result, *l)
	graphURL, err := writeLocallyData(string(data), "Avg Time in μs / Request by second")

	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println(graphURL)
	fmt.Println()
	fmt.Println("=========================================")

}

func stopInformation(conccurentConnection int, thread int, reason string) string {
	return fmt.Sprintf("Process stop at -c %d -t %d because %s", conccurentConnection, thread, reason)
}
