
FROM golang:1.25 AS builder

WORKDIR /backend

# Copy go modules first
COPY go.mod go.sum ./
COPY third_party/esphome_api/go.mod third_party/esphome_api/go.sum ./third_party/esphome_api/
RUN go mod download

# Copy the rest
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o nodeherder .


FROM gcr.io/distroless/base-debian12

WORKDIR /backend
COPY --from=builder /backend/nodeherder .

EXPOSE 4110

# Environment variable placeholders
ENV APP_ENV=production
ENV MQTT_URL=""

ENTRYPOINT ["./nodeherder"]
