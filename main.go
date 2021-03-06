package main

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"google.golang.org/api/gmail/v1"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Interval int      `yaml:"interval"`
	Labels   []string `yaml:"labels"`
}

func recordMetrics(interval int, unreadGauge *prometheus.GaugeVec, totalGauge *prometheus.GaugeVec, labelIds []string, srv *gmail.Service) {
	go func() {
		for {
			fmt.Printf("scraping %d labels\n", len(labelIds))
			for _, labelId := range labelIds {
				label, err := srv.Users.Labels.Get("me", labelId).Do()
				if err != nil {
					fmt.Printf("%v", err)
				} else {
					prometheusLabels := map[string]string{"Label": "gmail_" + strcase.ToSnake(label.Name)}
					totalGauge.With(prometheusLabels).Set(float64(label.ThreadsTotal))
					unreadGauge.With(prometheusLabels).Set(float64(label.ThreadsUnread))
				}
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
}

func main() {
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("could not parse config file: %v", err)
	}

	srv := createGmailService()

	var labelIds []string
	for _, lab := range getLabels(srv) {
		for _, desiredLabel := range config.Labels {
			if lab.Name == desiredLabel {
				labelIds = append(labelIds, lab.Id)
				break
			}
		}
	}

	unreadGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gmail_threads_unread",
			Help: "number of unread threads",
		},
		[]string{"Label"},
	)
	totalGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gmail_threads_total",
			Help: "total number of threads",
		},
		[]string{"Label"},
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(unreadGauge, totalGauge)

	recordMetrics(config.Interval, unreadGauge, totalGauge, labelIds, srv)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fmt.Println("http://localhost:2112/metrics")
	log.Fatal(http.ListenAndServe(":2112", nil))
}
