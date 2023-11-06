import type { User } from "@prisma/client";

export type UserDetails = Pick<User, "id" | "email" | "role">