import json
import logging
from dataclasses import asdict

import pika

from valve import config
from valve.domain import events

logger = logging.getLogger(__name__)

connection = pika.BlockingConnection(
    pika.ConnectionParameters(
        virtual_host=config.RABBITMQ_VHOST,
        host=config.RABBITMQ_HOST,
        credentials=pika.PlainCredentials(
            username=config.RABBITMQ_USER,
            password=config.RABBITMQ_PASSWORD
        )
    )
)

channel = connection.channel()

channel.exchange_declare(exchange=config.AMQP_EXCHANGE, exchange_type='topic')


def routing_key(event: events.Event) -> str:
    match type(event):
        case events.DeviceRetrieved:
            return (
                f"{config.INSTANCE_ID}.device.{event}"
            )
        case events.StateRetrieved:
            return (
                f"{config.INSTANCE_ID}.state.{event}"
            )
        case _:
            return (
                f"{config.INSTANCE_ID}.events.{event}"
            )


def publish(event: events.Event):
    channel.basic_publish(
        exchange=config.AMQP_EXCHANGE, routing_key=routing_key(event),
        body=json.dumps(asdict(event)).encode('utf-8'))

    logging.info(f"Publishing event {event}")
