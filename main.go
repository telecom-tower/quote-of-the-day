package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/namsral/flag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/telecom-tower/sdk"
	"golang.org/x/image/colornames"
	"google.golang.org/grpc"
	"net/http"
	"net/url"
)

const (
	quotesServer = "http://quotes.rest/qod.json"
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

	log.Infof("Connecting to tower server : %v", *grpcUrl)
	conn, err := grpc.Dial(*grpcUrl, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error dialing server: %v", err)
	}
	defer conn.Close() // nolint: errcheck
	client := sdk.NewClient(conn)

	u, _ := url.Parse(quotesServer)
	if *category != "" {
		q := u.Query()
		q.Set("category", *category)
		u.RawQuery = q.Encode()
	}

	log.Infof("Connecting to quote-of-the-day server: %v", u)
	res, err := http.Get(u.String())
	check(err, "Error connecting to server")
	defer res.Body.Close()

	q := qod{}
	err = json.NewDecoder(res.Body).Decode(&q)
	check(err, "Unable to decode quote")

	if q.Success.Total < 1 {
		log.Fatal("Invalid quote")
	}

	check(client.StartDrawing(context.Background()), "Error getting frame")
	check(client.Init(), "Error initializing display")

	log.Infof("Quote: \"%v\"", q.Contents.Quotes[0].Quote)
	log.Infof("Author: \"%v\"", q.Contents.Quotes[0].Author)

	format := "<text><font color=\"dogerblue\">%s</font> <font color=\"gold\">(%s)</font> <font color=\"lime\">&gt;&gt;&gt;</font> </text>"
	msg := fmt.Sprintf(format, q.Contents.Quotes[0].Quote, q.Contents.Quotes[0].Author)
	check(client.WriteText(msg, "6x8", 0, colornames.Dodgerblue, 0, sdk.PaintMode), "Error writing text")
	check(client.AutoRoll(0, sdk.RollingNext, 0, 0), "Error setting autoroll")
	check(client.Render(), "Error rendering")

	log.Debug("Done")
}
