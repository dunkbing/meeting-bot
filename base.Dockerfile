FROM ubuntu:22.04

ARG GSTREAMER_VERSION=1.18.5

RUN set -e; \
    export DEBIAN_FRONTEND=noninteractive; \
    sed -i 's/# deb-src/deb-src/g' /etc/apt/sources.list; \
    apt-get update; \
    apt-get dist-upgrade -y; \
    apt-get install -y --no-install-recommends \
        bison \
        bubblewrap \
        ca-certificates \
        cmake \
        flex \
        flite1-dev \
        gcc \
        gettext \
        git \
        gperf \
        iso-codes \
        liba52-0.7.4-dev \
        libaa1-dev \
        libaom-dev \
        libass-dev \
        libavcodec-dev \
        libavfilter-dev \
        libavformat-dev \
        libavutil-dev \
        libbs2b-dev \
        libbz2-dev \
        libcaca-dev \
        libcap-dev \
        libchromaprint-dev \
        libcurl4-gnutls-dev \
        libdca-dev \
        libde265-dev \
        libdrm-dev \
        libdv4-dev \
        libdvdnav-dev \
        libdvdread-dev \
        libdw-dev \
        libepoxy-dev \
        libfaac-dev \
        libfaad-dev \
        libfdk-aac-dev \
        libfluidsynth-dev \
        libgbm-dev \
        libgcrypt20-dev \
        libgirepository1.0-dev \
        libgl-dev \
        libgles-dev \
        libglib2.0-dev \
        libgme-dev \
        libgmp-dev \
        libgsl-dev \
        libgsm1-dev \
        libgudev-1.0-dev \
        libjpeg-dev \
        libkate-dev \
        liblcms2-dev \
        liblilv-dev \
        libmjpegtools-dev \
        libmodplug-dev \
        libmp3lame-dev \
        libmpcdec-dev \
        libmpeg2-4-dev \
        libmpg123-dev \
        libofa0-dev \
        libogg-dev \
        libopencore-amrnb-dev \
        libopencore-amrwb-dev \
        libopenexr-dev \
        libopenjp2-7-dev \
        libopus-dev \
        liborc-0.4-dev \
        libpango1.0-dev \
        libpng-dev \
        librsvg2-dev \
        librtmp-dev \
        libsbc-dev \
        libseccomp-dev \
        libshout3-dev \
        libsndfile1-dev \
        libsoundtouch-dev \
        libsoup2.4-dev \
        libspandsp-dev \
        libspeex-dev \
        libsrt-gnutls-dev \
        libsrtp2-dev \
        libssl-dev \
        libtag1-dev \
        libtheora-dev \
        libtwolame-dev \
        libudev-dev \
        libunwind-dev \
        libvisual-0.4-dev \
        libvo-aacenc-dev \
        libvo-amrwbenc-dev \
        libvorbis-dev \
        libvpx-dev \
        libvulkan-dev \
        libwavpack-dev \
        libwebp-dev \
        libwebrtc-audio-processing-dev \
        libwildmidi-dev \
        libwoff-dev \
        libx264-dev \
        libx265-dev \
        libxkbcommon-dev \
        libxslt1-dev \
        libzbar-dev \
        libzvbi-dev \
        python3 \
        python3-pip \
        ruby \
        wget \
        xdg-dbus-proxy

RUN pip3 install meson ninja
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*

RUN wget https://gstreamer.freedesktop.org/src/gstreamer/gstreamer-$GSTREAMER_VERSION.tar.xz && \
    tar -xf gstreamer-$GSTREAMER_VERSION.tar.xz && \
    rm gstreamer-$GSTREAMER_VERSION.tar.xz && \
    mv gstreamer-$GSTREAMER_VERSION gstreamer

RUN wget https://gstreamer.freedesktop.org/src/gst-plugins-base/gst-plugins-base-$GSTREAMER_VERSION.tar.xz && \
    tar -xf gst-plugins-base-$GSTREAMER_VERSION.tar.xz && \
    rm gst-plugins-base-$GSTREAMER_VERSION.tar.xz && \
    mv gst-plugins-base-$GSTREAMER_VERSION gst-plugins-base

RUN wget https://gstreamer.freedesktop.org/src/gst-plugins-bad/gst-plugins-bad-$GSTREAMER_VERSION.tar.xz && \
    tar -xf gst-plugins-bad-$GSTREAMER_VERSION.tar.xz && \
    rm gst-plugins-bad-$GSTREAMER_VERSION.tar.xz && \
    mv gst-plugins-bad-$GSTREAMER_VERSION gst-plugins-bad

RUN wget https://gstreamer.freedesktop.org/src/gst-plugins-good/gst-plugins-good-$GSTREAMER_VERSION.tar.xz && \
    tar -xf gst-plugins-good-$GSTREAMER_VERSION.tar.xz && \
    rm gst-plugins-good-$GSTREAMER_VERSION.tar.xz && \
    mv gst-plugins-good-$GSTREAMER_VERSION gst-plugins-good

RUN wget https://gstreamer.freedesktop.org/src/gst-plugins-ugly/gst-plugins-ugly-$GSTREAMER_VERSION.tar.xz && \
    tar -xf gst-plugins-ugly-$GSTREAMER_VERSION.tar.xz && \
    rm gst-plugins-ugly-$GSTREAMER_VERSION.tar.xz && \
    mv gst-plugins-ugly-$GSTREAMER_VERSION gst-plugins-ugly

ENV DEBUG=true
ENV OPTIMIZATIONS=false

COPY scripts/compile /
RUN chmod +x compile
RUN ["/compile"]

RUN apt-get update && apt-get install -y golang

# build prj
WORKDIR /workspace

COPY go.mod .
COPY go.sum .
RUN go mod download

# copy source
COPY cmd/ cmd/
COPY pkg/ pkg/

ARG TARGETPLATFORM=linux/amd64

RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then GOARCH=arm64; else GOARCH=amd64; fi && \
    CGO_ENABLED=1 GOOS=linux GOARCH=${GOARCH} GO111MODULE=on go build -a -o meeting-bot ./cmd
