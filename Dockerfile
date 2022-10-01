FROM dunkbing/meeting-bot:base as base

FROM ubuntu:22.04

ARG TARGETPLATFORM

RUN set -eux; \
    apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y --no-install-recommends \
        bubblewrap \
        ca-certificates \
        iso-codes \
        ladspa-sdk \
        liba52-0.7.4 \
        libaa1 \
        libaom3 \
        libass9 \
        libavcodec58 \
        libavfilter7 \
        libavformat58 \
        libavutil56 \
        libbs2b0 \
        libbz2-1.0 \
        libcaca0 \
        libcap2 \
        libchromaprint1 \
        libcurl3-gnutls \
        libdca0 \
        libde265-0 \
        libdv4 \
        libdvdnav4 \
        libdvdread8 \
        libdw1 \
        libegl1 \
        libepoxy0 \
        libfaac0 \
        libfaad2 \
        libfdk-aac2 \
        libflite1 \
        libfluidsynth3 \
        libgbm1 \
        libgcrypt20 \
        libgl1 \
        libgles1 \
        libgles2 \
        libglib2.0-0 \
        libgme0 \
        libgmp10 \
        libgsl27 \
        libgsm1 \
        libgudev-1.0-0 \
        libharfbuzz-icu0 \
        libjpeg8 \
        libkate1 \
        liblcms2-2 \
        liblilv-0-0 \
        libmjpegutils-2.1-0 \
        libmodplug1 \
        libmp3lame0 \
        libmpcdec6 \
        libmpeg2-4 \
        libmpg123-0 \
        libofa0 \
        libogg0 \
        libopencore-amrnb0 \
        libopencore-amrwb0 \
        libopenexr25 \
        libopenjp2-7 \
        libopus0 \
        liborc-0.4-0 \
        libpango-1.0-0 \
        libpng16-16 \
        librsvg2-2 \
        librtmp1 \
        libsbc1 \
        libseccomp2 \
        libshout3 \
        libsndfile1 \
        libsoundtouch1 \
        libsoup2.4-1 \
        libspandsp2 \
        libspeex1 \
        libsrt1.4-openssl \
        libsrtp2-1 \
        libssl3 \
        libtag1v5 \
        libtheora0 \
        libtwolame0 \
        libunwind8 \
        libvisual-0.4-0 \
        libvo-aacenc0 \
        libvo-amrwbenc0 \
        libvorbis0a \
        libvpx7 \
        libvulkan1 \
        libwavpack1 \
        libwebp7 \
        libwebpdemux2 \
        libwebpmux3 \
        libwebrtc-audio-processing1 \
        libwildmidi2 \
        libwoff1 \
        libx264-163 \
        libx265-199 \
        libxkbcommon0 \
        libxslt1.1 \
        libzbar0 \
        libzvbi0 \
        mjpegtools \
        intel-media-va-driver-non-free libva2 vainfo \
        quelcom flvmeta \
        fonts-takao-mincho \
        xdg-dbus-proxy && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# install deps
RUN apt-get update && \
    apt-get install -y \
    curl \
    gnupg \
    gstreamer1.0-pulseaudio gstreamer1.0-tools \
    dbus-x11 xserver-xorg-video-dummy \
    pulseaudio \
    unzip \
    wget \
    xvfb

# install chrome
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
    apt remove chromium-browser chromium-browser-l10n chromium-codecs-ffmpeg-extra && \
    echo "deb http://deb.debian.org/debian buster main \
          deb http://deb.debian.org/debian buster-updates main \
          deb http://deb.debian.org/debian-security buster/updates main" >> /etc/apt/sources.list.d/debian.list && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys DCC9EFBF77E11517 && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 648ACFD622F3D138 && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys AA8E81B4331F7F50 && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 112695A0E562B32A && \
    echo 'Package: * \
          Pin: release a=eoan \
          Pin-Priority: 500 \
          \
          \
          Package: * \
          Pin: origin "deb.debian.org" \
          Pin-Priority: 300 \
          \
          \
          Package: chromium* \
          Pin: origin "deb.debian.org" \
          Pin-Priority: 700' >> /etc/apt/preferences.d/chromium.pref && \
    apt update && \
    apt install -y chromium \
    ; else \
    wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
    apt-get install -y ./google-chrome-stable_current_amd64.deb && \
    rm google-chrome-stable_current_amd64.deb \
    ; fi

# clean
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# add root user to group for pulseaudio access
RUN adduser root pulse-access

# create xdg_runtime_dir
RUN mkdir -pv ~/.cache/xdgr

COPY --from=base /compiled-binaries /
COPY --from=base /workspace/meeting-bot /

RUN mkdir /data

COPY scripts/entrypoint .
RUN chmod +x entrypoint
ENTRYPOINT ["./entrypoint"]
