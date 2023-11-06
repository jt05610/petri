import type { Transition, Event, Field } from "@prisma/client";
import { z } from "zod";

export const EventFieldSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  type: z.string(),
  min: z.string().optional(),
  max: z.string().optional(),
  choices: z.preprocess((val) => {
      if (!Array.isArray(val)) {
        if (typeof val === "string") {
          return val.split(",");
        }
      }
      return val;
    },
    z.array(z.string())
  ),
  unit: z.string().optional()
});

export const EventInputSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  fields: z.preprocess((val) => {
    if (!Array.isArray(val)) {
      return [val];
    }
    return val;
  }, z.array(EventFieldSchema))
});

export type EventInput = z.infer<typeof EventInputSchema>;

export type EventField = z.infer<typeof EventFieldSchema>;

export const UpdateEventSchema = z.object({
  id: z.string().cuid().optional(),
  name: z.string().optional(),
  description: z.string().optional(),
  fields: z.array(EventFieldSchema)
});

export type EventUpdate = z.infer<typeof UpdateEventSchema>;

export const GetEventSchema = z.object({
  id: z.string().cuid()
});

export type GetEventInput = z.infer<typeof GetEventSchema>;

export type EventDetails = Pick<Event, "id" | "name" | "description" | "createdAt" | "updatedAt"> & {
  transitions: Pick<Transition, "id" | "name">[],
  fields: Pick<Field, "id" | "name" | "type">[]
}


export const ListEventsSchema = z.object({
  transitionID: z.string().cuid()
});

export type ListEventsInput = z.infer<typeof ListEventsSchema>;

export type EventListItem = Pick<Event, "id" | "name" | "description">


export const DeleteEventSchema = z.object({
  id: z.string().cuid()
});

export type DeleteEventInput = z.infer<typeof DeleteEventSchema>;
