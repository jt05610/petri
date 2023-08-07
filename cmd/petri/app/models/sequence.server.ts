import type { Sequence, DeviceEvent, DeviceEventConstant, SequenceEvent } from "@prisma/client";
import { z } from "zod";
import { prisma } from "~/db.server";

// Device event constant CRUD operations
export const DeviceEventConstantInputSchema = z.object({
  fieldID: z.string().cuid(),
  value: z.string()
});

// Device event constant CRUD operations
export const DeviceEventConstantInputDisplaySchema = z.object({
  fieldID: z.string().cuid(),
  constant: z.boolean(),
  fieldName: z.string(),
  value: z.string()
});

export type DeviceEventConstantInputDisplay = z.infer<typeof DeviceEventConstantInputDisplaySchema>;

export type DeviceEventConstantInput = z.infer<typeof DeviceEventConstantInputSchema>;

export const AddDeviceEventConstantToEventInputSchema = DeviceEventConstantInputSchema.extend({
  deviceEventId: z.string().cuid()
});

export type AddDeviceEventConstantToEventInput = z.infer<typeof AddDeviceEventConstantToEventInputSchema>;

export async function addDeviceEventConstant(input: AddDeviceEventConstantToEventInput) {
  const { fieldID, value, deviceEventId } = AddDeviceEventConstantToEventInputSchema.parse(input);
  return prisma.deviceEvent.update({
    where: { id: deviceEventId },
    data: {
      constants: {
        create: {
          field: {
            connect: {
              id: fieldID
            }
          },
          value
        }
      }
    }
  });
}

export const UpdateDeviceEventConstantSchema = z.object({
  deviceEventId: z.string().cuid(),
  constantID: z.string().cuid(),
  value: z.string()
});

export type UpdateDeviceEventConstantInput = z.infer<typeof UpdateDeviceEventConstantSchema>;

export async function updateDeviceEventConstant(input: UpdateDeviceEventConstantInput) {
  const { constantID, deviceEventId, value } = UpdateDeviceEventConstantSchema.parse(input);
  return prisma.deviceEvent.update({
    where: { id: deviceEventId },
    data: {
      constants: {
        update: {
          where: { id: constantID },
          data: {
            value
          }
        }
      }
    }
  });
}

export const DeleteDeviceEventConstantSchema = z.object({
  deviceEventId: z.string().cuid(),
  constantID: z.string().cuid()
});

export type DeleteDeviceEventConstantInput = z.infer<typeof DeleteDeviceEventConstantSchema>;

export async function deleteDeviceEventConstant(input: DeleteDeviceEventConstantInput) {
  const { constantID, deviceEventId } = DeleteDeviceEventConstantSchema.parse(input);
  return prisma.deviceEvent.update({
    where: { id: deviceEventId },
    data: {
      constants: {
        delete: {
          id: constantID
        }
      }
    }
  });
}

// Device event CRUD operations

export const DeviceEventInputSchema = z.object({
  deviceId: z.string().cuid(),
  eventID: z.string().cuid(),
  constants: z.array(DeviceEventConstantInputSchema)
});

export const DeviceInputDisplaySchema = z.object({
    deviceId: z.string().cuid(),
    eventID: z.string().cuid(),
    eventName: z.string(),
    constants: z.array(DeviceEventConstantInputDisplaySchema)
  }
);

export type DeviceInputDisplay = z.infer<typeof DeviceInputDisplaySchema>;

export const AddDeviceEventToSequenceEventInputSchema = DeviceEventInputSchema.extend({
  sequenceEventId: z.string().cuid()
});


export type DeviceEventInput = z.infer<typeof DeviceEventInputSchema>;

export type AddDeviceEventToSequenceEventInput = z.infer<typeof AddDeviceEventToSequenceEventInputSchema>;

export async function addDeviceEventToSequenceEvent(input: AddDeviceEventToSequenceEventInput) {
  const { deviceId, sequenceEventId, eventID, constants } = AddDeviceEventToSequenceEventInputSchema.parse(input);
  return prisma.deviceEvent.create({
    data: {
      device: {
        connect: {
          id: deviceId
        }
      },
      sequenceEvent: {
        connect: {
          id: sequenceEventId
        }
      },
      event: {
        connect: {
          id: eventID
        }
      },
      constants: {
        create: constants
      }
    }
  });
}

