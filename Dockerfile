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

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X main.version=${VERSION}" \
      -o /out/gmessages \
      .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S gmessages && \
    adduser -S -G gmessages -h /home/gmessages gmessages && \
    mkdir -p /data && chown gmessages:gmessages /data
USER gmessages
WORKDIR /home/gmessages
COPY --from=build /out/gmessages /usr/local/bin/gmessages

ENV GMESSAGES_DATA_DIR=/data \
    GMESSAGES_HOST=0.0.0.0 \
    GMESSAGES_PORT=7007

VOLUME ["/data"]
EXPOSE 7007

# Default to running the server. Override with `pair`, `import`, etc.
ENTRYPOINT ["gmessages"]
CMD ["serve"]
