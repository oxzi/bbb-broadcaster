FROM golang:1.16 AS builder

WORKDIR /go/src/nginx-hls-prometheus-exporter
COPY . .

RUN CGO_ENABLED=0 go build -o /nginx-hls-prometheus-exporter


FROM busybox

COPY --from=builder /nginx-hls-prometheus-exporter /nginx-hls-prometheus-exporter
COPY ./entrypoint.sh /entrypoint.sh

RUN adduser -G users -S -H user
USER user

CMD /entrypoint.sh
