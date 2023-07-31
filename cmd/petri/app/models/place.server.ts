import type { Net, Place } from "@prisma/client";
import { prisma } from "~/db.server";
import type { ArcDetails} from "~/models/arc.server";
import { getPlaceIO } from "~/models/arc.server";


export async function getPlace({ id }: Pick<Place, "id">) :Promise<UpdatePlaceInput | null> {
  const io = await getPlaceIO({ id });
  const res = await prisma.place.findFirst({
    select: {
      id: true,
      name: true,
      description: true,
      bound: true
    },
    where: { id }
  });

  if (!res) {
    return null;
  }

  return { ...res, inputs: io.inputs, outputs: io.outputs };
}


export type PlaceInput = Pick<Place, "name" | "bound" | "description">

export type UpdatePlaceInput = Pick<Place, "id" | "name" | "bound" | "description"> & {
  inputs: ArcDetails[]
  outputs: ArcDetails[]
}

export function addPlace({
                           name,
                           bound,
                           description,
                           netID
                         }: PlaceInput & { netID: Net["id"] }) {
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

export async function updatePlace({ id, name, bound, description, inputs, outputs }: UpdatePlaceInput) {

  // Update Arcs accordingly
  for (let input of inputs) {
    const arc = await prisma.arc.findUnique({
      where: { id: input.id }
    });
    if (!arc) {
      throw new Error("Arc not found");
    }
    if (arc.transitionID !== input.id || arc.fromPlace !== false) {
      await prisma.arc.update({
        where: { id: input.id },
        data: {
          transitionID: input.id,
          fromPlace: false
        }
      });
    }
  }

  for (let output of outputs) {
    const arc = await prisma.arc.findUnique({
      where: { id: output.id }
    });
    if (!arc) {
      throw new Error("Arc not found");
    }
    if (arc.transitionID !== output.id || arc.fromPlace !== true) {
      await prisma.arc.update({
        where: { id: output.id },
        data: {
          transitionID: output.id,
          fromPlace: true
        }
      });
    }
  }

  // Update the place
  return prisma.place.updateMany({
    where: { id },
    data: { name, bound, description }
  });
}
export function deletePlace({ id }: Pick<Place, "id">) {
  return prisma.place.deleteMany({
    where: { id }
  });
}
export type DeletePlaceInput = Pick<Place, "id">
