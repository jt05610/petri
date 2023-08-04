import { publisher } from "../../server/amqp.publisher";
import { z } from "zod";

export const SayHelloInputSchema = z.object({
  deviceID: z.string().cuid()
});

export type SayHelloInput = z.infer<typeof SayHelloInputSchema>;

export async function sayHello(inputs: SayHelloInput) {
  const { deviceID } = SayHelloInputSchema.parse(inputs);
  const channel = await publisher;
  await channel.assertQueue(deviceID + ".hello", { durable: false });
  channel.sendToQueue(deviceID + ".hello", Buffer.from("Hello World!"));
  console.log(" [x] Sent 'Hello World!'");
  return true;
}

export const SendCommandInputSchema = z.object({
  deviceID: z.string().cuid(),
  command: z.string(),
  data: z.string()
});

export type SendCommandInput = z.infer<typeof SendCommandInputSchema>;

export async function sendCommand(input: SendCommandInput) {
  const { deviceID, command, data } = SendCommandInputSchema.parse(input);
  const channel = await publisher;

  await channel.assertExchange(deviceID + ".commands", "topic", {
    durable: false
  });
  channel.publish(deviceID + ".commands", command, Buffer.from(data));

  console.log(` [x] Sent '${data}' to the ${command} channel at ${deviceID}`);
  return true;
}
