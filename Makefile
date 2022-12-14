default: help

# Build Docker image
build: docker-gstreamer docker-prod output

# Build and push Docker image
release: build docker-push output

# Image and binary can be overridden with env vars.
DOCKER_IMAGE ?= dunkbing/meeting-bot

# Get the latest commit.
GIT_COMMIT = $(strip $(shell git rev-parse --short HEAD))

# Get the version number from the code
CODE_VERSION = $(strip $(shell cat VERSION))

GSTREAMER_VERSION = 1.18.5

# Find out if the working directory is clean
GIT_NOT_CLEAN_CHECK = $(shell git status --porcelain)

# If we're releasing to Docker Hub, and we're going to mark it with the latest tag, it should exactly match a version release
ifeq ($(MAKECMDGOALS),release)
# Use the version number as the release tag.
DOCKER_TAG = $(GIT_COMMIT)

ifndef CODE_VERSION
$(error You need to create a VERSION file to build a release)
endif

# See what commit is tagged to match the version
VERSION_COMMIT = $(strip $(shell git rev-list $(CODE_VERSION) -n 1 | cut -c1-7))
ifneq ($(VERSION_COMMIT), $(GIT_COMMIT))
$(error echo You are trying to push a build based on commit $(GIT_COMMIT) but the tagged release version is $(VERSION_COMMIT))
endif

# Don't push to Docker Hub if this isn't a clean repo
ifneq (x$(GIT_NOT_CLEAN_CHECK), x)
$(error echo You are trying to release a build based on a dirty repo)
endif

else
# Add the commit ref for development builds. Mark as dirty if the working directory isn't clean
DOCKER_TAG = $(GIT_COMMIT)
endif

help:
	@echo "    build"
	@echo "        Build a docker production image."
	@echo "    test"
	@echo "        Run tests"
	@echo "    clean"
	@echo "        Clean project."

clean:
	rm -fr ./bin
	rm -fr ./out

docker-gstreamer:
	docker build -t $(DOCKER_IMAGE):gst-$(GSTREAMER_VERSION)-base --build-arg GSTREAMER_VERSION=$(GSTREAMER_VERSION) -f build/gstreamer/Dockerfile-base ./build/gstreamer
	docker build -t $(DOCKER_IMAGE):gst-$(GSTREAMER_VERSION)-dev --build-arg GSTREAMER_VERSION=$(GSTREAMER_VERSION) -f build/gstreamer/Dockerfile-dev ./build/gstreamer
	docker build -t $(DOCKER_IMAGE):gst-$(GSTREAMER_VERSION)-prod --build-arg GSTREAMER_VERSION=$(GSTREAMER_VERSION) -f build/gstreamer/Dockerfile-prod ./build/gstreamer

docker-prod:
	docker build \
  --build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
  --build-arg VERSION=$(CODE_VERSION) \
  --build-arg VCS_URL=`git config --get remote.origin.url` \
  --build-arg VCS_REF=$(GIT_COMMIT) \
  -t $(DOCKER_IMAGE):latest -f build/Dockerfile .

docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

output:
	@echo Docker Image: $(DOCKER_IMAGE):$(DOCKER_TAG)

run:
	docker run --rm --name meeting-bot -e CONFIG_BODY="$(cat config.yaml)" -e RECORDING_REQUEST="$(cat request.json)" -v ~/meeting-bot/recordings:/data dunkbing/meeting-bot

.PHONY: clean
