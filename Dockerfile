FROM golang:1.22.0-alpine3.19 as builder

LABEL version="v0.0.2"

ENV REPODIR=/go/src/gitlab.com/neverlless/cloudflare-prometheus-exporter
WORKDIR ${REPODIR}

COPY go.mod go.sum ${REPODIR}/
RUN go mod download

COPY . ${REPODIR}/

RUN apk update && apk add --no-cache git ca-certificates && \
    update-ca-certificates

RUN CGO_ENABLED=0 GOOS=linux go build -o /cloudflare_exporter -a -installsuffix cgo -ldflags '-extldflags "-static"' .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /cloudflare_exporter /cloudflare_exporter

ENTRYPOINT ["/cloudflare_exporter"]
