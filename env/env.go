package env

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
)

type Environment struct {
	URI        string
	Exchange   string
	DeviceID   string
	InstanceID string
	SerialPort string
	Baud       int
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
	serPort, found := os.LookupEnv("SERIAL_PORT")
	if !found {
		logger.Fatal("SERIAL_PORT not set")
	}
	baud, found := os.LookupEnv("SERIAL_BAUD")
	if !found {
		logger.Fatal("SERIAL_BAUD not set")
	}
	baudInt, err := strconv.ParseInt(baud, 10, 64)
	if err != nil {
		logger.Fatal("Failed to parse baud", zap.Error(err))
	}
	return &Environment{
		URI:        uri,
		Exchange:   exchange,
		DeviceID:   deviceID,
		InstanceID: instanceID,
		SerialPort: serPort,
		Baud:       int(baudInt),
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
