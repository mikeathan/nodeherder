
# Stage 1: Build Vue frontend
FROM node:latest AS build-stage
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/. .
RUN npm run build

# # Stage 2: Build Golang backend
# FROM golang:1.20-alpine AS backend-builder

# ENV GO111MODULE=on \
#     CGO_ENABLED=0 \
#     GOOS=linux \
#     GOARCH=amd64

# WORKDIR /core
# COPY core/ ./
# RUN go mod download
# COPY *.go ./

# # Build
# RUN go build -o /nodeherder  main.go


# # Stage 3: Final image
# FROM alpine:latest
# WORKDIR /nodeherder
# COPY --from=frontend-builder /frontend/dist ./dist
# COPY --from=backend-builder /core/nodeherder .


# EXPOSE 4100

# # Run
# CMD ["/nodeherder"]