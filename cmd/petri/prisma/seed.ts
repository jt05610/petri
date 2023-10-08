import { PrismaClient } from "@prisma/client";
import type { Net, Place, Transition } from "@prisma/client";
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

  type NodeParams = {
    name: string,
    description: string,
  }

  type ServiceNetParams = {
    name: string,
    description: string,
    language: "GO" | "PYTHON",
    parentID?: string,
    service: NodeParams
    transition: NodeParams
    output: NodeParams
    event: NodeParams
    emit: NodeParams
    initializeFields?: {
      name: string,
      type: string
    }[]

    eventFields?: {
      name: string,
      type: string
    }[]

    emitFields?: {
      name: string,
      type: string
    }[]
  }

  function syringe_pump(name: string): ServiceNetParams {
    return {
      name: name,
      description: "A syringe pump.",
      language: "GO",
      service: {
        name: "Pumping",
        description: "The syringe pump is pumping."
      },
      transition: {
        name: "Pumped",
        description: "The syringe pump pumped."
      },
      output: {
        name: "Flow through syringe",
        description: "The syringe pump pumped."
      },
      event: {
        name: "Pump",
        description: "Pump."
      },
      emit: {
        name: "Pumped",
        description: "Pumped."
      },
      initializeFields: [
        {
          name: "syringe_diameter",
          type: "number"
        },
        {
          name: "syringe_volume",
          type: "number"
        },
        {
          name: "steps_per_mm",
          type: "number"
        },
        {
          name: "rate",
          type: "number"
        }
      ],
      eventFields: [
        {
          name: "Volume",
          type: "number"
        },
        {
          name: "Rate",
          type: "number"
        }
      ],
      emitFields: [
        {
          name: "Volume",
          type: "number"
        },
        {
          name: "Rate",
          type: "number"
        }
      ]
    };
  }

  const services: Record<string, ServiceNetParams> = {
    "organic_pump": syringe_pump("organic pump"),
    "aqueous_pump": syringe_pump("aqueous pump"),
    "mixing_valve": {
      name: "mixing_valve",
      description: "A mixing valve.",
      language: "PYTHON",
      initializeFields: [
        {
          name: "Components",
          type: "string"
        }
      ],
      eventFields: [
        {
          name: "Proportions",
          type: "string"
        }
      ],
      service: {
        name: "Mixing",
        description: "Fluid is mixing."
      },
      transition: {
        name: "Mixed",
        description: "Fluid was mixed."
      },
      output: {
        name: "Mixed",
        description: "Fluid was mixed."

      },
      emitFields: [
        {
          name: "Proportions",
          type: "string"
        }
      ],
      event: {
        name: "Mix",
        description: "Mix fluid."
      },
      emit: {
        name: "Mixed",
        description: "Fluids were mixed."
      }
    },
    "fraction_collector": {
      name: "fraction_collector",
      description: "A fraction collector.",
      language: "GO",
      service: {
        name: "Collecting",
        description: "The fraction collector is collecting."
      },
      transition: {
        name: "Collected",
        description: "The fraction collector collected."
      },
      output: {
        name: "Collected",
        description: "The fraction collector collected."
      },
      event: {
        name: "Collect",
        description: "Collect."
      },
      eventFields: [
        {
          name: "Position",
          type: "string"
        },
        {
          name: "Grid",
          type: "string"
        }
      ],
      emit: {
        name: "Collected",
        description: "Collected."
      },
      emitFields: [
        {
          name: "Position",
          type: "string"
        },
        {
          name: "Grid",
          type: "string"
        }
      ]
    },
    "camera": {
      name: "camera",
      description: "A camera.",
      language: "PYTHON",
      service: {
        name: "Capturing",
        description: "The camera is capturing."
      },
      transition: {
        name: "Captured",
        description: "The camera captured."
      },
      output: {
        name: "Captured",
        description: "The camera captured."
      },
      event: {
        name: "Capture",
        description: "Capture."
      },
      eventFields: [
        {
          name: "Duration",
          type: "number"
        },
        {
          name: "Interval",
          type: "number"
        }
      ],
      emit: {
        name: "Images captured",
        description: "Images were captured."
      },
      emitFields: [
        {
          name: "url",
          type: "string"
        }
      ]
    }
  };

  const service = async (
    {
      name,
      description,
      parentID,
      service,
      transition,
      output,
      initializeFields,
      event,
      eventFields,
      emitFields,
      emit
    }: ServiceNetParams) => {
    const net = await prisma.net.create({
      data: {
        name: name,
        description: description,
        authorID: user.id,
        parentID: parentID,
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
              description: `${name} is idle.`,
              bound: 1
            },
            {
              name: service.name,
              description: service.description,
              bound: 1
            },
            {
              name: output.name,
              description: output.description,
              bound: 1
            }
          ]
        },
        transitions: {
          create: [
            {
              name: "Initialized",
              description: `${name} was initialized.`,
              events: {
                create: [{
                  name: "Initialize",
                  description: `Initialize ${name}.`,
                  fields: {
                    create: initializeFields
                  }
                }]
              }
            },
            {
              name: transition.name,
              description: transition.description,
              events: {
                create: [{
                  name: event.name,
                  description: event.description,
                  fields: {
                    create: eventFields
                  }
                }]
              }
            },
            {
              name: "Finished",
              description: `${name} finished.`,
              events: {
                create: [{
                  name: emit.name,
                  description: emit.description,
                  fields: {
                    create: emitFields
                  }
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
        transitionName: transition.name
      },
      {
        fromPlace: false,
        placeName: service.name,
        transitionName: transition.name
      },
      {
        fromPlace: true,
        placeName: service.name,
        transitionName: "Finished"
      },
      {
        fromPlace: false,
        placeName: "Idle",
        transitionName: "Finished"
      },
      {
        fromPlace: false,
        placeName: output.name,
        transitionName: "Finished"
      }
    ];


    for (let arc of pumpArcs) {
      await join({
        netID: net.id,
        places: net.places,
        transitions: net.transitions,
        ...arc
      });
    }

    let device = await prisma.device.create({
      data: {
        name: name,
        description: description,
        authorID: user.id,
      }
    });

    await prisma.devicesOnNets.create({
      data: {
        deviceID: device.id,
        netID: net.id
      }
    })

    return net;

  };

  const allNets: Record<string, Pick<Net, "id"> & {
    places: Pick<Place, "id" | "name">[],
    transitions: Pick<Transition, "id" | "name">[]
  }> = {
    "two_position_three_way_valve": await makeValve("two position three way valve", "A two position three way valve."),
    "six_port_rheodyne_valve": await makeValve("six port rheodyne valve", "A six port rheodyne valve."),
    "ten_port_rheodyne_valve": await makeValve("ten port rheodyne valve", "A ten port rheodyne valve.")
  };

  for (let [name, params] of Object.entries(services)) {
    allNets[name] = await service(params);
  }

  const mfNetIDs = [
    allNets["two_position_three_way_valve"].id,
    allNets["six_port_rheodyne_valve"].id,
    allNets["ten_port_rheodyne_valve"].id,
    allNets["organic_pump"].id,
    allNets["aqueous_pump"].id,
    allNets["mixing_valve"].id,
    allNets["fraction_collector"].id,
    allNets["camera"].id
  ];
  const microfluidicNet = await prisma.net.create({
    data: {
      name: "a microfluidic injection system",
      description: `this system has 3 syringe pumps, 2 rheodyne valves, 1 three way valve, and a fraction collector`,
      authorID: user.id,
      children: {
        connect: mfNetIDs.map(id => ({ id }))
      }
    },
    include: {
      places: true,
      transitions: true
    }
  });

  const connections: ConnectionParams[] = [
    {
      name: "Aqueous injection inlet",
      description: "The aqueous injection inlet.",
      bound: 1,
      parentID: microfluidicNet.id,
      places: [
        {
          id: allNets["six_port_rheodyne_valve"].places.find(place => place.name === "Flow main")!.id
        },
        {
          id: allNets["aqueous_pump"].places.find(place => place.name === "Flow through syringe")!.id
        }
      ]
    },
    {
      name: "Organic injection inlet",
      description: "The organic injection inlet.",
      bound: 1,
      parentID: microfluidicNet.id,
      places: [
        {
          id: allNets["two_position_three_way_valve"].places.find(place => place.name === "Flow B")!.id
        },
        {
          id: allNets["ten_port_rheodyne_valve"].places.find(place => place.name === "Flow main")!.id
        }
      ]
    },
    {
      name: "Organic syringe flow",
      description: "The organic syringe is flowing.",
      bound: 1,
      parentID: microfluidicNet.id,
      places: [
        {
          id: allNets["organic_pump"].places.find(place => place.name === "Flow through syringe")!.id
        },
        {
          id: allNets["two_position_three_way_valve"].places.find(place => place.name === "Flow main")!.id
        }
      ]
    },
    {
      name: "Organic mixing flow",
      description: "Fluid is flowing through the organic mixing valve.",
      bound: 1,
      parentID: microfluidicNet.id,
      places: [
        {
          id: allNets["mixing_valve"].places.find(place => place.name === "Mixed")!.id
        },
        {
          id: allNets["two_position_three_way_valve"].places.find(place => place.name === "Flow A")!.id
        }
      ]
    },
    {
      name: "Aqueous injection loop",
      description: "Aqueous injection loop.",
      bound: 1,
      parentID: microfluidicNet.id,
      places: [
        {
          id: allNets["six_port_rheodyne_valve"].places.find(place => place.name === "Flow A")!.id
        },
        {
          id: allNets["six_port_rheodyne_valve"].places.find(place => place.name === "Flow main")!.id
        }
      ]
    },
    {
      name: "Organic injection loop",
      description: "Organic injection loop.",
      bound: 1,
      parentID: microfluidicNet.id,
      places: [
        {
          id: allNets["ten_port_rheodyne_valve"].places.find(place => place.name === "Flow A")!.id
        },
        {
          id: allNets["ten_port_rheodyne_valve"].places.find(place => place.name === "Flow main")!.id
        }
      ]
    }
  ];

  await connect(microfluidicNet.id, connections);

  async function makeValve(name: string, description: string) {
    const valve = await prisma.net.create({
      data: {
        name: name,
        description: description,
        authorID: user.id,
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
                  description: "Open port A.",
                  fields: {
                    create: [
                      {
                        name: "Delay",
                        type: "number"
                      }
                    ]
                  }
                }]
              }
            },
            {
              name: "B opened",
              description: "Position B was opened.",
              events: {
                create: [{
                  name: "Open B",
                  description: "Open port B.",
                  fields: {
                    create: [
                      {
                        name: "Delay",
                        type: "number"
                      }
                    ]
                  }
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

    let valveDevice = await prisma.device.create({
      data: {
        name: name,
        description: description,
        authorID: user.id,
      }
    });

    await prisma.devicesOnNets.create({
      data: {
        deviceID: valveDevice.id,
        netID: valve.id
      }
    })

    return valve;
  }

  async function join({ netID, placeName, transitionName, fromPlace, places, transitions }: {
    netID: string,
    fromPlace: boolean,
    placeName: string,
    transitionName: string,
    places: {
      id: string,
      name: string
    }[],
    transitions: {
      id: string,
      name: string
    }[]

  }) {
    const placeID = places.find(place => place.name === placeName)?.id;
    const transitionID = transitions.find(transition => transition.name === transitionName)?.id;
    if (placeID && transitionID) {
      return prisma.arc.create({
        data: {
          netID,
          placeID,
          transitionID,
          fromPlace
        }
      });
    }
  }

  type ConnectionParams = {
    name: string,
    description: string,
    bound: number,
    parentID: string
    places: {
      id: string,
    }[]
  }

  async function connect(parentID: string, params: ConnectionParams[]) {
    await prisma.net.update({
      where: {
        id: parentID
      },
      data: {
        placeInterfaces: {
          create: params.map(param => ({
              name: param.name,
              bound: param.bound,
              places: {
                connect: param.places.map(paramPlace => ({
                  id: paramPlace.id
                }))
              }
            })
          )
        }
      }
    });
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