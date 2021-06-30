/*
 * Copyright (c) 2021 Alvar Penning
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"bufio"
	"net/http"
	"os"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// NginxHlsFragment describes a HLS fragment length,
	// <https://github.com/arut/nginx-rtmp-module/wiki/Directives#hls_fragment>.
	NginxHlsFragment = 5

	// NginxLogRegexp matches a nginx log entry,
	// based on <https://docs.fluentd.org/parser/nginx>.
	NginxLogRegexp = `(?P<remote>[^ ]*) (?P<host>[^ ]*) (?P<user>[^ ]*) \[(?P<time>[^\]]*)\] "(?P<method>\S+)(?: +(?P<path>[^\"]*?)(?: +\S*)?)?" (?P<code>[^ ]*) (?P<size>[^ ]*)(?: "(?P<referer>[^\"]*)" "(?P<agent>[^\"]*)"(?:\s+(?P<http_x_forwarded_for>[^ ]+))?)?`

	// PathRegexp matches an HLS file, indicating a viewer.
	PathRegexp = `^\/live\/.+\/.+\.m3u8$`

	// ListenAddr will be used for the httpd.
	ListenAddr = ":9101"
)

var (
	// counter updated by the nginx log as it passes by.
	counter *Counter

	// counterGauge is the Prometheus gauge for current HLS viewers.
	counterGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nginx_hls_viewer",
		Help: "Number of HLS viewers for all streams.",
	})
)

// readLog from stdin and forward to the counter.
func readLog() {
	pathRegexp := regexp.MustCompile(PathRegexp)
	logRegexp := regexp.MustCompile(NginxLogRegexp)

	logScanner := bufio.NewScanner(os.Stdin)

	for logScanner.Scan() {
		lineMatch := logRegexp.FindStringSubmatch(logScanner.Text())
		if lineMatch == nil {
			continue
		}

		lineFields := make(map[string]string)
		for i, name := range logRegexp.SubexpNames() {
			if i != 0 && name != "" {
				lineFields[name] = lineMatch[i]
			}
		}

		if !pathRegexp.MatchString(lineFields["path"]) {
			continue
		}

		log.WithField("fields", lineFields).Info("nginx log matches")
		counter.Update(lineFields)
	}
	if err := logScanner.Err(); err != nil {
		log.WithError(err).Error("Scanner errored")
	}
}

func main() {
	counter = NewCounter(NginxHlsFragment*time.Second, counterGauge.Set)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(ListenAddr, nil)

	readLog()
}
