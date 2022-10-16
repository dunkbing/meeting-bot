FROM dunkbing/meeting-bot:gst-1.18.5-dev

WORKDIR /workspace

ARG TARGETPLATFORM

# install deps
RUN apt-get update && \
    apt-get install -y \
    curl \
    ffmpeg \
    gnupg \
    golang \
    gstreamer1.0-pulseaudio \
    pulseaudio \
    unzip \
    wget \
    xvfb


RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then ARCH=arm64v8; else ARCH=amd64; fi && \
    wget https://github.com/aler9/rtsp-simple-server/releases/download/v0.17.6/rtsp-simple-server_v0.17.6_linux_${ARCH}.tar.gz && \
    tar -zxvf rtsp-simple-server_v0.17.6_linux_${ARCH}.tar.gz && \
    rm rtsp-simple-server_v0.17.6_linux_${ARCH}.tar.gz

# install chrome
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
RUN apt-get update
RUN apt-get -y install google-chrome-stable

# clean up
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# add root user to group for pulseaudio access
RUN adduser root pulse-access

# create xdg_runtime_dir
RUN mkdir -pv ~/.cache/xdgr

# download go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# copy source
COPY pkg/ pkg/
COPY test/ test/
COPY build/test/entrypoint .

# run
ENTRYPOINT ["./entrypoint"]
