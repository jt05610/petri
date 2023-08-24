import json
import logging

import pika

from valve import config, bootstrap
from valve.domain import commands
from valve.service_layer.handlers import COMMAND_HANDLERS

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)

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

queue = channel.queue_declare(queue='', exclusive=True)


def main():
    logger.info('Starting RabbitMQ consumer')
    bus = bootstrap.bootstrap()
    for key in COMMAND_HANDLERS.keys():
        channel.queue_bind(
            exchange=config.AMQP_EXCHANGE,
            queue=queue.method.queue,
            routing_key=f"{config.INSTANCE_ID}.commands.{key}"
        )
    channel.queue_bind(
        exchange=config.AMQP_EXCHANGE,
        queue=queue.method.queue,
        routing_key=f"devices"
    )
    channel.queue_bind(
        exchange=config.AMQP_EXCHANGE,
        queue=queue.method.queue,
        routing_key=f"{config.INSTANCE_ID}.state.get"
    )

    def callback(ch, method, properties: pika.BasicProperties, body):
        match method.routing_key:
            case "devices":
                cmd = commands.GetDevice()
                bus.handle(cmd)
            case _:
                logger.debug(f"Received message {body}")
                split = method.routing_key.split(".")
                topic = split[1]
                k = split[2]
                match topic:
                    case "state":
                        logger.info(f"Received state request")
                        cmd = commands.GetState()

                    case "commands":
                        logger.info(f"Received command {k}< {body} >")
                        match k:
                            case "open_a":
                                cmd = commands.OpenA()
                            case "open_b":
                                cmd = commands.OpenB()
                            case "flow":
                                body = json.loads(body)
                                cmd = commands.Flow(**body)
                            case _:
                                logger.error(f"Unknown command {k}")
                                return
                    case _:
                        logger.error(f"Unknown topic {topic}")
                        return

        bus.handle(cmd)

    logger.info('Waiting for messages')

    channel.basic_consume(
        queue=queue.method.queue,
        on_message_callback=callback,
        auto_ack=True
    )

    channel.start_consuming()


if __name__ == '__main__':
    main()
