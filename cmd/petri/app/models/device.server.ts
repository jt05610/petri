import type { Device } from "@prisma/client";
import { z } from "zod";
import { prisma } from "~/db.server";

export const DeviceInputSchema = z.object({
  name: z.string(),
  authorID: z.string().cuid(),
  description: z.string(),
  netIDs: z.array(z.string().cuid()).optional()
});

export type DeviceInput = z.infer<typeof DeviceInputSchema>;

export async function createDevice(device: DeviceInput) {
  const { authorID, name, description, netIDs } = DeviceInputSchema.parse(device);
  return prisma.device.create({
    data: {
      authorID,
      name,
      description,
      nets: {
        connect: netIDs?.map((id) => ({ id })) ?? []
      }
    }
  });
}

export const GetDeviceInputSchema = z.object({
  id: z.string().cuid()
});

export type GetDeviceInput = z.infer<typeof GetDeviceInputSchema>;

export type DeviceDetails = Pick<Device, "id" | "name" | "description" | "createdAt" | "updatedAt"> & {
  instances: {
    id: string;
    addr: string;
    language: string;
    name: string;
  }[]
  nets: {
    id: string
    name: string;
    description: string | null
    createdAt: Date;
    updatedAt: Date;
    transitions: {
      id: string;
      name: string;
      description: string | null
      createdAt: Date;
      updatedAt: Date;
    }[]
  }[]
}

export async function getDevice(inputs: GetDeviceInput): Promise<DeviceDetails> {
  const { id } = GetDeviceInputSchema.parse(inputs);
  return prisma.device.findFirst(
    {
      where: { id },
      include: {
        nets: {
          select: {
            id: true,
            name: true,
            description: true,
            createdAt: true,
            updatedAt: true,
            transitions: {
              select: {
                id: true,
                name: true,
                description: true,
                createdAt: true,
                updatedAt: true

              }
            }
          }
        },
        instances: {
          select: {
            id: true,
            addr: true,
            language: true,
            name: true
          }
        }
      }
    }
  ).then((device) => {
    if (!device) throw new Error("Device not found");
    return device;
  });
}

export type DeviceListItem = Pick<Device, "id" | "name" | "description" | "updatedAt">

export const ListDevicesInputSchema = z.object({
  authorID: z.string().cuid(),
  netID: z.string().cuid().optional()
});

export type ListDevicesInput = z.infer<typeof ListDevicesInputSchema>;

export async function listDevices(input: ListDevicesInput): Promise<DeviceListItem[]> {
  const { authorID, netID } = ListDevicesInputSchema.parse(input);
  return prisma.device.findMany({
    where: { authorID, nets: { some: { id: netID } } },
    select: {
      id: true,
      name: true,
      description: true,
      updatedAt: true
    }
  });
}


export const UpdateDeviceInputSchema = z.object({
  id: z.string().cuid(),
  name: z.string().optional(),
  description: z.string().optional(),
  netIDs: z.array(z.string().cuid()).optional()
});

export type UpdateDeviceInput = z.infer<typeof UpdateDeviceInputSchema>;

export async function updateDevice(input: UpdateDeviceInput) {
  const { id, name, description, netIDs } = UpdateDeviceInputSchema.parse(input);
  return prisma.device.update({
    where: { id },
    data: {
      name,
      description,
      nets: {
        connect: netIDs?.map((id) => ({ id })) ?? []
      }
    }
  });
}

export const DeleteDeviceInputSchema = z.object({
  id: z.string().cuid()
});

export type DeleteDeviceInput = z.infer<typeof DeleteDeviceInputSchema>;

export async function deleteDevice(input: DeleteDeviceInput) {
  const { id } = DeleteDeviceInputSchema.parse(input);
  return prisma.device.delete({ where: { id } });
}
