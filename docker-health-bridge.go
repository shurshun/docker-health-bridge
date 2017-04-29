package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os"
	"strings"
)

const version = "1.0.4"

type sensuCheckResult struct {
	Source      string  `json:"source"`
	Name        string  `json:"name"`
	Output      string  `json:"output"`
	Status      int     `json:"status"`
	Duration    float64 `json:"duration"`
	Occurrences int     `json:"occurrences"`
}

var (
	dockerClient *client.Client
)

func sendToSensu(c *cli.Context, payload []byte) {
	sensuApi := fmt.Sprintf("http://%s/results", c.String("sensu-api"))

	req, err := http.NewRequest("POST", sensuApi, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer resp.Body.Close()
}

func genPayload(source string, name string, output string, status int, duration float64, occurrences int) []byte {

	checkResult := &sensuCheckResult{Source: source,
		Name:        name,
		Output:      output,
		Status:      status,
		Duration:    duration,
		Occurrences: occurrences}

	payload, err := json.Marshal(checkResult)

	if err != nil {
		log.Fatal(err.Error())
	}

	return payload
}

func getHostname(c *cli.Context, conf *container.Config) string {
	if c.String("hostname") != "" {
		return c.String("hostname")
	}

	return conf.Hostname
}

func getRetries(c *cli.Context, conf *container.Config) int {
	if c.Int("retries") > 0 {
		return c.Int("retries")
	}

	if conf.Healthcheck.Retries > 0 {
		return conf.Healthcheck.Retries
	}

	return 3
}

func getLastState(health *types.Health) *types.HealthcheckResult {
	return health.Log[len(health.Log)-1]
}

func inspectContainer(c *cli.Context, id string) {
	info, err := dockerClient.ContainerInspect(context.Background(), id)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if len(info.State.Health.Log) > 0 {
		state := getLastState(info.State.Health)

		payload := genPayload(
			getHostname(c, info.Config),
			strings.TrimPrefix(info.Name, "/"),
			state.Output,
			state.ExitCode,
			state.End.Sub(state.Start).Seconds(),
			getRetries(c, info.Config))

		log.Info(string(payload))

		sendToSensu(c, payload)
	}
}

func listenEvents(c *cli.Context) {
	var err error

	dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	filters := filters.NewArgs()
	filters.Add("type", events.ContainerEventType)
	filters.Add("event", "exec_start")
	filters.Add("event", "health_status")

	messages, errs := dockerClient.Events(context.Background(), types.EventsOptions{Filters: filters})

	for {
		select {
		case err := <-errs:
			if err != nil && err != io.EOF {
				log.Warn(err.Error())
			}
		case e := <-messages:
			inspectContainer(c, e.ID)
		}
	}
}

func initLogging(c *cli.Context) {
	logLevel, err := log.ParseLevel(c.String("log-level"))

	if err != nil {
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(logLevel)
	}

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
}

func main() {
	app := cli.NewApp()

	app.Version = version
	app.Name = "docker-health-bridge"
	app.Usage = "Bridge container health to sensu"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "sensu-api, s",
			Value:  "sensu-api:4567",
			Usage:  "Sensu API host",
			EnvVar: "SENSU_API",
		},
		cli.StringFlag{
			Name:   "hostname, n",
			Usage:  "Hostname to use for events",
			EnvVar: "HOSTNAME",
		},
		cli.IntFlag{
			Name:   "retries, r",
			Value:  0,
			Usage:  "Retries before triggering an alert notification",
			EnvVar: "RETRIES",
		},
		cli.StringFlag{
			Name:   "log-level, l",
			Value:  "warning",
			Usage:  "Set logging level: info, warning, error, fatal, debug, panic",
			EnvVar: "LOG_LEVEL",
		},
	}

	app.Action = func(c *cli.Context) {
		initLogging(c)
		listenEvents(c)
	}

	app.Run(os.Args)
}
