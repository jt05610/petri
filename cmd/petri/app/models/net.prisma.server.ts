import type { Net } from "~/models/net";
import type { Arc } from "@prisma/client";
import { prisma } from "~/db.server";
import { builder } from "~/models/builder";
import { z } from "zod";
import { createId } from "@paralleldrive/cuid2";

export const NetInputSchema = z.object({
  name: z.string(),
  authorID: z.string().cuid(),
  description: z.string()
});

export type NetInput = z.infer<typeof NetInputSchema>;

export const NetUpdateSchema = z.object({
  id: z.string().cuid2().optional(),
  name: z.string().optional(),
  description: z.string().optional()
});

export type NetUpdate = z.infer<typeof NetUpdateSchema>;

class Service implements Net.Service<NetInput, NetUpdate> {

  convertArcFromPrisma(net: Net.PetriNet, arc: Arc): Net.Arc {
    const transition = net.transitions.find(t => t.id === arc.transitionID);
    if (!transition) {
      throw new Error("Transition not found");
    }
    const place = net.places.find(p => p.id === arc.placeID);
    if (!place) {
      throw new Error("Place not found");
    }
    if (arc.fromPlace) {
      return {
        ...arc,
        from: place,
        to: transition
      };
    }
    return {
      ...arc,
      from: transition,
      to: place
    };
  }

  async create(net: NetInput): Promise<Net.PetriNet> {
    const created = await prisma.net.create({
      data: {
        id: createId(),
        ...net
      },
      include: {
        places: true,
        children: true,
        arcs: true,
        transitions: {
          include: {
            events: {
              include: {
                fields: {
                  include: {
                    type: true
                  }
                }
              }
            }
          }
        }
      }
    });
    let converted: Net.PetriNet = {
      id: created.id,
      name: created.name,
      places: created.places.map(p => ({
        ...p,
        marking: p.initialMarking
      })),
      transitions: created.transitions.map(t => ({
        ...t,
        events: []
      })),
      arcs: [],
      children: []
    };
    created.arcs.forEach(arc => {
      converted = builder.addArc(converted, this.convertArcFromPrisma(converted, arc));
    });
    return converted;
  }

  delete(id: string): Promise<void> {
    return Promise.resolve(undefined);
  }

  flush(net: Net.PetriNet): Promise<void> {
    return Promise.resolve(undefined);
  }

  list(): Promise<Net.PetriNet[]> {
    return Promise.resolve([]);
  }

  load(id: string): Promise<Net.PetriNet> {
    return Promise.resolve(undefined);
  }

  update(id: string, net: NetUpdate): Promise<Net.PetriNet> {
    return Promise.resolve(undefined);
  }
}

export const netService = new Service();