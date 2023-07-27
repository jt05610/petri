import { PrismaClient } from "@prisma/client";
import bcrypt from "bcryptjs";

const prisma = new PrismaClient();

async function seed() {
  const email = "rachel@remix.run";

  // cleanup the existing database
  await prisma.user.deleteMany({ where: { email } }).catch(() => {
    // no worries if it doesn't exist yet
  });

  const hashedPassword = await bcrypt.hash("racheliscool", 10);
  const user = await prisma.user.create({
    data: {
      email,
      password: {
        create: {
          hash: hashedPassword
        }
      }
    }
  });

  const net = await prisma.net.create({
    data: {
      name: "door",
      description: "door example",
      authorID: user.id
    }
  });

  const pOpened = await prisma.place.create({
    data: {
      name: "opened",
      description: "the door was opened",
      bound: 1,
      nets: {
        create: {
          netID: net.id
        }
      }
    }
  });

  const pClosed = await prisma.place.create({
    data: {
      name: "closed",
      description: "the door was closed",
      bound: 1,
      nets: {
        create: {
          netID: net.id
        }
      }
    }
  });

  const tOpen = await prisma.transition.create({
    data: {
      name: "open",
      description: "open the door",
      nets: {
        create: {
          netID: net.id
        }
      }
    }
  });

  const tClose = await prisma.transition.create({
    data: {
      name: "close",
      description: "close the door",
      nets: {
        create: {
          netID: net.id
        }
      }
    }
  });

  await prisma.arc.create({
    data: {
      placeID: pClosed.id,
      transitionID: tOpen.id,
      netID: net.id,
      fromPlace: true
    }
  });

  await prisma.arc.create({
    data: {
      placeID: pOpened.id,
      transitionID: tOpen.id,
      netID: net.id,
      fromPlace: false
    }
  });
  await prisma.arc.create({
    data: {
      placeID: pOpened.id,
      transitionID: tClose.id,
      netID: net.id,
      fromPlace: true
    }
  });
  await prisma.arc.create({
    data: {
      placeID: pClosed.id,
      transitionID: tClose.id,
      netID: net.id,
      fromPlace: false
    }
  });

  console.log(`Database has been seeded. 🌱`);
}

seed()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
