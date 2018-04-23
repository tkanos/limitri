package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
)

type bombardier struct{}

/*
usage: bombardier [<flags>] <url>

Fast cross-platform HTTP benchmarking tool

Flags:
      --help                  Show context-sensitive help (also try --help-long and --help-man).
      --version               Show application version.
  -c, --connections=125       Maximum number of concurrent connections
  -t, --timeout=2s            Socket/request timeout
  -l, --latencies             Print latency statistics
  -m, --method=GET            Request method
  -b, --body=""               Request body
  -f, --body-file=""          File to use as request body
  -s, --stream                Specify whether to stream body using chunked transfer encoding or to serve it from memory
      --cert=""               Path to the client's TLS Certificate
      --key=""                Path to the client's TLS Certificate Private Key
  -k, --insecure              Controls whether a client verifies the server's certificate chain and host name
  -H, --header="K: V" ...     HTTP headers to use(can be repeated)
  -n, --requests=[pos. int.]  Number of requests
  -d, --duration=10s          Duration of test
  -r, --rate=[pos. int.]      Rate limit in requests per second
      --fasthttp              Use fasthttp client
      --http1                 Use net/http client with forced HTTP/1.x
      --http2                 Use net/http client with enabled HTTP/2.0
  -p, --print=<spec>          Specifies what to output. Comma-separated list of values 'intro' (short: 'i'), 'progress' (short: 'p'),
                              'result' (short: 'r'). Examples:

                                * i,p,r (prints everything)
                                * intro,result (intro & result)
                                * r (result only)
                                * result (same as above)
  -q, --no-print              Don't output anything
  -o, --format=<spec>         Which format to use to output the result. <spec> is either a name (or its shorthand) of some format understood
                              by bombardier or a path to the user-defined template, which uses Go's text/template syntax, prefixed with
                              'path:' string (without single quotes), i.e. "path:/some/path/to/your.template" or
                              "path:C:\some\path\to\your.template" in case of Windows. Formats understood by bombardier are:

                                * plain-text (short: pt)
                                * json (short: j)

Args:
  <url>  Target's URL
*/
// Execute bombardier -c {c} -d {d} -l -o j url
func (b bombardier) Execute(url string, c int, n int, d string) (*perf, error) {
	//o, err := exec.Command("bombardier", "-c", strconv.Itoa(c), "-d", d, "-l", "-o", "j", url).Output()
	o, err := exec.Command("bombardier", "-c", strconv.Itoa(c), "-d", d, "-l", "-o", "j", url).Output()
	if err != nil {
		return nil, err
	}

	return parse(o)
}

/* response format :
Bombarding http://www.example.com with 1 request(s) using 1 connection(s)
 1 / 1 [=================================================================================================================================================================] 100.00% 0s
Done!
{"spec":{"numberOfConnections":1,"testType":"number-of-requests","numberOfRequests":1,"method":"GET","url":"http://www.example.com","body":"","stream":false,"timeoutSeconds":2,"client":"fasthttp"},"result":{"bytesRead":1577,"bytesWritten":63,"timeTakenSeconds":0.071811825,"req1xx":0,"req2xx":1,"req3xx":0,"req4xx":0,"req5xx":0,"others":0,"latency":{"mean":71221,"stddev":0,"max":71221,"percentiles":{"50":71221,"75":71221,"90":71221,"95":71221,"99":71221}},"rps":{"mean":36.55632942867867,"stddev":57.800631945033224,"max":146.2253177147147,"percentiles":{"50":0.000000,"75":0.000000,"90":146.225318,"95":146.225318,"99":146.225318}}}}
{
  "spec": {
    "numberOfConnections": 1,
    "testType": "number-of-requests",
    "numberOfRequests": 1,
    "method": "GET",
    "url": "http:\/\/www.example.com",
    "body": "",
    "stream": false,
    "timeoutSeconds": 2,
    "client": "fasthttp"
  },
  "result": {
    "bytesRead": 1577,
    "bytesWritten": 63,
    "timeTakenSeconds": 0.071811825,
    "req1xx": 0,
    "req2xx": 1,
    "req3xx": 0,
    "req4xx": 0,
    "req5xx": 0,
    "others": 0,
    "latency": {
      "mean": 71221,
      "stddev": 0,
      "max": 71221,
      "percentiles": {
        "50": 71221,
        "75": 71221,
        "90": 71221,
        "95": 71221,
        "99": 71221
      }
    },
    "rps": {
      "mean": 36.556329428679,
      "stddev": 57.800631945033,
      "max": 146.22531771471,
      "percentiles": {
        "50": 0,
        "75": 0,
        "90": 146.225318,
        "95": 146.225318,
        "99": 146.225318
      }
    }
  }
}
*/
// parse the response of bombardier
func parse(o []byte) (*perf, error) {
	type model struct {
		Result struct {
			Req1xx  int `json:"req1xx"`
			Req2xx  int `json:"req2xx"`
			Req3xx  int `json:"req3xx"`
			Req4xx  int `json:"req4xx"`
			Req5xx  int `json:"req5xx"`
			Latency struct {
				Mean        float64 `json:"mean"`
				Percentiles struct {
					Perc50 float64 `json:"50"`
					Perc75 float64 `json:"75"`
					Perc90 float64 `json:"90"`
					Perc95 float64 `json:"95"`
					Perc99 float64 `json:"99"`
				} `json:"percentiles"`
			} `json:"latency"`
			Rps struct {
				Mean float64 `json:"mean"`
			} `json:"rps"`
		} `json:"result"`
	}

	index := bytes.Index(o, []byte("{\"spec\":{\""))

	var m model
	if err := json.Unmarshal(o[index:], &m); err != nil {
		return nil, err
	}

	return &perf{
		avg:      m.Result.Latency.Mean,
		reqbysec: m.Result.Rps.Mean,
		req1xx:   m.Result.Req1xx,
		req2xx:   m.Result.Req2xx,
		req3xx:   m.Result.Req3xx,
		req4xx:   m.Result.Req4xx,
		req5xx:   m.Result.Req5xx,
		perc50:   m.Result.Latency.Percentiles.Perc50,
		perc75:   m.Result.Latency.Percentiles.Perc75,
		perc90:   m.Result.Latency.Percentiles.Perc90,
		perc95:   m.Result.Latency.Percentiles.Perc95,
		perc99:   m.Result.Latency.Percentiles.Perc99,
	}, nil
}
