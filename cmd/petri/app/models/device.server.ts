import type { Device } from "@prisma/client";
import { z } from "zod";
import { prisma } from "~/db.server";

export const DeviceInputSchema = z.object({
  name: z.string(),
  authorID: z.string().cuid(),
  description: z.string(),
  netID: z.string().cuid()
});

export type DeviceInput = z.infer<typeof DeviceInputSchema>;

export async function createDevice(device: DeviceInput) {
  const { authorID, name, description, netID } = DeviceInputSchema.parse(device);
  return prisma.device.create({
    data: {
      authorID,
      name,
      description,
      nets: {
        connect: { id: netID }
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
            transitions: {}
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


const UpdateDeviceInputSchema = z.object({
  id: z.string().cuid(),
  name: z.string().optional(),
  description: z.string().optional()
});

export type UpdateDeviceInput = z.infer<typeof UpdateDeviceInputSchema>;

export async function updateDevice(input: UpdateDeviceInput) {
  const { id, name, description } = UpdateDeviceInputSchema.parse(input);
  return prisma.device.update({
    where: { id },
    data: {
      name,
      description
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
