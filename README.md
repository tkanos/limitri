# Limitri

Benchmark and test the limit of your program.

Limitri uses bombardier (so it should be installed) to perform the benchmark tests.
```bash
go get -u github.com/codesenberg/bombardier
```

for more documentation : https://github.com/codesenberg/bombardier


## Usage :

```bash
Usage of limitri: limitri -u <url> 
   -b string
        Body
  -d int
        duration in seconds of each requests (default 5)
  -l    show the latency on the output graphic
  -m string
        method GET/POST?PUT/DELETE (default "GET")
  -u string
        The url to test
  
```

Doing so, limitri will send benchmark request to bombardier to test the limit of your app (can take some times)
And when it's found, limitri provide you a report.

## example

```bash
$ ./limitri -u "http://localhost:8080"
concurrent Connection : .....1......2......3......4......5......6......7
Process stop at -c 7 -t 7 because the nb of requests are not increasing anymore

Avg Baseline: 33.729546 μs
Avg Max: 60.595068 μs
Request Max: 112614.959772 req/s
=========================================

/tmp/limitri582687153.html (you can see this report of the examples folder)

=========================================

```
# Conditions

Of course it should only be done on test mode, and not in a cluster.

Limitri will benchmark, thanks to bombardier, increasing the benchmark eash time.
and will only stop if it founds one of the following conditions :
- avg time increase more than 10 times (compared to the baseline)
- Nb error is more thn 10% of the request
- we have less requests serve than previously (asking more)
- the increase of the number of request tend to be stable (<3%)

If you want more conditions, don't hesitate to do a Pull Request.


