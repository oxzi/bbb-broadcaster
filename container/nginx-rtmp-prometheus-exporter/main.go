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
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/clbanning/mxj/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// NginxStatUrl is where rtmp_stats are available.
	NginxStatUrl = "http://nginx/stat"

	// ListenAddr will be used for the httpd.
	ListenAddr = ":9102"
)

var (
	// counterGauge is the Prometheus gauge for current RTMP viewers.
	counterGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nginx_rtmp_viewer",
		Help: "Number of RTMP viewers for all streams.",
	})

	// httpClient for the HTTP connections to reuse connections.
	httpClient = &http.Client{}
)

// fetchStat performs a single fetch.
func fetchStat() {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("recover", r).Error("recovered from fetching")
		}
	}()

	resp, err := httpClient.Get(NginxStatUrl)
	if err != nil {
		log.WithError(err).Error("cannot fetch nginx-rtmp's stats")
		return
	}
	defer func() {
		// make sure to completely read the body or the connection will be closed
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	xmlResp, err := mxj.NewMapXmlReader(resp.Body)
	if err != nil {
		log.WithError(err).Error("cannot decode xml")
		return
	}

	xmlRtmp := xmlResp["rtmp"].(map[string]interface{})
	xmlServer := xmlRtmp["server"].(map[string]interface{})
	xmlApplications := xmlServer["application"].([]interface{})

	for _, xmlApplicationIf := range xmlApplications {
		xmlApplication := xmlApplicationIf.(map[string]interface{})
		if xmlApplication["name"].(string) != "stream" {
			continue
		}

		xmlLive := xmlApplication["live"].(map[string]interface{})
		xmlNclients := xmlLive["nclients"].(string)

		nclients, err := strconv.Atoi(xmlNclients)
		if err != nil {
			log.WithError(err).Error("failed to parse nclients")
			continue
		}

		log.WithField("nclients", nclients).Debug("fetched clients")

		counterGauge.Set(float64(nclients))
		return
	}

	log.Error("found no matching application")
}

// fetchStats performs multiple fetches within a loop.
func fetchStats() {
	for {
		fetchStat()
		time.Sleep(time.Second)
	}
}

func main() {
	go fetchStats()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(ListenAddr, nil)
}
