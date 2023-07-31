import type { User, Net } from "@prisma/client";
import { prisma } from "~/db.server";

export function getNet({
                         id,
                         authorID
                       }: Pick<Net, "id"> & {
  authorID: User["id"];
}) {
  return prisma.net.findFirst({
    select: {
      id: true,
      type: true,
      authorID: true,
      parentID: true,
      places: {
        select: {
          id: true,
          name: true
        }
      },
      transitions: {
        select: {
          id: true,
          name: true
        }
      },
      arcs: true,
      name: true,
      createdAt: true,
      updatedAt: true,
      description: true
    },
    where: { id, authorID }
  }).then((net) => {
    if (!net) {
      throw new Error("net not found");
    }
    return net;
  });
}

export type NetListItem = {
  id: Net["id"];
  authorID: Net["authorID"];
  name: Net["name"];
  createdAt: string;
  updatedAt: string;
}

export function getNetListItems({ authorID }: { authorID: User["id"] }) {
  return prisma.net.findMany({
    where: { authorID },
    select: { id: true, authorID: true, name: true, createdAt: true, updatedAt: true },
    orderBy: { updatedAt: "desc" }
  });
}

type CreateNetInput = Pick<Net, "name" | "description"> & {
  authorID: User["id"];
}

export function createNet({ name, description, authorID }: CreateNetInput): Promise<Net> {
  return prisma.net.create({
    data: {
      name: name,
      description: description,
      author: {
        connect: {
          id: authorID
        }
      }
    }
  });
}


export type UpdateNetInput = Pick<Net, "id" | "name" | "description">

export function updateNet({
                            id,
                            name,
                            authorID
                          }: UpdateNetInput & { authorID: User["id"]; }) {
  return prisma.net.updateMany({
    where: { id, authorID },
    data: { name }
  });
}


export function deleteNet({ id, authorID }: Pick<Net, "id"> & { authorID: User["id"] }) {
  return prisma.net.deleteMany({
    where: { id, authorID }
  });
}
