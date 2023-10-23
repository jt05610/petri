import type { Type } from "@prisma/client";
import { TypeScope } from "@prisma/client";
import { prisma } from "~/db.server";

import { z } from "zod";

export const TypeInputSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  scope: z.nativeEnum(TypeScope),
  userId: z.string().optional()
});

export type TypeInput = z.infer<typeof TypeInputSchema>;

export const GLOBAL_TYPES: TypeInput[] = [
  {
    name: "number",
    description: "A number.",
    scope: "GLOBAL"
  },
  {
    name: "string",
    description: "A string.",
    scope: "GLOBAL"
  },
  {
    name: "boolean",
    description: "A boolean.",
    scope: "GLOBAL"
  },
  {
    name: "object",
    description: "An object.",
    scope: "GLOBAL"
  },
  {
    name: "array",
    description: "An array.",
    scope: "GLOBAL"
  },
  {
    name: "null",
    description: "A null value.",
    scope: "GLOBAL"
  },
  {
    name: "int",
    description: "An integer.",
    scope: "GLOBAL"
  }
];

export async function seedGlobalTypes() {
  const globals: Record<string, Type> = {};
  for (let typeInput of GLOBAL_TYPES) {
    globals[typeInput.name] = await prisma.type.create({
      data: {
        ...typeInput,
        deletable: false
      }
    });
  }
  return globals;
}

export async function createType(typeInput: TypeInput) {
  return await prisma.type.create({
    data: typeInput
  });
}

export async function getType(typeID: string) {
  return await prisma.type.findUnique({
    where: { id: typeID }
  });
}

export const GetTypesSchema = z.object({
  scope: z.array(z.nativeEnum(TypeScope)).optional().default(["GLOBAL", "USER"])
});

export type GetTypesInput = z.infer<typeof GetTypesSchema>;

export async function getTypes(getTypesInput: GetTypesInput) {
  return await prisma.type.findMany({
    where: {
      scope: {
        in: getTypesInput.scope
      }
    }
  });
}

export const UpdateTypeSchema = z.object({
  typeID: z.string(),
  data: z.object({
    name: z.string().optional(),
    description: z.string().optional(),
    scope: z.nativeEnum(TypeScope).optional()
  })
});

export type UpdateTypeInput = z.infer<typeof UpdateTypeSchema>;

export async function updateType(updateTypeInput: UpdateTypeInput) {
  return await prisma.type.update({
    where: { id: updateTypeInput.typeID },
    data: updateTypeInput.data
  });
}

export const DeleteTypeSchema = z.object({
  typeID: z.string(),
  authorID: z.string()
});

export type DeleteTypeInput = z.infer<typeof DeleteTypeSchema>;

export async function deleteType(deleteTypeInput: DeleteTypeInput) {
  return await prisma.type.delete({
    where: { id: deleteTypeInput.typeID }
  });
}