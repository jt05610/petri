import { resolve } from "path";

import * as TJS from "typescript-json-schema";

function makeSchema() {
// optionally pass ts compiler options
// optionally pass a base path

  const program = TJS.getProgramFromFiles(
    [resolve("app/models/net.server.ts")],
  );
  const generator = TJS.buildGenerator(program);
  const schema = generator?.getSchemaForSymbol("PlaceInput");
  console.log(schema);

}

makeSchema();
