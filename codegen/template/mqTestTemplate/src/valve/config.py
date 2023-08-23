from dotenv import load_dotenv
import os

load_dotenv()

# user config
OPEN_A_COMMAND = "M5"
OPEN_B_COMMAND = "M3"
INITIAL_POSITION = "A"
SERIAL_PORT = "loop://?logging=debug"

SERIAL_BAUD = 115200
SERIAL_TIMEOUT = 1

# system config
RABBITMQ_USER = os.getenv("RABBITMQ_USER")

RABBITMQ_PASSWORD = os.getenv("RABBITMQ_PASSWORD")
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST")
RABBITMQ_VHOST = os.getenv("RABBITMQ_VHOST")
RABBITMQ_URI = os.getenv("RABBITMQ_URI")
DEVICE_ID = os.getenv("DEVICE_ID")
INSTANCE_ID = os.getenv("INSTANCE_ID")
AMQP_EXCHANGE = os.getenv("AMQP_EXCHANGE")
