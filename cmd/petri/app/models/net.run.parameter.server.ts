import { prisma } from "~/db.server";
import type { Parameter } from "~/models/parameter";
import { ParameterSchema } from "~/models/parameter";


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
