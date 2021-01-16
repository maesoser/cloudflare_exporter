FROM golang:alpine as builder

LABEL version=1.0

ENV REPODIR=/go/src/gitlab.com/maesoser/cloudflare-prometheus-exporter

WORKDIR ${REPODIR}
COPY *.go ${REPODIR}/
COPY collector/*.go ${REPODIR}/collector/
COPY *.toml ${REPODIR}/

RUN apk update && \
    apk add --no-cache git && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates 2>/dev/null || true && \
    go get -u github.com/golang/dep/cmd/dep && \
    dep ensure

RUN CGO_ENABLED=0 GOOS=linux go build -o /cloudflare_exporter -a -installsuffix cgo -ldflags '-extldflags "-static"' .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /cloudflare_exporter /cloudflare_exporter

ENTRYPOINT ["/cloudflare_exporter"]
