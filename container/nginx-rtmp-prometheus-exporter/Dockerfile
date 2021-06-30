FROM golang:1.16 AS builder

WORKDIR /go/src/nginx-rtmp-prometheus-exporter
COPY . .

RUN CGO_ENABLED=0 go build -o /nginx-rtmp-prometheus-exporter


FROM busybox

COPY --from=builder /nginx-rtmp-prometheus-exporter /nginx-rtmp-prometheus-exporter

RUN adduser -G users -S -H user
USER user

CMD /nginx-rtmp-prometheus-exporter
