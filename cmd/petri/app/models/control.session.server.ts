import { z } from "zod";

export const NewSessionInputSchema = z.object({
  sequenceID: z.string(),
  userID: z.string(),
  parameters: z.record(z.any()).optional(),
  instances: z.array(z.object({
    deviceID: z.string(),
    instanceID: z.string()
  }))
});
