import type { Transition, Event, Field } from "@prisma/client";
import { prisma } from "~/db.server";
import { z } from "zod";

export const EventFieldSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  type: z.string()
});
export const EventInputSchema = z.object({
  transitionID: z.string().uuid(),
  name: z.string(),
  description: z.string().optional(),
  fields: z.array(EventFieldSchema)
});

export type EventInput = z.infer<typeof EventInputSchema>;

export async function addEvent(input: EventInput) {
  const { transitionID, name, description, fields } = EventInputSchema.parse(input);
  return prisma.event.create({
    data: {
      transitions: {
        connect: {
          id: transitionID
        }
      },
      name,
      description,
      fields: {
        create: fields
      }
    }
  });
}

export const EventUpdateSchema = z.object({
  id: z.string().uuid().optional(),
  name: z.string().optional()
});

export type EventUpdate = z.infer<typeof EventUpdateSchema>;

export async function updateEvent(input: EventUpdate) {
  const { id, name } = EventUpdateSchema.parse(input);
  return prisma.event.update({
    where: { id },
    data: {
      name
    }
  });
}

export const GetEventSchema = z.object({
  id: z.string().uuid()
});

export type GetEventInput = z.infer<typeof GetEventSchema>;

export type EventDetails = Pick<Event, "id" | "name" | "description" | "createdAt" | "updatedAt"> & {
  transitions: Pick<Transition, "id" | "name">[],
  fields: Pick<Field, "id" | "name" | "type">[]
}

export async function getEvent(input: GetEventInput): Promise<EventDetails> {
  const { id } = GetEventSchema.parse(input);
  return prisma.event.findFirst({
    select: {
      id: true,
      name: true,
      description: true,
      createdAt: true,
      updatedAt: true,
      transitions: {
        select: {
          id: true,
          name: true
        }
      },
      fields: {
        select: {
          id: true,
          name: true,
          type: true
        }
      }
    },
    where: { id }
  }).then((event) => {
    if (!event) throw new Error("Event not found");
    return event;
  });
}

export const ListEventsSchema = z.object({
  transitionID: z.string().uuid()
});

export type ListEventsInput = z.infer<typeof ListEventsSchema>;

export type EventListItem = Pick<Event, "id" | "name" | "description">

export async function listEvents(input: ListEventsInput): Promise<EventListItem[]> {
  const { transitionID } = ListEventsSchema.parse(input);
  return prisma.event.findMany({
    select: {
      id: true,
      name: true,
      description: true
    },
    where: {
      transitions: {
        some: {
          id: transitionID
        }
      }
    }
  });
}

export const DeleteEventSchema = z.object({
  id: z.string().uuid()
});

export type DeleteEventInput = z.infer<typeof DeleteEventSchema>;

export async function deleteEvent(input: DeleteEventInput) {
  const { id } = DeleteEventSchema.parse(input);
  return prisma.event.delete({
    where: { id }
  });
}