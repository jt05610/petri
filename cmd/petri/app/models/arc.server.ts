import type { Arc, Place, Transition } from "@prisma/client";
import { prisma } from "~/db.server";
import { z } from "zod";

export const ArcInputFormSchema = z.object({
  netID: z.string().cuid(),
  placeID: z.string().cuid(),
  transitionID: z.string().cuid(),
  fromPlace: z.preprocess((val) => {
    return val === "true" || val === "on";
  }, z.boolean())
});
export const ArcInputSchema = z.object({
  netID: z.string().cuid(),
  placeID: z.string().cuid(),
  transitionID: z.string().cuid(),
  fromPlace: z.boolean()
});

export type ArcInput = z.infer<typeof ArcInputSchema>;

export async function addArc(input: ArcInput) {
  const { netID, placeID, transitionID, fromPlace } = input;
  return prisma.arc.create({
    data: {
      place: {
        connect: {
          id: placeID
        }
      },
      transition: {
        connect: {
          id: transitionID
        }
      },
      net: {
        connect: {
          id: netID
        }
      },
      fromPlace: fromPlace
    }
  });
}

export const ArcUpdateSchema = z.object({
  id: z.string().cuid().optional(),
  fromPlace: z.boolean().optional(),
  placeID: z.string().cuid().optional(),
  transitionID: z.string().cuid().optional()
});

export const ArcUpdateFormSchema = z.object({
  id: z.string().cuid().optional(),
  fromPlace: z.preprocess((val) => {
    return val === "true" || val === "on";
  }, z.boolean()),
  placeID: z.string().cuid().optional(),
  transitionID: z.string().cuid().optional()
});
export type ArcUpdate = z.infer<typeof ArcUpdateSchema>;

export async function updateArc(input: ArcUpdate) {
  const { id, fromPlace, placeID, transitionID } = ArcUpdateSchema.parse(input);
  return prisma.arc.update({
    where: { id },
    data: {
      fromPlace: fromPlace,
      place: {
        connect: {
          id: placeID
        }
      },
      transition: {
        connect: {
          id: transitionID
        }
      }
    }
  });
}

export type ArcDetails = Pick<Arc, "id" | "fromPlace" | "createdAt" | "updatedAt"> & {
  place: Pick<Place, "id" | "name">,
  transition: Pick<Transition, "id" | "name">
}

export const GetArcSchema = z.object({
  id: z.string().cuid()
});

export type GetArcInput = z.infer<typeof GetArcSchema>

export async function getArc(input: GetArcInput): Promise<ArcDetails> {
  const { id } = GetArcSchema.parse(input);
  return prisma.arc.findFirst({
    select: {
      id: true,
      fromPlace: true,
      createdAt: true,
      updatedAt: true,
      place: {
        select: {
          id: true,
          name: true
        }
      },
      transition: {
        select: {
          id: true,
          name: true
        }
      }
    },
    where: { id }
  }).then((arc) => {
    if (!arc) {
      throw new Error("arc not found");
    }
    return arc;
  });
}

export type ArcListItem = Pick<Arc, "id" | "fromPlace"> & {
  place: Pick<Place, "id" | "name">,
  transition: Pick<Transition, "id" | "name">
}

export const ListArcsSchema = z.object({
  netID: z.string().cuid()
});

export type ListArcsInput = z.infer<typeof ListArcsSchema>

export async function listArcs(input: ListArcsInput): Promise<ArcListItem[]> {
  const { netID } = ListArcsSchema.parse(input);
  return prisma.arc.findMany({
    select: {
      id: true,
      fromPlace: true,
      place: {
        select: {
          id: true,
          name: true
        }
      },
      transition: {
        select: {
          id: true,
          name: true
        }
      }
    },
    where: { netID }
  });
}

export type DeleteArcInput = Pick<Arc, "id">

export function deleteArc({ id }: DeleteArcInput) {
  return prisma.arc.delete({
    where: { id }
  });
}
