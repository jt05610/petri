import { Net, Transition } from "@prisma/client";
import { prisma } from "~/db.server";
import type { ArcDetails } from "~/models/arc.server";
import { getTransIO } from "~/models/arc.server";

export type UpdateTransitionInput = Pick<Transition, "id" | "name" | "description" | "condition"> & {
  inputs: ArcDetails[]
  outputs: ArcDetails[]
}

export async function getTransition({ id }: Pick<Transition, "id">): Promise<UpdateTransitionInput | null> {
  const io = await getTransIO({ id });
  const res = await prisma.transition.findFirst({
    select: {
      id: true,
      name: true,
      description: true,
      condition: true
    },
    where: { id }
  });
  if (!res) {
    return null;
  }
  let inputs: ArcDetails[] = [];
  let outputs: ArcDetails[] = [];
  io.forEach((arc) => {
    if (arc.fromPlace) {
      inputs.push({ id: arc.place.id, name: arc.place.name, arcID: arc.id });
    } else {
      outputs.push({ id: arc.place.id, name: arc.place.name, arcID: arc.id });
    }
  });
  return { ...res, inputs: inputs, outputs: outputs };
}


export type TransitionInput = Pick<Transition, "name" | "description" | "condition">

export function addTransition({ name, description, condition, netID }: TransitionInput & { netID: Net["id"] }) {
  return prisma.transition.create({
    data: {
      name: name,
      description: description,
      condition: condition,
      nets: {
        connect: {
          id: netID
        }
      }
    }
  });
}

export function updateTransition({ id, name, description, condition }: UpdateTransitionInput) {
  return prisma.transition.updateMany({
    where: { id },
    data: { name, description, condition }
  });
}

export type DeleteTransitionInput = Pick<Transition, "id">
export function deleteTransition({ id }: Pick<Transition, "id">) {
  return prisma.transition.deleteMany({
    where: {
      id: id
    }
  });
}