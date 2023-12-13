package pump_bank

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	proto "github.com/jt05610/petri/comm/grbl/proto/v1"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

//go:embed device.yaml
var deviceYaml embed.FS

func loadPumpParams() *BankSettings {
	f, err := os.Open("pump_params.json")
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create("pump_params.json")
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				err := f.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
			bs := &BankSettings{
				Aqueous: &Settings{
					SyringeDiameter: 0,
					SyringeVolume:   0,
					MaxDistance:     0,
					MaxFeedRate:     0,
				},
				Organic: &Settings{
					SyringeDiameter: 0,
					SyringeVolume:   0,
					MaxDistance:     0,
					MaxFeedRate:     0,
				},
			}

			bytes, err := json.MarshalIndent(bs, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			_, err = f.Write(bytes)
			if err != nil {
				log.Fatal(err)
			}
			panic("Please edit pump_params.json and restart the server")
		}
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	var bs BankSettings
	err = json.NewDecoder(f).Decode(&bs)
	if err != nil {
		log.Fatal(err)
	}
	return &bs
}

func Run(ctx context.Context, conn *amqp.Connection, client proto.GRBLServer) {
	req := loadPumpParams()
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	d := NewPumpBank(client)
	d.valveCanDispense.Store(true)
	dev := d.load()
	err = d.Initialize(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	failOnError(err, "Failed to initialize device")
	if err != nil {
		log.Fatal(err)
	}
	srv := server.New(dev.Nets[0], conn.Channel, environ.Exchange, environ.DeviceID, environ.InstanceID, dev.EventMap(), d.Handlers(), logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	err = d.positionOrganicValve(ctx, &StartPumpRequest{
		Volume: -5,
	})
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Started 🐰 server")
	srv.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
