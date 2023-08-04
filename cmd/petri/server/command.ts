import { z } from "zod";

export const CommandSchema = z.object({
  deviceID: z.string().cuid(),
  command: z.string(),
  data: z.any().optional()
})

export type Command = z.infer<typeof CommandSchema>;

