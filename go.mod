module "https://github.com/telecom-tower/quote-of-the-day"

go 1.14

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/namsral/flag v1.7.4-pre
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.4.2
	github.com/telecom-tower/sdk v1.0.0-alpha.1
	github.com/telecom-tower/towerapi v1.0.0-alpha.2 // indirect
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200302123026-7795fca6ccb1 // indirect
	google.golang.org/grpc v1.27.1
)
