import { z } from "zod";
import { prisma } from "~/db.server";

export const UserParameterSchema = z.object({
  name: z.string(),
  expression: z.string(),
  userId: z.string()
});

export const RunParameterSchema = z.object({
  name: z.string(),
  expression: z.string(),
  runId: z.string()
});

export const DeviceParameterSchema = z.object({
  name: z.string(),
  expression: z.string(),
  deviceId: z.string()
});

export const ParameterSchema = z.union([UserParameterSchema, RunParameterSchema, DeviceParameterSchema]);

export type UserParameter = z.infer<typeof UserParameterSchema>;
export type RunParameter = z.infer<typeof RunParameterSchema>;
export type DeviceParameter = z.infer<typeof DeviceParameterSchema>;
export type Parameter = z.infer<typeof ParameterSchema>;


export function createParameter(parameter: Parameter) {
  const result = ParameterSchema.parse(parameter);
  return prisma.parameter.create({ data: result });
}


export function updateParameter(id: string, parameter: Parameter) {
  const result = ParameterSchema.parse(parameter);
  return prisma.parameter.update({
    where: { id },
    data: result
  });
}

export function deleteParameter(id: string) {
  return prisma.parameter.delete({ where: { id } });
}

export function getParameter(id: string) {
  return prisma.parameter.findUnique({ where: { id } });
}

export function getRunParameters(runId: string) {
  return prisma.parameter.findMany({ where: { runId } });
}

export function getUserParameters(userId: string) {
  return prisma.parameter.findMany({ where: { userId } });
}

export function getDeviceParameters(deviceId: string) {
  return prisma.parameter.findMany({ where: { deviceId } });
}

export function getAllRunParameters(runId: string, deviceIds: string[], userId: string) {
  return prisma.parameter.findMany({
    where: {
      OR: [
        { runId },
        { userId },
        { deviceId: { in: deviceIds } }
      ]
    }
  });
}
