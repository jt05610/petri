import { expect, vi } from "vitest";
import { createType, deleteType, getType, getTypes, updateType } from "~/models/type";
import type { TypeScope } from "@prisma/client";

// eslint-disable-next-line jest/no-mocks-import
import { prisma } from "~/__mocks__/db.server";

vi.mock("~/db.server");

let SEEN = 0;


async function createFakeType(name: string, scope: TypeScope) {
  prisma.type.create.mockResolvedValue({
    id: `test${SEEN++}`,
    name: name,
    description: "test",
    scope: scope,
    createdAt: new Date(),
    updatedAt: new Date(),
    userId: "testUser",
    deviceId: null,
    deletable: true
  });

  return await createType({
    name: name,
    description: "test",
    scope: scope,
    userId: "testUser"
  });
}

test("should create a type", async () => {
  const type = await createFakeType("test", "GLOBAL");
  expect(type.name).toBe("test");
  expect(type.description).toBe("test");
  expect(type.scope).toBe("GLOBAL");
});

test("should get a type", async () => {
  prisma.type.findUnique.mockResolvedValue({
    id: "test",
    name: "test",
    description: "test",
    scope: "GLOBAL",
    createdAt: new Date(),
    updatedAt: new Date(),
    userId: "testUser",
    deviceId: null,
    deletable: true
  });
  const type = await getType("test");
  if (!type) throw new Error("Type not found.");
  expect(type.name).toBe("test");
  expect(type.description).toBe("test");
  expect(type.scope).toBe("GLOBAL");
  expect(type.deletable).toBe(true);
})
;
test("should update a type", async () => {
  prisma.type.update.mockResolvedValue({
    id: "test",
    name: "test2",
    description: "test2",
    scope: "USER",
    createdAt: new Date(),
    updatedAt: new Date(),
    userId: "testUser",
    deviceId: null,
    deletable: true
  });
  const type = await updateType({
    typeID: "test",
    data: {
      name: "test2",
      description: "test2",
      scope: "USER"
    }
  });
  expect(type.name).toBe("test2");
});
test("should delete a type", async () => {
  prisma.type.delete.mockResolvedValue({
    id: "test",
    name: "test2",
    description: "test2",
    scope: "USER",
    createdAt: new Date(),
    updatedAt: new Date(),
    userId: "testUser",
    deviceId: null,
    deletable: true
  });

  const type = await deleteType({
    typeID: "test",
    authorID: "test"
  });
  expect(type.name).toBe("test2");

  const type2 = await getType("test");
  expect(type2).toBe(undefined);
})
;
test("should get types", async () => {
  const expects = [];

  for (let i = 0; i < 10; i++) {
    expects.push(await createFakeType(`test${i}`, "GLOBAL"));
  }
  prisma.type.findMany.mockResolvedValue(expects);
  const types = await getTypes({
    scope: ["GLOBAL"]
  });
  expect(types.length).toBe(10);
});