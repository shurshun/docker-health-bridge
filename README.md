# Docker Health Bridge

[![Go Report Card](https://goreportcard.com/badge/github.com/shurshun/docker-health-bridge)](https://goreportcard.com/report/github.com/shurshun/docker-health-bridge) [![Docker Automated buil](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/shurshun/docker-health-bridge/) [![Join the chat at https://gitter.im/shurshun/Lobby](https://badges.gitter.im/shurshun/Lobby.svg)](https://gitter.im/shurshun/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link)

This util checks containers health status and sends it to [sensu api](https://sensuapp.org/docs/0.29/api/results-api.html#results-post)

## Configuration

* `--sensu-api`, `-s` - Sensu API host (default: "sensu-api:4567") [env SENSU_API]
* `--hostname`, `-n` - Hostname to use for events [env HOSTNAME]
* `--retries`, `r` - Set occurrences for sensu check before triggering an alert notification. Using Docker healthcheck param `--retries` (default: 3) [env RETRIES]
* `--log-level`, `-l` - Set logging level: info, warning, error, fatal, debug, panic (default: warning) [env LOG_LEVEL]

Docker client initiliazing by env params DOCKER_HOST, DOCKER_TLS and other.
See [cli/#environment-variables](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables) for more details.

## Installation

`docker pull shurshun/docker-health-bridge`

## Running

```
docker run -d \
  -h $HOSTNAME \
  --name="docker-health-bridge" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e SENSU_API=<...> \
  shurshun/docker-health-bridge
```

## Building

`make compile`

By default Linux and Mac OS builds are made.

You will need a local docker instance running that supports mounting in your host volume path into the container. If you are on OSX, this can be achieved using docker-machine (with virtualbox or vmware fusion). If you are on linux, and running a local docker daemon this is already supported.