
FROM golang:1.25 AS builder

WORKDIR /backend

# Copy go modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o nodeherder .


FROM gcr.io/distroless/base-debian12

WORKDIR /backend
COPY --from=builder /backend/nodeherder .

EXPOSE 4100

# Environment variable placeholders
ENV APP_ENV=production
ENV MQTT_URL=""

ENTRYPOINT ["./nodeherder"]
