import type { Arc, Transition, Place, Event } from "@prisma/client";
import { prisma } from "~/db.server";
import { z } from "zod";

export const TransitionInputSchema = z.object({
  netID: z.string().cuid(),
  name: z.string(),
  description: z.string().optional(),
  condition: z.string().optional()
});

export type TransitionInput = z.infer<typeof TransitionInputSchema>;

export type TransitionDetails =
  Pick<Transition, "id" | "name" | "description" | "condition" | "createdAt" | "updatedAt">
  & {
  arcs: {
    id: Arc["id"];
    fromPlace: boolean;
    place: {
      id: Place["id"]
      name: Place["name"]
    },
  }[],
  events: {
    id: Event["id"];
    name: Event["name"];
    description?: Event["description"];
    fields: {
      id: string;
      name: string;
      type: string;
    }[]
  }[]
}

export async function getTransition({ id }: Pick<Transition, "id">): Promise<TransitionDetails> {
  return prisma.transition.findFirst({
    where: { id },
    select: {
      id: true,
      createdAt: true,
      updatedAt: true,
      name: true,
      description: true,
      condition: true,
      arcs: {
        select: {
          id: true,
          fromPlace: true,
          place: {
            select: {
              id: true,
              name: true
            }
          }
        }
      },
      events: {
        select: {
          id: true,
          name: true,
          description: true,
          fields: {
            select: {
              id: true,
              name: true,
              type: true
            }
          }
        }
      }
    }
  }).then((transition) => {
    if (!transition) {
      throw new Error("transition not found");
    }
    return transition;
  });
}

export type TransitionListItem = Pick<Transition, "id" | "name"> & {
  hasEvent: boolean
}

export async function listTransitions({ netID }: { netID: string }): Promise<TransitionListItem[]> {
  return prisma.transition.findMany({
    where: {
      nets: {
        some: {
          id: netID
        }
      }
    },
    select: {
      id: true,
      name: true,
      events: {
        select: {
          id: true
        }
      }
    }
  }).then((transitions) => {
    return transitions.map((transition) => {
      return {
        id: transition.id,
        name: transition.name,
        hasEvent: transition.events.length > 0
      };
    });
  });
}

export async function addTransition(input: TransitionInput) {
  const { netID, name, description, condition } = TransitionInputSchema.parse(input);
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

export const UpdateTransitionSchema = z.object({
  id: z.string().cuid(),
  name: z.string(),
  description: z.string().optional(),
  condition: z.string().optional()
});

export type UpdateTransitionInput = z.infer<typeof UpdateTransitionSchema>

export async function updateTransition(input: UpdateTransitionInput) {
  const { id, name, description, condition } = UpdateTransitionSchema.parse(input);
  return prisma.transition.updateMany({
    where: { id },
    data: { name, description, condition }
  });
}

export async function deleteTransition({ id }: Pick<Transition, "id">) {
  return prisma.transition.deleteMany({
    where: {
      id: id
    }
  });
}