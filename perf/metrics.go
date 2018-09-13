package perf

import (
	"expvar"
	"fmt"
	"net/http"
	"strconv"
)

var (
	handleExported *expvar.Int
	connExported   *expvar.Int
	timeExported   *expvar.Float
	qpsExported    *expvar.Float
)

func init() {
	handleExported = expvar.NewInt("TotalHandle")
	connExported = expvar.NewInt("TotalConn")
	timeExported = expvar.NewFloat("TotalTime")
	qpsExported = expvar.NewFloat("QPS")
}

// MonitorOn starts up an HTTP monitor on port.
func MonitorOn(port int) {
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			fmt.Println(err)
			return
		}
	}()
}

func addTotalConn(delta int64) {
	connExported.Add(delta)
	calculateQPS()
}

func addTotalHandle() {
	handleExported.Add(1)
	calculateQPS()
}

func addTotalTime(seconds float64) {
	timeExported.Add(seconds)
	calculateQPS()
}

func calculateQPS() {
	totalConn, err := strconv.ParseInt(connExported.String(), 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	totalTime, err := strconv.ParseFloat(timeExported.String(), 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	totalHandle, err := strconv.ParseInt(handleExported.String(), 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	if float64(totalConn)*totalTime != 0 {
		// take the average time per worker go-routine
		qps := float64(totalHandle) / (float64(totalConn) * (totalTime / float64(1)))
		qpsExported.Set(qps)
	}
}
