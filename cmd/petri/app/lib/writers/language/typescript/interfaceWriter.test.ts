import { describe, expect, it } from "vitest";

import Writer from "./interfaceWriter";
import { PetriNet } from "~/util/petrinet";

import { promises as fs } from "fs";

const testNet = new PetriNet({
  arcs: [
    {
      placeID: "idle",
      transitionID: "pump",
      fromPlace: true
    },
    {
      placeID: "pumping",
      transitionID: "pump",
      fromPlace: false
    },
    {
      placeID: "pumping",
      transitionID: "finish",
      fromPlace: true
    },
    {
      placeID: "idle",
      transitionID: "finish",
      fromPlace: false
    }
  ],
  children: [],
  placeInterfaces: [],
  places: [
    {
      id: "idle",
      name: "idle",
      bound: 1
    },
    {
      id: "pumping",
      name: "pumping",
      bound: 1
    }
  ],
  transitionInterfaces: [],
  transitions: [{
    id: "pump",
    name: "pump",
    events: [
      {
        id: "pump",
        name: "pump",
        fields: [
          {
            id: "flowRate",
            name: "flowRate",
            type: "number"
          }
        ]
      }
    ]
  },
    {
      id: "finish",
      name: "finish",
      events: [
        {
          id: "finish",
          name: "finish",
          fields: [
            {
              id: "dispensedVolume",
              name: "dispensedVolume",
              type: "number"
            },
            {
              id: "remainingVolume",
              name: "remainingVolume",
              type: "number"
            }
          ]
        }
      ]
    }
  ],
  id: "test",
  name: "pump",
  initialMarking: [1, 0],
  description: "test",
  devices: []
});

async function readFileIntoString(filepath: string): Promise<string> {
  return await fs.readFile(filepath, "utf-8");
}

const testFile = "./app/lib/writers/language/typescript/interfaceWriter.test.data.ts";

const testFileContents = readFileIntoString(testFile);

describe.concurrent("InterfaceWriter", () => {
  it("should have a language", () => {
    const writer = new Writer();
    expect(writer.language).toBe("TypeScript");
  });
  it("should write a petri net", async () => {
    const writer = new Writer();
    const contents = await readFileIntoString(testFile);
    expect(writer.writePetriNet(testNet)).toBe(contents);
  });
});
