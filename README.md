# BigBlueButton Broadcaster

## Install

This software was made to be deployed entirely with Docker Compose.
Thus, the installation is quite straightforward.

First, install both Docker and Docker Compose.
For Debian follow [this link for Docker](https://docs.docker.com/engine/install/debian/) and [this one for Docker Compose](https://docs.docker.com/compose/install/).

```sh
# Clone this repository including all its submodules
git clone --recurse-submodules https://github.com/oxzi/bbb-broadcaster.git
cd bbb-broadcaster

# Adjust the configuration
cp .env{.template,}
vim .env

# Go for a test drive
docker-compose up --build
```

## Prometheus and Grafana

By default, three Prometheus exporters are exposed, each on its own port.

- Port 9100: official [Node Exporter](https://github.com/prometheus/node_exporter)
- Port 9101: custom RTMP viewer exporter
- Port 9102: custom HLS viewer exporter

For the two custom exporters a Grafana Dashboard is shipped as `grafana_dashboard.json`.
