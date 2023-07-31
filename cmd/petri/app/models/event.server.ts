import type {Transition, Event, Field } from "@prisma/client";
import { prisma } from "~/db.server";

type UpdateEventInput = Pick<Event, "id" | "name"> & {
  transitions: Pick<Transition, "id" | "name">[]
  fields: Pick<Field, "name" | "type">[]
}
export async function getEvent({ id }: Pick<Event, "id">): Promise<UpdateEventInput | null> {
  const res = await prisma.event.findFirst({
    select: {
      id: true,
      name: true,
      transitions: {
        select: {
          id: true,
          name: true
        }
      },
      fields: {
        select: {
          name: true,
          type: true
        }
      }
    },
    where: { id }
  });

  if (!res) {
    return null;
  }

  return { ...res };
}
