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
    return (
        f"{config.INSTANCE_ID}.events.{event.__class__.__name__}"
    )


def publish(event: events.Event):
    channel.basic_publish(
        exchange=config.AMQP_EXCHANGE, routing_key=routing_key,
        body=json.dumps(asdict(event)))

    logging.info(f"Publishing event {event}")
