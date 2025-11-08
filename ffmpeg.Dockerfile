FROM debian:trixie-slim

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

ENTRYPOINT ["ffmpeg"]
