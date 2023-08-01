import { PrismaClient } from "@prisma/client";
import bcrypt from "bcryptjs";

const prisma = new PrismaClient();

async function seed() {
  const email = "jonathan@petri.local";

  // cleanup the existing database
  await prisma.user.delete({
    where: {
      email
    }
  }).catch(() => {
    // no worries if it doesn't exist yet
  });

  const hashedPassword = await bcrypt.hash("jonathaniscool", 10);
  const user = await prisma.user.create({
    data: {
      email,
      role: "ADMIN",
      password: {
        create: {
          hash: hashedPassword
        }
      }
    }
  });

  const mainNet = await prisma.net.create({
    data: {
      name: "transfer syringe pump",
      description: `A syringe pump connected to a two-position three-way valve. 
      The three-way valve is connected to two separate containers, and fluid can 
      be transferred between the containers by switching the valve to the source 
      container, reloading the pump with the desired volume, switching the valve 
      to the destination container, and then dispensing the volume. The pump can 
      be reloaded and the process repeated.`,
      authorID: user.id,
      places: {
        create: [
          {
            name: "source",
            description: "the syringe is connected to the source container",
            bound: 1
          },
          {
            name: "destination",
            description: "the syringe is connected to the destination container",
            bound: 1
          }
        ]
      }
    }
  });

  const valve = await prisma.net.create({
    data: {
      name: "two-position three-way valve",
      description: "A valve with two ",
      authorID: user.id,
      parentID: mainNet.id,
      initialMarking: [1, 0, 0, 0, 0],
      places: {
        create: [
          {
            name: "Position A",
            description: "The valve is in position A.",
            bound: 1
          },
          {
            name: "Position B",
            description: "The valve is in position B.",
            bound: 1
          },
          {
            name: "Flow main",
            description: "Fluid is flowing through the main port.",
            bound: 1
          },
          {
            name: "Flow A",
            description: "Fluid is flowing through port A.",
            bound: 1
          },
          {
            name: "Flow B",
            description: "Fluid is flowing through port B.",
            bound: 1
          }
        ]
      },
      transitions: {
        create: [
          {
            name: "A opened",
            description: "Position A was opened.",
            events: {
              create: [{
                name: "Open A",
                description: "Open port A."
              }]
            }
          },
          {
            name: "B opened",
            description: "Position B was opened.",
            events: {
              create: [{
                name: "Open B",
                description: "Open port B."
              }]
            }
          },
          {
            name: "Flowed through A",
            description: "Fluid flowed through port A."
          },
          {
            name: "Flowed through B",
            description: "Fluid flowed through port B."
          }
        ]
      }
    },
    include: {
      places: true,
      transitions: true
    }
  });

  function join({ netID, placeName, transitionName, fromPlace }: {
    netID: string,
    fromPlace: boolean,
    placeName: string,
    transitionName: string
  }) {
    const placeID = valve.places.find(place => place.name === placeName)?.id;
    const transitionID = valve.transitions.find(transition => transition.name === transitionName)?.id;
    if (placeID && transitionID) {
      prisma.arc.create({
        data: {
          netID,
          placeID,
          transitionID,
          fromPlace
        }
      });
    }
  }

  const arcs: { fromPlace: boolean, placeName: string, transitionName: string }[] = [
    {
      fromPlace: true,
      placeName: "Position A",
      transitionName: "B opened"
    },
    {
      fromPlace: false,
      placeName: "Position A",
      transitionName: "A opened"
    },
    {
      fromPlace: true,
      placeName: "Position B",
      transitionName: "A opened"
    },
    {
      fromPlace: false,
      placeName: "Position B",
      transitionName: "B opened"
    },
    {
      fromPlace: true,
      placeName: "Flow main",
      transitionName: "Flowed through A"
    },
    {
      fromPlace: true,
      placeName: "Position A",
      transitionName: "Flowed through A"
    },
    {
      fromPlace: false,
      placeName: "Position A",
      transitionName: "Flowed through A"
    },
    {
      fromPlace: false,
      placeName: "Flow A",
      transitionName: "Flowed through A"
    },
    {
      fromPlace: true,
      placeName: "Flow main",
      transitionName: "Flowed through B"
    },
    {
      fromPlace: true,
      placeName: "Position B",
      transitionName: "Flowed through B"
    },
    {
      fromPlace: false,
      placeName: "Position B",
      transitionName: "Flowed through B"
    },
    {
      fromPlace: false,
      placeName: "Flow B",
      transitionName: "Flowed through B"
    }
  ];

  arcs.forEach(arc => join({ netID: valve.id, ...arc }));

  const pump = await prisma.net.create({
    data: {
      name: "syringe pump",
      description: "A syringe pump.",
      authorID: user.id,
      parentID: mainNet.id,
      places: {
        create: [
          {
            name: "Unknown",
            description: "The pump is on but the position is unknown.",
            bound: 1
          },
          {
            name: "Idle",
            description: "The pump is idle.",
            bound: 1
          },
          {
            name: "Dispensing",
            description: "The pump is dispensing.",
            bound: 1
          },
          {
            name: "Reloading",
            description: "The pump is reloading.",
            bound: 1
          },
          {
            name: "Flow syringe",
            description: "Fluid is flowing through the syringe.",
            bound: 1
          }
        ]
      },
      transitions: {
        create: [
          {
            name: "Initialized",
            description: "The pump was initialized.",
            events: {
              create: [{
                name: "Initialize",
                description: "Initialize the pump."
              }]
            }
          },
          {
            name: "Dispensed",
            description: "Fluid was dispensed.",
            events: {
              create: [{
                name: "Dispense",
                description: "Dispense fluid."
              }]
            }
          },
          {
            name: "Reloaded",
            description: "The pump was reloaded.",
            events: {
              create: [{
                name: "Reload",
                description: "Reload the pump."
              }]
            }
          },
          {
            name: "Stopped",
            description: "The pump was stopped.",
            events: {
              create: [{
                name: "Stop",
                description: "Stop the pump."
              }]
            }
          }
        ]
      }
    },
    include: {
      places: true,
      transitions: true
    }
  });

  const pumpArcs: { fromPlace: boolean, placeName: string, transitionName: string }[] = [
    {
      fromPlace: true,
      placeName: "Unknown",
      transitionName: "Initialized"
    },
    {
      fromPlace: false,
      placeName: "Idle",
      transitionName: "Initialized"
    },
    {
      fromPlace: true,
      placeName: "Idle",
      transitionName: "Dispensed"
    },
    {
      fromPlace: false,
      placeName: "Dispensing",
      transitionName: "Dispensed"
    },
    {
      fromPlace: true,
      placeName: "Dispensing",
      transitionName: "Stopped"
    },
    {
      fromPlace: true,
      placeName: "Idle",
      transitionName: "Reloaded"
    },
    {
      fromPlace: false,
      placeName: "Reloading",
      transitionName: "Reloaded"
    },
    {
      fromPlace: true,
      placeName: "Reloading",
      transitionName: "Stopped"
    },
    {
      fromPlace: false,
      placeName: "Idle",
      transitionName: "Stopped"
    },
    {
      fromPlace: true,
      placeName: "Flow syringe",
      transitionName: "Stopped"
    }
  ];

  pumpArcs.forEach(arc => join({ netID: pump.id, ...arc }));

  let syringes = [];
  for (let i = 0; i < 3; i++) {

    const syringe = await prisma.net.create({
      data: {
        name: "syringe",
        description: "A syringe.",
        authorID: user.id,
        parentID: mainNet.id,
        places: {
          create: [
            {
              name: "Solution",
              description: "The syringe contains solution.",
              bound: 1
            },
            {
              name: "Flow",
              description: "Fluid is flowing through the syringe.",
              bound: 1
            }
          ]
        },
        transitions: {
          create: [
            {
              name: "Flowed",
              description: "Fluid flowed through the syringe.",
              events: {
                create: [{
                  name: "Flow",
                  description: "Flow fluid through the syringe."
                }]
              }
            }
          ]
        }
      },
      include: {
        places: true,
        transitions: true
      }
    });

    const syringeArcs: { fromPlace: boolean, placeName: string, transitionName: string }[] = [
      {
        fromPlace: false,
        placeName: "Solution",
        transitionName: "Flowed"
      },
      {
        fromPlace: true,
        placeName: "Flow",
        transitionName: "Flowed"
      }
    ];

    syringeArcs.forEach(arc => join({ netID: syringe.id, ...arc }));
    syringes.push(syringe);
  }

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