package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

var rawFields = []string{
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

var fsFields = []string{
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

func getMetrics(fields []string) map[string]float64 {
	values := parseFloat64(fields[2:])
	if values == nil {
		return nil
	}

	metrics := map[string]float64{}
	fLen := len(values)
	rawLen := len(rawFields)
	fsLen := len(fsFields)

	if fLen == rawLen {
		for index, key := range rawFields {
			metrics[key] = values[index]
		}
	} else if fLen == fsLen {
		for index, key := range fsFields {
			metrics[key] = values[index]
		}
	} else {
		return nil
	}
	return metrics
}

func main() {
	targetPattern := regexp.MustCompile(`^(\d|\s|\:|\.)+?$`)
	splitPattern := regexp.MustCompile(`\s+`)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		if matched := targetPattern.Match([]byte(line)); matched {
			fields := splitPattern.Split(line, -1)
			if l := len(fields); l == len(rawFields)+2 || l == len(fsFields)+2 {
				metrics := getMetrics(fields)
				if metrics != nil {
					log.Printf("%+v\n", metrics)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
