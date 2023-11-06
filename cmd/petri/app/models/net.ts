import { z } from "zod";
import { Visibility } from "@prisma/client";

export const NetInputSchema = z.object({
  name: z.string(),
  authorID: z.string().cuid(),
  description: z.string()
});

export type NetInput = z.infer<typeof NetInputSchema>;

export const UpdateNetSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(1).optional(),
  visibility: z.nativeEnum(Visibility).optional(),
  sharedWith: z.array(z.string().email()).optional()
});

export type UpdateNet = z.infer<typeof UpdateNetSchema>;
