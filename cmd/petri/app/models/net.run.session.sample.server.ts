import { z } from "zod";
import type { Datum } from "@prisma/client";
import { prisma } from "~/db.server";

const sampleInputSchema = z.object({
  name: z.string(),
  group: z.string(),
  batchId: z.string().cuid().optional(),
  runID: z.string().cuid().optional(),
  params: z.record(z.string().or(z.number()))
});

const batchInputSchema = z.object({
  name: z.string(),
  runID: z.string().cuid(),
  samples: z.array(sampleInputSchema).min(1)
});

export type BatchInput = z.infer<typeof batchInputSchema>;

export type SampleInput = z.infer<typeof sampleInputSchema>;

export async function createSample(userID: string, input: SampleInput) {
  const { name, group, params, batchId, runID } = sampleInputSchema.parse(input);
  return prisma.sample.create({
    data: {
      user: {
        connect: {
          id: userID
        }
      },
      batch: {
        connect: {
          id: batchId
        }
      },
      run: {
        connect: {
          id: runID
        }
      },
      name,
      group,
      params
    }
  });
}


export const GetSamplesInput = z.object({
  batchId: z.string().cuid().optional(),
  runID: z.string().cuid().optional(),
  group: z.string().optional(),
  name: z.string().optional(),
  createdBefore: z.date().optional(),
  createdAfter: z.date().optional(),
  createdBy: z.string().cuid().optional()
});

export type GetSamplesInput = z.infer<typeof GetSamplesInput>;

export async function getSamples(input: GetSamplesInput): Promise<SampleDetails[]> {
  const { batchId, runID, group, name, createdBefore, createdAfter, createdBy } = GetSamplesInput.parse(input);
  return prisma.sample.findMany({
    select: {
      id: true,
      name: true,
      group: true,
      params: true
    },
    where: {
      batch: {
        id: batchId
      },
      run: {
        id: runID
      },
      group,
      name: {
        contains: name
      },
      createdAt: {
        lte: createdBefore,
        gte: createdAfter
      },
      user: {
        id: createdBy
      }
    }
  }).then((samples) => {
    return samples.map((sample) => {
      const paramsRecord = sample.params as Record<string, string | number>;

      return {
        id: sample.id,
        name: sample.name,
        group: sample.group,
        ...paramsRecord
      };

    });
  });
}

export type SampleDetails = {
  id: string;
  name: string;
  group: string;
} & Record<string, string | number>

export const ListDataInputSchema = z.object({
  sessionID: z.string().cuid()
});

export async function createBatch(userID: string, runID: string, input: BatchInput) {
  const { name, samples } = batchInputSchema.parse(input);
  return prisma.batch.create({
    data: {
      name,
      samples: {
        createMany: {
          data: samples.map((sample) => {
            return {
              name: sample.name,
              group: sample.group,
              params: sample.params,
              userID: userID,
              runID: runID
            };
          })
        }
      }
    }
  });
}

export async function getBatch(id: string) {
  return prisma.batch.findUnique({
    where: {
      id
    }
  });
}

export type ListDataInput = z.infer<typeof ListDataInputSchema>;

export type DataListItem = Pick<Datum, "id" | "createdAt" | "updatedAt" | "value" | "instanceID">

export async function listData(input: ListDataInput): Promise<DataListItem[]> {
  const { sessionID } = ListDataInputSchema.parse(input);
  return prisma.session.findUniqueOrThrow({
    where: {
      id: sessionID
    },
    include: {
      data: {
        include: {
          instance: true,
          event: true
        }
      }
    }
  }).then(session => {
    return session.data;
  });
}
