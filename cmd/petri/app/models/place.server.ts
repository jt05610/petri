import type { Arc, Transition, Place } from "@prisma/client";
import { prisma } from "~/db.server";
import { z } from "zod";
export const PlaceFormSchema = z.object({
  name: z.string(),
  bound: z.preprocess(
    (bound) => parseInt(z.string().parse(bound), 10),
    z.number().gt(0).positive()
  ),
  description: z.string().optional(),
  netID: z.string().uuid()
});
export const PlaceInputSchema = z.object({
  name: z.string(),
  bound: z.number().gt(0).positive(),
  description: z.string().optional(),
  netID: z.string().uuid()
});

export type PlaceInput = z.infer<typeof PlaceInputSchema>;


export async function getPlace({ id }: Pick<Place, "id">): Promise<PlaceDetails> {
  return prisma.place.findFirst({
    select: {
      id: true,
      createdAt: true,
      updatedAt: true,
      name: true,
      description: true,
      bound: true,
      arcs: {
        select: {
          id: true,
          fromPlace: true,
          transition: {
            select: {
              id: true,
              name: true
            }
          }
        }
      }
    },
    where: { id }
  }).then((place) => {
    if (!place) {
      throw new Error("place not found");
    }

    return place;
  });
}

export type PlaceListItem = Pick<Place, "id" | "name" | "bound">;

export async function listPlaces({ netID }: { netID: string }): Promise<PlaceListItem[]> {
  return prisma.place.findMany({
    where: {
      nets: {
        some: {
          id: netID
        }
      }
    },
    select: { id: true, name: true, bound: true }
  });
}

export type PlaceDetails = Pick<Place, "id" | "name" | "bound" | "description" | "createdAt" | "updatedAt"> & {
  arcs: {
    id: Arc["id"];
    fromPlace: boolean;
    transition: {
      id: Transition["id"]
      name: Transition["name"]
    },
  }[]
}

export async function addPlace(place: PlaceInput) {
  const { name, bound, description, netID } = PlaceInputSchema.parse(place);
  return prisma.place.create({
    data: {
      name: name,
      bound: bound,
      description: description,
      nets: {
        connect: {
          id: netID
        }
      }
    }
  });
}

export const UpdatePlaceFormSchema = z.object({
  name: z.string().optional(),
  bound: z.preprocess(
    (bound) => parseInt(z.string().parse(bound), 10),
    z.number().gt(0).positive()
  ),
  description: z.string().optional()
});

export const UpdatePlaceInputSchema = z.object({
  name: z.string().optional(),
  bound: z.number().gt(0).positive(),
  description: z.string().optional()
});

export type UpdatePlaceInput = z.infer<typeof UpdatePlaceInputSchema>;

export async function updatePlace(id: Place["id"], place: UpdatePlaceInput) {
  const { name, bound, description } = UpdatePlaceInputSchema.parse(place);
  return prisma.place.updateMany({
    where: { id },
    data: { name, bound, description }
  });
}

export async function deletePlace({ id }: Pick<Place, "id">) {
  return prisma.place.deleteMany({
    where: { id }
  });
}
