## Overview
This project is an application that automatically join Google Meet, Microsoft Teams and Zoom.
And records audio/video that can be saved to a local file or stream to other services.

### Architecture
Following libraries used for the project:
- **Rod** a devtools driver for web automation and scraping (written in golang). We use it to find buttons, textbox on a browser page to enter necessaries information for the meeting.
- **pulseaudio** server for audio playback.
- **xvfb** for faked screen display. We use it to open chrome inside a container.
- **GStreamer** and **go-gst** for recording audio from pulseaudio source.

### Features
- [ ] Supports Zoom
- [x] Supports Google Meet
- [x] Supports Microsoft Teams
- [x] Write audio to file /data/file.mp3 or /data/file.mp4
- [ ] Supports streaming audio. Need to integrate with live text asr
- [ ] Auto stop if no one is in the meeting for 5 minutes
- [ ] Auto stop if the meeting is ended

### Building
```bash
make docker-gstreamer
make docker-prod
```

### Running
```bash
docker run --rm --name meeting-bot -e CONFIG_BODY="$(cat config.yaml)" -v ~/meeting-bot/recordings:/data dunkbing/meeting-bot
```
