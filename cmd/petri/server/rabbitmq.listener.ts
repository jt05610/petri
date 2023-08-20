import * as dotenv from "dotenv";
import amqp from "amqplib/callback_api";
import invariant from "tiny-invariant";
import { singleton } from "./singleton.server";

dotenv.config();

invariant(process.env.RABBITMQ_URI, "RABBITMQ_URI is required");
invariant(process.env.RABBITMQ_EXCHANGE, "RABBITMQ_EXCHANGE is required");
const uri = process.env.RABBITMQ_URI;
const exchange = process.env.RABBITMQ_EXCHANGE;

export const listen = (callback: (msg: amqp.Message | null) => void) => {
  return singleton("rabbitmq.listener", () => {
    return amqp.connect(uri, function(error0, connection) {
      if (error0) {
        throw error0;
      }
      return connection.createChannel(function(error1, channel) {
        if (error1) {
          throw error1;
        }

        channel.assertExchange(exchange, "topic", {
          durable: false
        });

        channel.assertQueue("", {
          exclusive: true
        }, function(error2, q) {
          if (error2) {
            throw error2;
          }
          console.log(" [*] Waiting for events. To exit press CTRL+C");

          channel.bindQueue(q.queue, exchange, "*.events.*");
          channel.consume(q.queue, callback, {
            noAck: true
          });
          return channel;
        });
      });
    });
  });
};
