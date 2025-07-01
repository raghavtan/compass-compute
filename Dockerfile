FROM golang:1.24-alpine AS builder

ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH

RUN apk add --no-cache git ca-certificates tzdata
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -a -installsuffix cgo \
    -o compass-compute \
    ./cmd/compass-compute

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /build/compass-compute /compass-compute

USER appuser:appgroup

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/compass-compute", "--version"]

ENTRYPOINT ["/compass-compute"]
CMD ["--help"]

LABEL maintainer="CloudRuntime <cloud-runtime@onefootball.com>"
LABEL version="${VERSION}"
LABEL description="Compass Compute CLI Tool"