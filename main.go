package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	flag "github.com/spf13/pflag"
)

var (
	rawFields []string
	fsFields  []string
)

func init() {
	rawFields = []string{
		"io_rate",
		"mb_ps",
		"bytes",
		"read_pct",
		"resp_time",
		"read_resp",
		"write_resp",
		"read_max",
		"write_max",
		"resp_stddev",
		"q_depth",
		"cpu_total",
		"cpu_sys",
	}

	fsFields = []string{
		"req_std_ops_rate",
		"req_std_ops_resp",
		"cpu_total",
		"cpu_sys",
		"read_pct",
		"read_rate",
		"read_resp",
		"write_rate",
		"write_resp",
		"read_mb_ps",
		"write_mb_ps",
		"total_mb_ps",
		"xfer_size",
		"mkdir_rate",
		"mkdir_resp",
		"rmdir_rate",
		"rmdir_resp",
		"create_rate",
		"create_resp",
		"open_rate",
		"open_resp",
		"close_rate",
		"close_resp",
		"delete_rate",
		"delete_resp",
	}
}

type vdbenchCollector struct {
	descMap map[string]*prometheus.Desc
	host    string             // where the test is run
	latest  map[string]float64 // latest vdbench output metrics
	mutex   sync.Mutex
}

// fs: true - file system workload; false - raw workload
func newCollector(fs bool) *vdbenchCollector {
	var fields []string
	if fs {
		fields = fsFields
	} else {
		fields = rawFields
	}

	descMap := map[string]*prometheus.Desc{}
	for _, name := range fields {
		descMap[name] = prometheus.NewDesc(
			name,
			strings.ReplaceAll(name, "_", " "),
			[]string{"host"},
			nil,
		)
	}

	var host string
	hostname, err := os.Hostname()
	if err != nil {
		host = "unknown"
	} else {
		host = hostname
	}

	vc := vdbenchCollector{
		descMap: descMap,
		host:    host,
	}
	return &vc
}

func (vc *vdbenchCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range vc.descMap {
		ch <- desc
	}
}

func (vc *vdbenchCollector) Collect(ch chan<- prometheus.Metric) {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	for k, v := range vc.latest {
		ch <- prometheus.MustNewConstMetric(
			vc.descMap[k],
			prometheus.GaugeValue,
			v,
			vc.host,
		)
	}
}

func parseFloat64(rawValues []string) []float64 {
	var values []float64

	for _, rawValue := range rawValues {
		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil
		}
		values = append(values, value)
	}
	return values
}

func getValues(fields []string) map[string]float64 {
	values := parseFloat64(fields[2:])
	if values == nil {
		return nil
	}

	valueMap := map[string]float64{}
	fLen := len(values)
	rawLen := len(rawFields)
	fsLen := len(fsFields)

	if fLen == rawLen {
		for index, key := range rawFields {
			valueMap[key] = values[index]
		}
	} else if fLen == fsLen {
		for index, key := range fsFields {
			valueMap[key] = values[index]
		}
	} else {
		return nil
	}
	return valueMap
}

func scanOutput(vmch chan map[string]float64) {
	targetPattern := regexp.MustCompile(`^(\d|\s|\:|\.)+?$`)
	splitPattern := regexp.MustCompile(`\s+`)

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}()
	for scanner.Scan() {
		line := scanner.Text()
		// Output the original line for reference
		fmt.Println(line)

		if matched := targetPattern.Match([]byte(line)); matched {
			fields := splitPattern.Split(line, -1)
			if l := len(fields); l == len(rawFields)+2 || l == len(fsFields)+2 {
				valueMap := getValues(fields)
				if valueMap != nil {
					vmch <- valueMap
				}
			}
		}
	}
}

func deleteJob(pusher *push.Pusher, gateway, job string) {
	if err := pusher.Delete(); err != nil {
		fmt.Println("Fail to delete the Pushgateway job:", err)
		fmt.Println("Please delete the job manually as:", fmt.Sprintf("curl -X DELETE %s/metrics/job/%s", gateway, job))
	}
}

func main() {
	// Register signal handler
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Parse command line options
	var job string
	var gateway string
	flag.StringVarP(&job, "job", "j", "", "Pushgateway job name")
	flag.StringVarP(&gateway, "gateway", "g", "http://127.0.0.1:9091", "Pushgateway URL")
	flag.Parse()
	if job == "" || gateway == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Start to intercept vdbench output
	vmch := make(chan map[string]float64)
	go scanOutput(vmch)

	// Use the first vdbench valid result for collector initialization
	initValueMap := <-vmch
	var fs bool
	if len(initValueMap) == len(rawFields) {
		fs = false
	} else {
		fs = true
	}
	vc := newCollector(fs)
	reg := prometheus.NewRegistry()
	reg.MustRegister(vc)

	// Init pusher
	pusher := push.New(gateway, job).Collector(vc)

	// Delete the job if the application is stoped unexpectedly
	go func() {
		<-sigc
		fmt.Println("Signal captured, exit the application")
		deleteJob(pusher, gateway, job)
		defer os.Exit(1)
	}()

	// Delete the Pushgateway job after finishing the job
	defer func() {
		deleteJob(pusher, gateway, job)
	}()

	// Iterate vdbench valid results and push data to Pushgateway
	for valueMap := range vmch {
		// Always use the latest vdbench valid result
		vc.mutex.Lock()
		vc.latest = valueMap
		vc.mutex.Unlock()

		// Push data
		if err := pusher.Push(); err != nil {
			fmt.Println("Fail to push data to:", gateway)
			fmt.Println(err)
			break
		}
	}
}