export const RemoveDeviceEventFromSequenceEventInputSchema = z.object({
  sequenceEventId: z.string().cuid(),
  deviceEventId: z.string().cuid()
});

export type RemoveDeviceEventFromSequenceEventInput = z.infer<typeof RemoveDeviceEventFromSequenceEventInputSchema>;

export async function removeDeviceEventFromSequenceEvent(input: RemoveDeviceEventFromSequenceEventInput) {
  const { sequenceEventId, deviceEventId } = RemoveDeviceEventFromSequenceEventInputSchema.parse(input);
  return prisma.sequenceEvent.update({
    where: { id: sequenceEventId },
    data: {
      events: {
        delete: {
          id: deviceEventId
        }
      }
    }
  });
}

// Sequence event CRUD operations

export const SequenceEventInputSchema = z.object({
  name: z.string(),
  notes: z.array(z.string()),
  events: z.array(DeviceEventInputSchema)
});

export const SequenceEventInputDisplaySchema = z.object({
  name: z.string(),
  notes: z.array(z.string()),
  events: z.array(DeviceInputDisplaySchema)
});

export type SequenceEventInputDisplay = z.infer<typeof SequenceEventInputDisplaySchema>;

export type SequenceEventInput = z.infer<typeof SequenceEventInputSchema>;

export const AddSequenceEventInputSchema = SequenceEventInputSchema.extend({
  sequenceID: z.string().cuid()
});

export type AddSequenceEventInput = z.infer<typeof AddSequenceEventInputSchema>;

export async function addSequenceEvent(input: AddSequenceEventInput) {
  const { name, notes, events, sequenceID } = AddSequenceEventInputSchema.parse(input);
  return prisma.sequenceEvent.create({
    data: {
      name,
      notes,
      sequenceID,
      events: {
        create: events.map((event) => ({
          device: {
            connect: {
              id: event.deviceId
            }
          },
          event: {
            connect: {
              id: event.eventID
            }
          },
          constants: {
            create: event.constants
          }
        }))
      }
    }
  });
}

export const UpdateSequenceEventInputSchema = z.object({
  sequenceEventId: z.string().cuid(),
  name: z.string(),
  notes: z.array(z.string())
});

export type UpdateSequenceEventInput = z.infer<typeof UpdateSequenceEventInputSchema>;

export async function updateSequenceEvent(input: UpdateSequenceEventInput) {
  const { sequenceEventId, name, notes } = UpdateSequenceEventInputSchema.parse(input);
  return prisma.sequenceEvent.update({
    where: { id: sequenceEventId },
    data: {
      name,
      notes
    }
  });
}

// Sequence CRUD operations

export const SequenceInputSchema = z.object({
  name: z.string(),
  description: z.string(),
  netID: z.string().cuid(),
  events: z.array(SequenceEventInputSchema)
});

export const SequenceInputDisplaySchema = z.object({
  name: z.string(),
  description: z.string(),
  netID: z.string().cuid(),
  deviceNames: z.array(z.string()),
  events: z.array(SequenceEventInputDisplaySchema)
});

export type SequenceInputDisplay = z.infer<typeof SequenceInputDisplaySchema>;

export type SequenceInput = z.infer<typeof SequenceInputSchema>;

export async function addSequence(input: SequenceInput) {
  const { name, description, netID, events } = SequenceInputSchema.parse(input);
  return prisma.sequence.create({
    data: {
      name,
      description,
      net: {
        connect: {
          id: netID
        }
      },
      events: {
        create: [
          ...events.map((event) => ({
            name: event.name,
            notes: event.notes,
            events: {
              create: event.events.map((deviceEvent) => ({
                device: {
                  connect: {
                    id: deviceEvent.deviceId
                  }
                },
                event: {
                  connect: {
                    id: deviceEvent.eventID
                  }
                },
                constants: {
                  create: deviceEvent.constants
                }
              }))
            }
          }))
        ]
      }
    }
  });
}

export const AddSequenceEventToSequenceInputSchema = SequenceEventInputSchema.extend({
    sequenceID: z.string().cuid()
  }
);

export type AddSequenceEventToSequenceInput = z.infer<typeof AddSequenceEventToSequenceInputSchema>;

