FROM golang:1.23.8-alpine AS builder
ARG VERSION
ARG COMMIT_ID
WORKDIR /app
RUN apk add --no-cache build-base tzdata
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build VERSION=${VERSION} COMMIT_ID=${COMMIT_ID}

FROM metacubex/mihomo:v1.19.12
WORKDIR /app
RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/dist/mihomo-updater /mihomo-updater
RUN chmod +x /mihomo-updater
ENTRYPOINT ["/mihomo-updater"]
