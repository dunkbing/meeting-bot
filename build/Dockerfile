FROM dunkbing/meeting-bot:gst-1.18.5-dev

ARG TARGETPLATFORM

WORKDIR /workspace

RUN apt-get update && apt-get install -y golang

# download go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# copy source
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then GOARCH=arm64; else GOARCH=amd64; fi && \
    CGO_ENABLED=1 GOOS=linux GOARCH=${GOARCH} GO111MODULE=on go build -a -o meeting-bot ./cmd/server

FROM dunkbing/meeting-bot:gst-1.18.5-prod

ARG TARGETPLATFORM

# install deps
RUN apt-get update && \
    apt-get install -y \
    curl \
    gnupg \
    gstreamer1.0-pulseaudio \
    pulseaudio \
    unzip \
    wget \
    xvfb

# install chrome
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
RUN apt-get update
RUN apt-get -y install google-chrome-stable

# install chromedriver
# NOTE: this is a relatively old version.  Try a newer version if this does not work.
RUN wget -N http://chromedriver.storage.googleapis.com/2.25/chromedriver_linux64.zip
RUN unzip chromedriver_linux64.zip
RUN chmod +x chromedriver
RUN mv -f chromedriver /usr/local/bin/chromedriver

# clean
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# add root user to group for pulseaudio access
RUN adduser root pulse-access

# create xdg_runtime_dir
RUN mkdir -pv ~/.cache/xdgr

# copy binary
COPY --from=0 /workspace/meeting-bot /
COPY build/entrypoint .

ENTRYPOINT ["./entrypoint"]
