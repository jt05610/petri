package env

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"os"
)

type Environment struct {
	URI        string
	Exchange   string
	DeviceID   string
	InstanceID string
	RPCAddress string
}

func LoadEnv(logger *zap.Logger) *Environment {
	err := godotenv.Load()
	failOnError(err, "Error loading .env file")

	logger.Info("Starting 🐰 server")
	// Setup rabbitmq channel
	uri, ok := os.LookupEnv("RABBITMQ_URI")
	if !ok {
		logger.Fatal("RABBITMQ_URI not set")
	}
	exchange, ok := os.LookupEnv("AMQP_EXCHANGE")
	if !ok {
		logger.Fatal("AMQP_EXCHANGE not set")
	}
	deviceID, ok := os.LookupEnv("DEVICE_ID")
	if !ok {
		logger.Fatal("DEVICE_ID not set")
	}
	instanceID, ok := os.LookupEnv("INSTANCE_ID")
	if !ok {
		logger.Fatal("INSTANCE_ID not set")
	}
	rpcAddr, found := os.LookupEnv("RPC_ADDRESS")
	if !found {
		logger.Fatal("RPC_ADDRESS not set")
	}
	return &Environment{
		URI:        uri,
		Exchange:   exchange,
		DeviceID:   deviceID,
		InstanceID: instanceID,
		RPCAddress: rpcAddr,
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
