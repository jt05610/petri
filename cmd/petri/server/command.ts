import { z } from "zod";

export const CommandSchema = z.object({
  deviceID: z.string().cuid(),
  command: z.string(),
  data: z.any().optional(),
  input: z.any(),
  output: z.any()
})

export type Command = z.infer<typeof CommandSchema>;


export const EventSchema = z.object({
  deviceID: z.string().cuid(),
  event: z.string(),
  data: z.any().optional(),
})

export type EventReceived = z.infer<typeof EventSchema>;

