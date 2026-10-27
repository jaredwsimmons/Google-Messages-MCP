# Google Messages MCP — headless MCP server
#
# Run the backend as a server, exposing the local message store and MCP
# endpoint to assistants. Useful when you want to keep the inbox running on a
# home server / NAS and connect from desktop clients.
#
# Build:   docker build -t gmessages .
# Run:     docker run -p 7007:7007 -v gmessages-data:/data gmessages
# Pair:    docker exec -it <container> gmessages pair
# Connect: claude mcp add -s user --transport sse gmessages http://<host>:7007/mcp/sse

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build
RUN apk add --no-cache tzdata && \
    addgroup -S -g 101 gmessages && \
    adduser -S -u 100 -G gmessages -h /home/gmessages gmessages && \
    mkdir -p /data
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
      -trimpath \
      -ldflags "-s -w -X main.version=${VERSION}" \
      -o /out/gmessages \
      .

# BusyBox keeps the shell and wget needed by the Compose health check.
FROM busybox:1.37.0-musl
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/passwd /etc/group /etc/
COPY --from=build --chown=100:101 /home/gmessages /home/gmessages
COPY --from=build --chown=100:101 /data /data
USER 100:101
WORKDIR /home/gmessages
COPY --from=build /out/gmessages /usr/local/bin/gmessages

ENV HOME=/home/gmessages \
    GMESSAGES_DATA_DIR=/data \
    GMESSAGES_HOST=0.0.0.0 \
    GMESSAGES_PORT=7007

VOLUME ["/data"]
EXPOSE 7007

# Default to running the server. Override with `pair`, `import`, etc.
ENTRYPOINT ["gmessages"]
CMD ["serve"]
