## Overview
This project is an application that automatically joins Google Meet, Microsoft Teams, and Zoom.
And records audio/video that can be saved to a local file or streamed to other services.

### Architecture
Following libraries used for the project:
- **Rod** a dev-tools driver for web automation and scraping (written in Golang). We use it to find buttons, and text-boxes on a browser page to enter the necessary information for the meeting.
- **PulseAudio** server for audio playback.
- **Xvfb** for faked screen display. We use it to open chrome inside a container.
- **GStreamer** and **go-gst** for recording audio from pulseaudio source.

### Features
- [ ] Supports Zoom
- [x] Supports Google Meet
- [x] Supports Microsoft Teams
- [x] Write audio to file /data/file.mp3 or /data/file.mp4
- [ ] Supports streaming audio.
- [ ] Auto-stop if no one is in the meeting for 5 minutes
- [ ] Auto-stop if the meeting is ended

### Building
```bash
make docker-gstreamer
make docker-prod
```

### Running
```bash
docker run --rm --name meeting-bot -e CONFIG_BODY="$(cat config.yaml)" -v ~/meeting-bot/recordings:/data dunkbing/meeting-bot
```
