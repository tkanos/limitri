package main

type perf struct {
	avg      float64
	reqbysec float64
	req2xx   int
	req1xx   int
	req3xx   int
	req4xx   int
	req5xx   int
	perc50   float64
	perc75   float64
	perc90   float64
	perc95   float64
	perc99   float64
}

func (p perf) No2xxStatus() int {
	return p.req1xx + p.req3xx + p.req4xx + p.req5xx
}
