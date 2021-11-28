package main

import (
	"fmt"
	"google.golang.org/api/gmail/v1"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Interval int      `yaml:"interval"`
	Labels   []string `yaml:"labels"`
}

type GmailGaugeType int

const (
	Total GmailGaugeType = iota
	Unread
)

type GaugeConfig struct {
	Type  GmailGaugeType
	Gauge prometheus.Gauge
}

func createGauges(config Config, labelIdsByName map[string]string) map[string][]GaugeConfig {
	var gaugeConfigByLabelId = make(map[string][]GaugeConfig)
	for _, l := range config.Labels {
		var gauges []GaugeConfig
		gauges = append(gauges, GaugeConfig{
			Type:    Total,
			Gauge: promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("gmail_threads_%s_total", strcase.ToSnake(l)),
				Help: fmt.Sprintf("total number of threads with the label %s", l),
			}),
		})
		gauges = append(gauges, GaugeConfig{
			Type:    Unread,
			Gauge: promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("gmail_threads_%s_unread", strcase.ToSnake(l)),
				Help: fmt.Sprintf("number of unread threads with the label %s", l),
			}),
		})
		labelId := labelIdsByName[l]
		gaugeConfigByLabelId[labelId] = gauges
	}
	return gaugeConfigByLabelId
}

func recordMetrics(interval int, gaugeConfigByLabelId map[string][]GaugeConfig, srv *gmail.Service) {
	go func() {
		for {
			for labelId, gaugeConfigs := range gaugeConfigByLabelId {
				label, err := srv.Users.Labels.Get("me", labelId).Do()
				if err != nil {
					fmt.Printf("%v", err)
				} else {
					for _, gaugeConfig := range gaugeConfigs {
						switch gaugeConfig.Type {
						case Total:
							gaugeConfig.Gauge.Set(float64(label.ThreadsTotal))
						case Unread:
							gaugeConfig.Gauge.Set(float64(label.ThreadsUnread))
						}
					}
				}
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
}

func _main() {
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

	labelIdsByName := make(map[string]string)
	for _, lab := range getLabels(srv) {
		labelIdsByName[lab.Name] = lab.Id
	}

	gaugeConfigByLabel := createGauges(config, labelIdsByName)
	registry := prometheus.NewRegistry()
	var collectors []prometheus.Collector
	for _, gaugeConfigs := range gaugeConfigByLabel {
		for _, gaugeConfig := range gaugeConfigs {
			collectors = append(collectors, gaugeConfig.Gauge)
		}
	}
	registry.MustRegister(collectors...)

	recordMetrics(config.Interval, gaugeConfigByLabel, srv)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fmt.Println("http://localhost:2112/metrics")
	log.Fatal(http.ListenAndServe(":2112", nil))
}

func main() {
	http.HandleFunc("/", HelloServer)
	fmt.Println("http://localhost:2112/")
	log.Fatal(http.ListenAndServe(":2112", nil))
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}
