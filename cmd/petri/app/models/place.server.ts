import type { Net, Place } from "@prisma/client";
import { prisma } from "~/db.server";
import type { ArcDetails } from "~/models/arc.server";
import { getPlaceIO } from "~/models/arc.server";


export async function getPlace({ id }: Pick<Place, "id">): Promise<UpdatePlaceInput> {
  const arcs = await getPlaceIO({ id });
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
    return {} as UpdatePlaceInput;
  }
  let inputs: ArcDetails[] = [];
  let outputs: ArcDetails[] = [];
  arcs.forEach((arc) => {
    if (arc.fromPlace) {
      inputs.push({ id: arc.transition.id, name: arc.transition.name, arcID: arc.id });
    } else {
      outputs.push({ id: arc.transition.id, name: arc.transition.name, arcID: arc.id });
    }
  });
  return { ...res, inputs: inputs, outputs: outputs };
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
  const allArcs = await prisma.arc.findMany({
    where: { placeID: id }
  });
  const io = inputs.concat(outputs);

  if (allArcs.length > io.length) {
    const toDelete = allArcs.filter((arc) => {
      return !io.some((io) => io.arcID === arc.id);
    });
    await prisma.arc.deleteMany({
      where: {
        id: {
          in: toDelete.map((arc) => arc.id)
        }
      }
    });
  }

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
