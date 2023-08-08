import { z } from "zod";
import type { Datum, Event, Instance, Prisma } from "@prisma/client";
import { prisma } from "~/db.server";
import { JSONSchema } from "~/models/net.run.server";

type InputJsonValue = Prisma.InputJsonValue;

export const DatumInputSchema = z.object({
  instanceID: z.string().cuid(),
  value: JSONSchema,
  eventID: z.string().cuid(),
  sessionID: z.string().cuid()
});

export type DatumInput = z.infer<typeof DatumInputSchema>;

export async function createDatum(input: DatumInput) {
  const { instanceID, value, eventID, sessionID } = DatumInputSchema.parse(input);
  return prisma.session.update({
    where: {
      id: sessionID
    },
    data: {
      data: {
        create: {
          instance: {
            connect: {
              id: instanceID
            }
          },
          value: value as InputJsonValue,
          event: {
            connect: {
              id: eventID
            }
          }
        }
      }
    }
  });
}

export const GetDatumInputSchema = z.object({
  id: z.string().cuid(),
  sessionID: z.string().cuid()
});

export type GetDatumInput = z.infer<typeof GetDatumInputSchema>;

export type DatumDetails = Pick<Datum, "id" | "value" | "createdAt" | "updatedAt"> & {
  instance: Pick<Instance, "id" | "name" | "addr">,
  event: Pick<Event, "id" | "name" | "description">
}

export async function getDatum(input: GetDatumInput): Promise<DatumDetails> {
  const { id, sessionID } = GetDatumInputSchema.parse(input);
  return prisma.session.findUniqueOrThrow({
    where: {
      id: sessionID
    },
    include: {
      data: {
        where: {
          id
        },
        include: {
          instance: true,
          event: true
        }
      }
    }
  }).then(session => {
    if (session.data.length === 0) {
      throw new Error("Datum not found");
    }
    return session.data[0];
  });
}

export const ListDataInputSchema = z.object({
  sessionID: z.string().cuid()
});

export type ListDataInput = z.infer<typeof ListDataInputSchema>;

export type DataListItem = Pick<Datum, "id" | "createdAt" | "updatedAt" | "value" | "instanceID">

export async function listData(input: ListDataInput): Promise<DataListItem[]> {
  const { sessionID } = ListDataInputSchema.parse(input);
  return prisma.session.findUniqueOrThrow({
    where: {
      id: sessionID
    },
    include: {
      data: {
        include: {
          instance: true,
          event: true
        }
      }
    }
  }).then(session => {
    return session.data;
  });
}
