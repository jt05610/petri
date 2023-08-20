import invariant from "tiny-invariant";
import type { Connection } from "amqplib";
import { connect } from "amqplib";
import { singleton } from "./singleton.server";

const publisher = singleton("amqp", async () => {
  invariant(process.env.RABBITMQ_URI, "RABBITMQ_URI is required");
  return connect(process.env.RABBITMQ_URI, function(error0: Error, connection: Connection) {
      if (error0) {
        throw error0;
      }
      return connection;
    }
  ).then(async (connection) => {
    return connection.createChannel().then((channel) => {
      invariant(process.env.RABBITMQ_EXCHANGE, "RABBITMQ_URI is required");
      channel.assertExchange(process.env.RABBITMQ_EXCHANGE, "topic", {
        durable: false
      });
      return channel;
    });
  });
});

export { publisher };
