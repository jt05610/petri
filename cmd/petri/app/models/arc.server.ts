import type { Arc, Place, Transition } from "@prisma/client";
import { prisma } from "~/db.server";
import { Net } from "@prisma/client";

export type ArcDetails = Pick<Place | Transition, "id" | "name"> & {
  arcID: Arc["id"]
}

export function getPlaceIO({ id }: Pick<Place, "id">) {
  return prisma.arc.findMany({
    where: {
      placeID: id
    },
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
  });
}

export function getTransIO({ id }: Pick<Transition, "id">) {
  return prisma.arc.findMany({
    where: {
      transitionID: id
    },
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
  });
}


export type ArcInput = Pick<Arc, "placeID" | "transitionID" | "fromPlace">

export function addArc({ placeID, transitionID, netID, fromPlace }: ArcInput & { netID: Net["id"] }) {
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

export type DeleteArcInput = Pick<Arc, "id">

export function deleteArc({ id }: DeleteArcInput) {
  return prisma.arc.delete({
    where: { id }
  });
}
