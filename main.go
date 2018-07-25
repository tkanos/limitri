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
	d := flag.Int("d", 5, "duration in seconds of each requests")
	l := flag.Bool("l", false, "show the latency on the output graphic")
	m := flag.String("m", "GET", "method GET/POST?PUT/DELETE")
	b := flag.String("b", "", "Body")
	flag.Parse()

	if *url == "" {
		fmt.Println("url not defined")
		os.Exit(-1)
	}

	type Exec func(url string, c int, n int, duration string, method string, body string) (*perf, error)
	var exec Exec = bombardier{}.Execute

	var result []*perf

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	fmt.Print("concurrent Connection : ")
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker.C {
			fmt.Print(".")
		}
	}()

	concconection := 1
	duration := fmt.Sprintf("%ds", *d)
	thread := 1
	baseLine, err := exec(*url, concconection, thread, duration, *m, *b)
	fmt.Printf("%d", concconection)
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
			} else {
				concconection = concconection + 50
			}
			p, err := exec(*url, concconection, 0, duration, *m, *b)
			fmt.Printf("%d", concconection)
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
	ticker.Stop()
	fmt.Println()
	if len(result) == 0 {
		fmt.Println("No result found")
		os.Exit(-1)
	}
	div, valueType := getValueType(result)
	data := graphData(result, div, *l)
	graphURL, err := writeLocallyData(string(data), "Avg Time in "+valueType+" / Request by second", valueType)
	max := getMax(result)
	fmt.Printf("Avg Baseline: %f %s\r\n", result[0].avg/div, valueType)
	fmt.Printf("Avg Max: %f %s\r\n", max.avg/div, valueType)
	fmt.Printf("Request Max: %f req/s\r\n", max.reqbysec)
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println(graphURL)
	fmt.Println()
	fmt.Println("=========================================")

}

func stopInformation(conccurentConnection int, thread int, reason string) string {
	return fmt.Sprintf("\nProcess stop at -c %d -t %d because %s", conccurentConnection, thread, reason)
}

func getValueType(result []*perf) (float64, string) {
	var div float64 = 1
	valueType := "μs"
	var totalAvg float64
	for _, p := range result {
		totalAvg = totalAvg + p.avg
	}
	if (totalAvg / float64(len(result))) > 1000 {
		div = 1000
		valueType = "ms"
	}

	return div, valueType
}

func getMax(result []*perf) *perf {
	if len(result) == 1 {
		return result[0]
	} else if len(result) == 0 {
		return nil
	}

	last := result[len(result)-1]
	plast := result[len(result)-2]

	if last.reqbysec > plast.reqbysec {
		return last
	}

	return plast
}
