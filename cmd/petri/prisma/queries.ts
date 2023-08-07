import { PrismaClient } from "@prisma/client";
import { createWriteStream, createReadStream } from "fs";
import { readdir } from "fs/promises";
import invariant from "tiny-invariant";

const createStream = createWriteStream("./sqlc/query.sql", { flags: "w" });

const prisma = new PrismaClient({
  log: [
    {
      emit: "event",
      level: "query"
    },
    {
      emit: "stdout",
      level: "error"
    },
    {
      emit: "stdout",
      level: "info"
    },
    {
      emit: "stdout",
      level: "warn"
    }
  ]
});

const stream = createStream;
prisma.$on("query", (e) => {
  // concatenate the "public"."User"."id" formatted table names and columns into "public.User.id" format
  // and do it for the table names as well
  const q = e.query.replace(/"public"\."(\w+)"\."(\w+)"/g, `"$2"`)
    .replace(/"public"\."(\w+)"/g, `"$1"`);
  stream.write("\n" + q);
});

type QueryResult = {
  [key: string]: any
}

type HeaderParams = {
  name: string,
  returns: ":many" | ":one" | ":exec" | ":execresult",
  call: (r: QueryResult) => Promise<any>
}

const query = async (result: QueryResult, { name, returns, call }: HeaderParams) => {
  stream.write(`-- name: ${name} ${returns}`);
  result[name] = await call(result);
  stream.write(`;\n`);
  invariant(result[name], "User not found");
};

const queries: HeaderParams[] = [
  {
    name: "FindUserByEmail",
    returns: ":one",
    call: async () => prisma.user.findUnique({
      where: {
        email: "jonathan@petri.local"
      }
    })
  },
  {
    name: "FindUserByID",
    returns: ":one",
    call: async (r: QueryResult) => prisma.user.findUnique({
      where: {
        id: r["FindUserByEmail"].id
      }
    })
  },
  {
    name: "ListNetsByAuthor",
    returns: ":many",
    call: async (r: QueryResult) => prisma.net.findMany({
      where: {
        authorID: r["FindUserByEmail"].id
      }
    })
  },
  {
    name: "FindNetByID",
    returns: ":one",
    call: async (r: QueryResult) => prisma.net.findUnique({
      where: {
        id: r["ListNetsByAuthor"][1].id
      },
      select: {
        id: true,
        name: true,
        description: true,
        initialMarking: true,
        deviceId: true
      }
    })
  },
  {
    name: "FindPlacesByNetID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.place.findMany({
      where: {
        nets: {
          some: {
            id: r["FindNetByID"].id
          }
        }
      },
      select: {
        id: true,
        name: true,
        bound: true
      }
    })
  },
  {
    name: "FindTransitionsByNetID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.transition.findMany({
      where: {
        nets: {
          some: {
            id: r["FindNetByID"].id
          }
        }
      },
      select: {
        id: true,
        name: true
      }
    })
  },
  {
    name: "FindArcsByNetID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.arc.findMany({
      where: {
        netID: r["FindNetByID"].id
      },
      select: {
        id: true,
        fromPlace: true,
        placeID: true,
        transitionID: true
      }
    })
  },
  {
    name: "FindNetsByParentID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.net.findMany({
      where: {
        parentID: r["FindNetByID"].id
      },
      select: {
        id: true,
        name: true,
        description: true,
        initialMarking: true,
        deviceId: true
      }
    })
  },
  {
    name: "FindDevicesByUser",
    returns: ":many",
    call: async (r: QueryResult) => prisma.device.findMany({
      where: {
        authorID: r["FindUserByEmail"].id
      },
      select: {
        id: true,
        name: true,
        description: true
      }
    })
  },
  {
    name: "FindNetsByDeviceID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.net.findMany({
      where: {
        deviceId: r["FindDevicesByUser"][0].id
      },
      select: {
        id: true,
        name: true,
        description: true,
        initialMarking: true,
        deviceId: true
      }
    })
  },
  {
    name: "FindInstancesByUserID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.deviceInstance.findMany({
      where: {
        authorID: r["FindUserByEmail"].id
      },
      select: {
        id: true,
        name: true,
        addr: true
      }
    })
  },
  {
    name: "FindInstancesByDeviceID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.deviceInstance.findMany({
      where: {
        deviceId: r["FindNetByID"].id
      },
      select: {
        id: true,
        name: true,
        addr: true
      }
    })
  },
  {
    name: "FindInstancesByAddr",
    returns: ":many",
    call: async (r: QueryResult) => prisma.deviceInstance.findMany({
      where: {
        addr: r["FindInstancesByUserID"][0].addr
      },
      select: {
        id: true,
        name: true,
        addr: true
      }
    })
  },
  {
    name: "FindTransitionsWithNonNullEventsByNetID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.transition.findMany({
      where: {
        nets: {
          some: {
            id: r["FindNetByID"].id
          }
        },
        events: {
          some: {}
        }
      },
      select: {
        id: true,
        name: true
      }
    })
  },
  {
    name: "FindEventsByTransitionID",
    returns: ":many",
    call: async (r: QueryResult) => prisma.event.findMany({
      where: {
        transitions: {
          some: {
            id: r["FindTransitionsByNetID"][0].id
          }
        }
      },
      select: {
        id: true,
        name: true
      }
    })
  },
  {
    name: "FindFieldsByEvent",
    returns: ":many",
    call: async (r: QueryResult) => prisma.field.findMany({
      where: {
        eventID: r["FindEventsByTransitionID"][0].id
      }
    })
  }
];

const executeQueries = async (params: HeaderParams[]) => {
  const result: QueryResult = {};
  for (const param of params) {
    await query(result, param);
  }
  return result;
};

// a function that finds the more current folder in the ./prisma/migrations folder and copies the migration file to ./sqlc/schema.sql
// this is a workaround for the fact that sqlc does not support prisma migrations
const copyMigration = async () => {
  const migrations = await readdir("./prisma/migrations").then((files) => {
    return files.filter((file) => !file.endsWith(".toml"));
  });

  const latestMigration = migrations[migrations.length - 1];
  const migrationPath = `./prisma/migrations/${latestMigration}`;
  const migration = await readdir(migrationPath);
  const migrationFile = migration.filter((file) => file.endsWith(".sql"))[0];
  const migrationFilePath = `${migrationPath}/${migrationFile}`;
  const schemaPath = "./sqlc/schema.sql";
  const schema = createWriteStream(schemaPath, { flags: "w" });
  const migrationStream = createReadStream(migrationFilePath, { flags: "r" });
  migrationStream.pipe(schema);
  migrationStream.on("end", () => {
    console.log("copied migration file");
  });
};


const main = async () => {
  await executeQueries(queries).then(async (result) => {
    await copyMigration();
    // write results to a json file
    const json = createWriteStream("./sqlc/results.json", { flags: "w" });
    json.write(JSON.stringify(result, null, 2));
    json.end();
  });
};

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
    stream.end();
  });