import { PrismaClient } from "@prisma/client";
import bcrypt from "bcryptjs";

const prisma = new PrismaClient();

async function seed() {
  const email = "jonathan@petri.local";

  // cleanup the existing database
  await prisma.user.deleteMany({}).catch(() => {
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
    },
    include: {
      places: true,
      transitions: true
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
      }
    }
  });

  async function join({ netID, placeName, transitionName, fromPlace, places, transitions }: {
    netID: string,
    fromPlace: boolean,
    placeName: string,
    transitionName: string,
    places: (typeof valve.places)
    transitions: (typeof valve.transitions)
  }) {
    const placeID = places.find(place => place.name === placeName)?.id;
    const transitionID = transitions.find(transition => transition.name === transitionName)?.id;
    if (placeID && transitionID) {
      return await prisma.arc.create({
        data: {
          netID,
          placeID,
          transitionID,
          fromPlace
        }
      });
    }
  }

  const arcs: {
    fromPlace: boolean,
    placeName: string,
    transitionName: string
  }[] = [
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

  for (let arc of arcs) {
    await join({
      netID: valve.id,
      fromPlace: arc.fromPlace,
      placeName: arc.placeName,
      transitionName: arc.transitionName,
      places: valve.places,
      transitions: valve.transitions
    });
  }

  const pump = await prisma.net.create({
    data: {
      name: "syringe pump",
      description: "A syringe pump.",
      authorID: user.id,
      parentID: mainNet.id,
      initialMarking: [1, 0, 0, 0],
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
            name: "Pumping",
            description: "The pump is dispensing.",
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
            name: "Pumped",
            description: "Fluid was pumped.",
            events: {
              create: [{
                name: "Pump",
                description: "Pump fluid.",
                fields: {
                  create: [{
                    name: "Volume",
                    type: "number"
                  },
                    {
                      name: "Rate",
                      type: "number"
                    }]
                }
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

  const pumpArcs: {
    fromPlace: boolean,
    placeName: string,
    transitionName: string
  }[] = [
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
      transitionName: "Pumped"
    },
    {
      fromPlace: false,
      placeName: "Pumping",
      transitionName: "Pumped"
    },
    {
      fromPlace: true,
      placeName: "Pumping",
      transitionName: "Stopped"
    },
    {
      fromPlace: false,
      placeName: "Idle",
      transitionName: "Stopped"
    },
    {
      fromPlace: false,
      placeName: "Flow syringe",
      transitionName: "Stopped"
    }
  ];

  for (let arc of pumpArcs) {
    await join({
      netID: pump.id,
      places: pump.places,
      transitions: pump.transitions,
      ...arc
    });
  }


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
            description: "Fluid flowed through the syringe."
          }
        ]
      }
    },
    include: {
      places: true,
      transitions: true
    }
  });

  const syringeArcs: {
    fromPlace: boolean,
    placeName: string,
    transitionName: string
  }[] = [
    {
      fromPlace: true,
      placeName: "Solution",
      transitionName: "Flowed"
    },
    {
      fromPlace: false,
      placeName: "Flow",
      transitionName: "Flowed"
    }
  ];
  for (let arc of syringeArcs) {
    await join({
      netID: syringe.id,
      places: syringe.places,
      transitions: syringe.transitions,
      ...arc
    });
  }

  await prisma.net.update({
    where: {
      id: mainNet.id
    },
    data: {
      placeInterfaces: {
        create: [
          {
            name: "Flow through syringe",
            bound: 1,
            places: {
              connect: [
                {
                  id: syringe.places.find(place => place.name === "Solution")!.id
                },
                {
                  id: pump.places.find(place => place.name === "Flow syringe")!.id
                }
              ]
            }
          },
          {
            name: "Flow through main",
            bound: 1,
            places: {
              connect: [
                {
                  id: syringe.places.find(place => place.name === "Flow")!.id
                },
                {
                  id: valve.places.find(place => place.name === "Flow main")!.id
                }
              ]
            }
          },
          {
            name: "Flow through inlet",
            bound: 1,
            places: {
              connect: [
                {
                  id: valve.places.find(place => place.name === "Flow A")!.id
                },
                {
                  id: mainNet.places.find(place => place.name === "source")!.id
                }
              ]
            }
          },
          {
            name: "Flow through outlet",
            bound: 1,
            places: {
              connect: [
                {
                  id: valve.places.find(place => place.name === "Flow B")!.id
                },
                {
                  id: mainNet.places.find(place => place.name === "destination")!.id
                }
              ]
            }
          }
        ]
      }
    }
  });
  await prisma.device.create({
    data: {
      name: "two position three way valve",
      description: "A two position three way valve.",
      authorID: user.id,
      nets: {
        connect: [
          {
            id: valve.id
          }
        ]
      },
      instances: {
        create: [
          {
            authorID: user.id,
            name: "valve",
            language: "GO",
            addr: "http://localhost:8080/valve"
          }
        ]
      }
    }
  });
  await prisma.device.create({
    data: {
      name: "syringe pump",
      description: "A syringe pump.",
      authorID: user.id,
      nets: {
        connect: [
          {
            id: pump.id
          }
        ]
      },
      instances: {
        create: [
          {
            authorID: user.id,
            name: "pump",
            language: "GO",
            addr: "http://localhost:8080/pump"
          }
        ]
      }
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