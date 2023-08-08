import { vi } from "vitest";
import { prisma } from "~/db.server";

vi.mock("~/db.server", () => {
  return prisma;
});
