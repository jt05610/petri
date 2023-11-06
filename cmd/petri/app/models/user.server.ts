import type { User } from "@prisma/client";
import { z } from "zod";
import bcrypt from "bcryptjs";

import { prisma } from "~/db.server";
import type { UserDetails } from "~/models/user";

export type { User } from "@prisma/client";

export async function getUserById(id: User["id"]): Promise<UserDetails | null> {
  return prisma.user.findUnique(
    {
      where: { id },
      select: {
        id: true,
        email: true,
        role: true
      }
    }
  );
}

export async function getUserByEmail(email: User["email"]) {
  return prisma.user.findUnique({ where: { email } });
}

export const CreateUserSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8)
});

export type CreateUserInput = z.infer<typeof CreateUserSchema>;

export async function createUser({ email, password }: CreateUserInput) {
  const hashedPassword = await bcrypt.hash(password, 10);
  return prisma.user.create({
    data: {
      email,
      role: "USER",
      password: {
        create: {
          hash: hashedPassword
        }
      }
    }
  });
}

export const VerifyLoginSchema = z.object({
  email: z.string().email(),
  password: z.string(),
  redirectTo: z.string().optional(),
  remember: z.preprocess((value) => Boolean(value), z.boolean()).optional().default(false)
});

export type VerifyLoginInput = z.infer<typeof VerifyLoginSchema>;

export async function verifyLogin({ email, password }: VerifyLoginInput) {
  const userWithPassword = await prisma.user.findUnique({
    where: { email },
    include: {
      password: true
    }
  });

  if (!userWithPassword || !userWithPassword.password) {
    return null;
  }

  const isValid = await bcrypt.compare(
    password,
    userWithPassword.password.hash
  );

  if (!isValid) {
    return null;
  }

  const { password: _password, ...userWithoutPassword } = userWithPassword;

  return userWithoutPassword;
}
