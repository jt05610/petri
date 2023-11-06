import { z } from "zod";


export const UserParameterSchema = z.object({
  name: z.string(),
  expression: z.string(),
  userId: z.string()
});

export const RunParameterSchema = z.object({
  name: z.string(),
  expression: z.string(),
});

export const UpdateRunParameterSchema = z.object({
  id: z.string().cuid(),
  name: z.string(),
  expression: z.string()
});

export const DeviceParameterSchema = z.object({
  name: z.string(),
  expression: z.string(),
  deviceId: z.string()
});

export const UpdateDeviceParameterSchema = z.object({
  id: z.string().cuid(),
  name: z.string(),
  expression: z.string()
});

export const UpdateUserParameterSchema = z.object({
  id: z.string().cuid(),
  name: z.string(),
  expression: z.string()
});

export const ParameterSchema = z.union([UserParameterSchema, RunParameterSchema, DeviceParameterSchema]);

export const UpdateParameterSchema = z.union([UpdateUserParameterSchema, UpdateRunParameterSchema, UpdateDeviceParameterSchema]);

export type UserParameter = z.infer<typeof UserParameterSchema>;
export type RunParameter = z.infer<typeof RunParameterSchema>;
export type DeviceParameter = z.infer<typeof DeviceParameterSchema>;
export type Parameter = z.infer<typeof ParameterSchema>;

export type UpdateUserParameter = z.infer<typeof UpdateUserParameterSchema>;
export type UpdateRunParameter = z.infer<typeof UpdateRunParameterSchema>;
export type UpdateDeviceParameter = z.infer<typeof UpdateDeviceParameterSchema>;
export type UpdateParameter = z.infer<typeof UpdateParameterSchema>;