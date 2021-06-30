#!/bin/sh

tail -f /var/log/nginx/access.log | /nginx-hls-prometheus-exporter