export async function addSequenceEventToSequence(input: AddSequenceEventToSequenceInput) {
  const { sequenceID, name, notes, events } = AddSequenceEventToSequenceInputSchema.parse(input);
  return prisma.sequence.update({
    where: { id: sequenceID },
    data: {
      events: {
        create: [
          {
            name,
            notes,
            events: {
              create: events.map((deviceEvent) => ({
                device: {
                  connect: {
                    id: deviceEvent.deviceId
                  }
                },
                event: {
                  connect: {
                    id: deviceEvent.eventID
                  }
                },
                constants: {
                  create: deviceEvent.constants
                }
              }))
            }
          }
        ]
      }
    }
  });
}

export const removeSequenceEventFromSequenceInputSchema = z.object({
  sequenceID: z.string().cuid(),
  sequenceEventID: z.string().cuid()
});

export type RemoveSequenceEventFromSequenceInput = z.infer<typeof removeSequenceEventFromSequenceInputSchema>;

export async function removeSequenceEventFromSequence(input: RemoveSequenceEventFromSequenceInput) {
  const { sequenceID, sequenceEventID } = removeSequenceEventFromSequenceInputSchema.parse(input);
  return prisma.sequence.update({
    where: { id: sequenceID },
    data: {
      events: {
        delete: {
          id: sequenceEventID
        }
      }
    }
  });
}

export const updateSequenceInputSchema = z.object({
  sequenceID: z.string().cuid(),
  name: z.string().optional(),
  description: z.string().optional()
});

export type UpdateSequenceInput = z.infer<typeof updateSequenceInputSchema>;

export async function updateSequence(input: UpdateSequenceInput) {
  const { sequenceID, name, description } = updateSequenceInputSchema.parse(input);
  return prisma.sequence.update({
    where: { id: sequenceID },
    data: {
      name,
      description
    }
  });
}

export const removeSequenceInputSchema = z.object({
  sequenceID: z.string().cuid()
});

export type RemoveSequenceInput = z.infer<typeof removeSequenceInputSchema>;

export async function removeSequence(input: RemoveSequenceInput) {
  const { sequenceID } = removeSequenceInputSchema.parse(input);
  return prisma.sequence.delete({
    where: { id: sequenceID }
  });
}

export const getSequenceInputSchema = z.object({
  sequenceID: z.string().cuid()
});

export type GetSequenceInput = z.infer<typeof getSequenceInputSchema>;

export type DeviceEventConstantDetails = Pick<DeviceEventConstant, "id" | "fieldID" | "value">

export type DeviceEventDetails = Pick<DeviceEvent, "id"> & {
  constants: DeviceEventConstantDetails[]
}

export type SequenceEventDetails = Pick<SequenceEvent, "id" | "name" | "notes"> & {
  events: DeviceEventDetails[]
}

export type SequenceDetails = Pick<Sequence, "id" | "name" | "description"> & {
  events: SequenceEventDetails[]
}

export async function getSequenceDetails(input: GetSequenceInput): Promise<SequenceDetails> {
  const { sequenceID } = getSequenceInputSchema.parse(input);
  return prisma.sequence.findUniqueOrThrow({
    where: { id: sequenceID },
    select: {
      id: true,
      name: true,
      description: true,
      events: {
        select: {
          id: true,
          name: true,
          notes: true,
          events: {
            select: {
              id: true,
              event: {
                select: {
                  id: true,
                  name: true,
                  fields: {
                    select: {
                      id: true,
                      name: true,
                      type: true
                    }
                  }
                }
              },
              constants: {
                select: {
                  id: true,
                  fieldID: true,
                  value: true
                }
              }
            }
          }
        }
      }
    }
  });
}

export const SequenceListInputSchema = z.object({
  netID: z.string().cuid()
});

export type SequenceListInput = z.infer<typeof SequenceListInputSchema>;

export type SequenceListItem = Pick<Sequence, "id" | "name" | "description">;

export async function listSequences(input: SequenceListInput): Promise<SequenceListItem[]> {
  const { netID } = SequenceListInputSchema.parse(input);
  return prisma.sequence.findMany({
    where: { netID },
    select: {
      id: true,
      name: true,
      description: true
    }
  });
}

