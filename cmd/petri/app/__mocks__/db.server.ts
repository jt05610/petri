import type { PrismaClient } from "@prisma/client";
import { beforeEach } from "vitest";
import { mockDeep, mockReset } from "vitest-mock-extended";
import { singleton } from "~/singleton.server";

// Hard-code a unique key, so we can look up the client when this module gets re-imported
const prisma = singleton("prisma", () => mockDeep<PrismaClient>());

beforeEach(() => {
  mockReset(prisma);
});

export { prisma };