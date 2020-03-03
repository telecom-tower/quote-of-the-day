package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/namsral/flag"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/telecom-tower/sdk"
	"golang.org/x/image/colornames"
	"google.golang.org/grpc"
)

const (
	quotesServer = "http://quotes.rest/qod.json"
	secret       = "OneFlagToRuleThemAll"
	mqttTopic    = "heiafriscctf"
)

type success struct {
	Total int `json:"total"`
}

type quote struct {
	Quote      string   `json:"quote"`
	Length     string   `json:"length"`
	Author     string   `json:"author"`
	Tags       []string `json:"tags"`
	Category   string   `json:"category"`
	Date       string   `json:"date"`
	Title      string   `json:"title"`
	Background string   `json:"background"`
	ID         string   `json:"id"`
}

type contents struct {
	Quotes    []quote `json:"quotes"`
	Copyright string  `json:"copyright"`
}

type qod struct {
	Success  success  `json:"success"`
	Contents contents `json:"contents"`
}

func check(err error, msg string) {
	if err != nil {
		err = errors.WithMessage(err, msg)
		log.Fatal(err)
	}
}

func getQod(category string) (*qod, error) {
	u, _ := url.Parse(quotesServer)
	if category != "" {
		q := u.Query()
		q.Set("category", category)
		u.RawQuery = q.Encode()
	}

	log.Infof("Connecting to quote-of-the-day server: %v", u)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, errors.WithMessage(err, "Error connecting to server")
	}
	defer res.Body.Close()

	q := qod{}
	err = json.NewDecoder(res.Body).Decode(&q)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to decode quote")
	}

	if q.Success.Total < 1 {
		return nil, errors.WithMessage(err, "Invalid quote")
	}

	log.Infof("Quote: \"%v\"", q.Contents.Quotes[0].Quote)
	log.Infof("Author: \"%v\"", q.Contents.Quotes[0].Author)

	return &q, nil
}

func updateDisplay(client *sdk.Client, q *qod) {
	log.Debug("Updating Display")
	check(client.StartDrawing(context.Background()), "Error getting frame")
	check(client.Init(), "Error initializing display")
	format := "<text><font color=\"dogerblue\">%s</font> <font color=\"gold\">(%s)</font> <font color=\"lime\">&gt;&gt;&gt;</font> </text>"
	msg := fmt.Sprintf(format, q.Contents.Quotes[0].Quote, q.Contents.Quotes[0].Author)
	check(client.WriteText(msg, "6x8", 0, colornames.Dodgerblue, 0, sdk.PaintMode), "Error writing text")
	check(client.AutoRoll(0, sdk.RollingNext, 0, 0), "Error setting autoroll")
	check(client.Render(), "Error rendering")
	log.Debug("Done")
}

func showSecret(client *sdk.Client, secret string) {
	log.Debug("Show secret")
	for _, c := range secret {
		check(client.StartDrawing(context.Background()), "Error getting frame")
		check(client.Init(), "Error initializing display")
		check(client.WriteText(string(c), "8x8", 60, colornames.Orange, 0, sdk.PaintMode), "Error writing text")
		check(client.AutoRoll(0, sdk.RollingStop, 0, 0), "Error setting autoroll")
		check(client.Render(), "Error rendering")

		time.Sleep(70 * time.Millisecond)

		check(client.StartDrawing(context.Background()), "Error getting frame")
		check(client.Init(), "Error initializing display")
		check(client.Render(), "Error rendering")

		time.Sleep(150 * time.Millisecond)
	}
	log.Debug("Done")
}

func serve(client *sdk.Client) (chan *qod, chan [2]string) {

	qodQueue := make(chan *qod, 1)
	secretQueue := make(chan [2]string, 1)

	go func() {
		for {
			select {
			case q := <-qodQueue:
				updateDisplay(client, q)

			case <-secretQueue:
				showSecret(client, secret)
			}
		}
	}()
	return qodQueue, secretQueue
}

func main() {
	debug := flag.Bool("debug", false, "Run in debug mode")
	category := flag.String("category", "students", "quote category")
	grpcUrl := flag.String("url", "localhost:10000", "grpc URL")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	var q *qod
	var err error

	log.Infof("Connecting to tower server : %v", *grpcUrl)
	conn, err := grpc.Dial(*grpcUrl, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error dialing server: %v", err)
	}
	defer conn.Close() // nolint: errcheck

	qodQueue, secretQueue := serve(sdk.NewClient(conn))

	q, err = getQod(*category)
	check(err, "Error getting Quote")
	qodQueue <- q

	c := cron.New()
	c.AddFunc("*/5 * * * *", func() {
		qodQueue <- q
	})
	c.AddFunc("2,32 * * * *", func() {
		q, err = getQod(*category)
		if err != nil {
			log.Warn("Error updating quote : %v", err)
		}
	})
	c.Start()

	opts := MQTT.NewClientOptions()
	opts.AddBroker("mqtt.eclipse.org:1883")
	opts.SetCleanSession(true)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		secretQueue <- [2]string{msg.Topic(), string(msg.Payload())}
		qodQueue <- q
	})

	mqttClient := MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := mqttClient.Subscribe(mqttTopic, byte(0), nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	<-done
	c.Stop()

	log.Infof("Finished")
}
