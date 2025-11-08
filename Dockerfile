FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -o streamecho .

FROM debian:trixie-slim
WORKDIR /app
ENV DEBIAN_FRONTEND=noninteractive
RUN apt update && \
    apt install -y \
    ca-certificates \
    libssl-dev \
    libz-dev \
    libvpx-dev \
    libx264-dev \
    libx265-dev \
    libopus-dev \
    libsdl2-dev \
    libavdevice-dev \
    libfreetype6-dev \
    libass-dev \
    librtmp-dev \
    ffmpeg
RUN rm -rf /var/lib/apt/lists/*

COPY --from=build /app/streamecho /app/streamecho
ENTRYPOINT ["./streamecho"]
