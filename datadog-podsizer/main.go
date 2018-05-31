package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"
)

var build = "0" // build number set at compile-time

// Config to hold Command configuration
type Config struct {
	DataDogAPIKey         string `json:"datadog_api_key"`
	DataDogApplicationKey string `json:"datadog_app_key"`
}

var cfg = new(Config)

func main() {
	app := cli.NewApp()
	app.Name = "datadog-podsizer"
	app.Version = fmt.Sprintf("0.1.%s", build)
	app.Usage = "Determine memory and cpu requests and limits based on DataDog queries"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "datadog-api-key",
			Usage:  "DataDog API Key",
			EnvVar: "DATADOG_API_KEY",
		},
		cli.StringFlag{
			Name:   "datadog-app-key",
			Usage:  "DataDog Application Key",
			EnvVar: "DATADOG_APP_KEY",
		},
		cli.StringFlag{
			Name:  "helm-release,r",
			Usage: "Helm release",
		}, cli.StringFlag{
			Name:  "kube-container,c",
			Usage: "Kubernetes containername",
		},
	}

	app.Action = run

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	cfg = &Config{
		DataDogAPIKey:         c.String("datadog-api-key"),
		DataDogApplicationKey: c.String("datadog-app-key"),
	}

	if len(c.String("helm-release")) == 0 || len(c.String("kube-container")) == 0 {
		return fmt.Errorf("Helm Release and Kube Container are mandatory parameters")
	}
	release := c.String("helm-release")
	container := c.String("kube-container")

	client := datadog.NewClient(cfg.DataDogAPIKey, cfg.DataDogApplicationKey)

	maxMem, err := maxMemory(client, release, container)
	if err != nil {
		return err
	}
	log.Printf("Max Memory: %.0f M in last week", maxMem)

	avgMem, err := avgMemory(client, release, container)
	if err != nil {
		return err
	}
	log.Printf("Avg Memory: %.0f M in last week", avgMem)

	maxnCore, err := maxNanoCore(client, release, container)
	if err != nil {
		return err
	}
	log.Printf("Max milliCore: %0.2f", maxnCore/1000000)

	avgnCore, err := avgNanoCore(client, release, container)
	if err != nil {
		return err
	}
	log.Printf("Avg milliCore: %0.2f", avgnCore/1000000)
	return nil
}

func maxMemory(client *datadog.Client, release, container string) (float64, error) {
	var max float64
	for i := 0; i < 7; i++ {
		dayMax, err := maxDayMemory(client, i, release, container)
		if err != nil {
			return 0, err
		}
		if max < dayMax {
			max = dayMax
		}
	}
	return max, nil
}

func maxDayMemory(client *datadog.Client, dayOffset int, release, container string) (float64, error) {
	query := fmt.Sprintf("max:kubernetes.memory.usage{helm_release:%s,kube_container_name:%s}", release, container)
	offset := time.Hour * 24 * time.Duration(dayOffset)
	duration := time.Hour * 24 // 1 day

	end := time.Now().Add(offset * -1)
	start := end.Add(duration * -1)

	//log.Printf("Querying metric. Start: %v, End: %v, query: %v", start, end, query)
	metrics, err := client.QueryMetrics(start.Unix(), end.Unix(), query)
	if err != nil {
		return 0, err
	}

	var max float64
	for _, m := range metrics {
		// log.Printf("Interval: %v - Aggr: %v", *m.Interval, *m.Aggr)
		for _, p := range m.Points {
			val := *p[1] / 1024 / 1024 //return Mb rounded up
			//log.Printf("  value: %.3f", val)
			if max < val+.44 {
				max = val + .44
			}
		}
	}
	return max, nil
}

func avgMemory(client *datadog.Client, release, container string) (float64, error) {
	query := fmt.Sprintf("avg:kubernetes.memory.usage{helm_release:%s,kube_container_name:%s}", release, container)
	duration := time.Hour * 24 // 1 day

	end := time.Now()
	start := end.Add(duration * -1)

	metrics, err := client.QueryMetrics(start.Unix(), end.Unix(), query)
	if err != nil {
		return 0, err
	}

	for _, m := range metrics {
		// log.Printf("Interval: %v - Aggr: %v", *m.Interval, *m.Aggr)
		var total float64
		for _, p := range m.Points {
			val := *p[1] / 1024 / 1024 //return Mb rounded up
			//log.Printf("  value: %.3f", val)
			total += val
		}
		return total / float64(len(m.Points)), err
	}
	return 0, nil
}

func maxNanoCore(client *datadog.Client, release, container string) (float64, error) {
	var max float64
	for i := 0; i < 7; i++ {
		dayMax, err := maxDayNanoCore(client, i, release, container)
		if err != nil {
			return 0, err
		}
		if max < dayMax {
			max = dayMax
		}
	}
	return max, nil
}

func maxDayNanoCore(client *datadog.Client, dayOffset int, release, container string) (float64, error) {
	query := fmt.Sprintf("max:kubernetes.cpu.usage.total{helm_release:%s,kube_container_name:%s}", release, container)
	offset := time.Hour * 24 * time.Duration(dayOffset)
	duration := time.Hour * 24 // 1 day

	end := time.Now().Add(offset * -1)
	start := end.Add(duration * -1)

	// log.Printf("Querying metric. Start: %v, End: %v, query: %v", start, end, query)
	metrics, err := client.QueryMetrics(start.Unix(), end.Unix(), query)
	if err != nil {
		return 0, err
	}

	var max float64
	for _, m := range metrics {
		// log.Printf("Day: %d - Interval: %v - Aggr: %v", dayOffset, *m.Interval, *m.Aggr)
		for _, p := range m.Points {
			val := *p[1]
			// log.Printf("  value: %.3f", val)
			if max < val {
				max = val
			}
		}
	}
	return max, nil
}

func avgNanoCore(client *datadog.Client, release, container string) (float64, error) {
	query := fmt.Sprintf("avg:kubernetes.cpu.usage.total{helm_release:%s,kube_container_name:%s}", release, container)
	duration := time.Hour * 24 // 1 day

	end := time.Now()
	start := end.Add(duration * -1)

	metrics, err := client.QueryMetrics(start.Unix(), end.Unix(), query)
	if err != nil {
		return 0, err
	}

	for _, m := range metrics {
		// log.Printf("Interval: %v - Aggr: %v", *m.Interval, *m.Aggr)
		var total float64
		for _, p := range m.Points {
			//log.Printf("  value: %.3f", val)
			total += *p[1]
		}
		return total / float64(len(m.Points)), err
	}
	return 0, nil
}
