import type { ActionArgs } from "@remix-run/node";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { Form, useLoaderData, useNavigate } from "@remix-run/react";
import { useContextSelector } from "use-context-selector";
import type { FormEvent } from "react";
import { Suspense, useRef } from "react";
import { PetriNetContext } from "~/context";
import { useMutation, useQuery } from "@apollo/client";
import type { RunDetails, ConstantDetails, ActionDetails } from "~/models/net.run.server";
import { getRunDetails } from "~/models/net.run.server";

import type {
  DeviceInstanceInput
} from "~/models/__generated__/graphql";
import {
  DevicesDocument,
  NewSessionDocument,
  SessionsDocument
} from "~/models/__generated__/graphql";
import { parse } from "@conform-to/zod";
import { z } from "zod";

export const loader = async ({ params, request }: ActionArgs) => {
  const user = await requireUser(request);
  const sequenceID = params.sequenceID;
  invariant(sequenceID, "sequenceID is required");
  const run = await getRunDetails({ runID: sequenceID });
  const runParams = getParameters(run);
  return { run, runParams, user };
};

type EventConstant = {
  event: ActionDetails["event"]
  order: number
  constants: ConstantDetails[]
  deviceID: string
}

type Parameter = {
  order: number
  eventID: string
  deviceID: string
  fieldID: string
  fieldName: string
  fieldType: "string" | "number" | "boolean" | "date" | string
}

function getParameters(run: RunDetails): Record<string, Parameter[]> {
  // make all EventConstants from the run
  const eventConstants: EventConstant[] = run.steps.map(({ action, order }) => {
    return {
      order,
      deviceID: action.device.id,
      event: action.event!,
      constants: action.constants
    };
  });
  const params: Record<string, Parameter[]> = {};
  eventConstants.forEach(({ order, event, constants, deviceID }) => {
    event.fields.forEach((field) => {
      const constant = constants.find((constant) => constant.fieldID === field.id);
      if (constant) return;
      if (!params[deviceID]) params[deviceID] = [];
      params[deviceID].push({
        order,
        eventID: event.id,
        deviceID: deviceID,
        fieldID: field.id,
        fieldName: field.name,
        fieldType: field.type
      });
    });
  });
  return params;
}

export default function PlaySequence() {
  const { run, runParams, user } = useLoaderData<typeof loader>();
  const net = useContextSelector(PetriNetContext, (ctx) => ctx?.petriNet);
  const marking = useContextSelector(PetriNetContext, (ctx) => ctx?.marking);
  const paramRecord = useRef<Record<string, any>>({});
  const sessions = useQuery(
    SessionsDocument, { variables: { runID: run.id } }
  );

  const devices = useQuery(
    DevicesDocument, { variables: { filter: "" } }
  );
  const [newSession] = useMutation(NewSessionDocument);
  const navigate = useNavigate();

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const instances = formData.getAll("instances");

    if (!instances) {
      console.log("No instances");
      return;
    }
    console.log(instances);
    formData.set("sequenceID", run.id);
    formData.set("userID", user.id);
    formData.set("parameters", JSON.stringify(paramRecord.current));
    console.log(formData);
    const submission = parse(formData, {
      schema: z.object({
        sequenceID: z.string(),
        userID: z.string(),
        parameters: z.preprocess((s) => JSON.parse(s as string), z.record(z.any()).optional()),
        instances: z.preprocess((instStrings) => {
          const instanceStrings = instStrings as string[];
          return instanceStrings.map((instString) => {
            return JSON.parse(instString.toString()) as DeviceInstanceInput;
          });
        }, z.array(z.object({
          deviceID: z.string(),
          instanceID: z.string()
        })))
      })
    });
    if (!submission.value || submission.intent !== "submit") {
      console.log(submission.error);
      return;
    }
    console.log(submission.value);
    const session = await newSession({
      variables: { input: submission.value }
    }).then((session) => {
      return session.data?.newSession;
    });
    if (!session) return;
    const sessionID = session.id;
    const url = `/control/${net!.net.id}/sequences/${run.id}/play/${sessionID}`;
    navigate(url);
  }

  return (
    <div>
      <h3>Active sessions</h3>
      <Suspense fallback={<div>Loading...</div>}>
        {!sessions.loading && sessions.data && sessions.data.sessions.map((session) => {
            return (
              <div key={session.id}>
                {session.id}
              </div>
            );
          }
        )}
      </Suspense>
      <Suspense fallback={<div>Loading...</div>}>
        {net && marking &&
          <Form onSubmit={handleSubmit}>
            {net.childDeviceEvents(marking!).map(({ id, name, instances, events }, index) => {
              return (
                <div
                  key={index}
                  className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                >
                  <h2 className={"text-xl font-bold"}>{name}</h2>
                  <select
                    defaultValue={""}
                    className={"rounded-full p-2"}
                    name={`instances`}
                    key={"instanceID"}
                  >
                    <option value={""}>Select a device</option>
                    {devices.data && devices.data.devices && devices.data.devices.find((device) => device.id === id) &&
                      devices.data.devices.find((device) => device.id === id)!.instances.map((instance) => {
                          return (
                            <option key={instance.id} value={JSON.stringify({ instanceID: instance.id, deviceID: id })}>
                              {instance.id}
                            </option>
                          );
                        }
                      )}
                  </select>
                  <h1 className={"text-lg font-semibold"}>Parameters</h1>
                  {runParams && runParams[id] && runParams[id].map(({
                                                                      order,
                                                                      eventID,
                                                                      fieldID,
                                                                      fieldName,
                                                                      fieldType
                                                                    }, index) => {
                    return (
                      <div
                        className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                        key={index}
                      >
                        <div
                          className={"m-2 flex flex-row space-x-2 items-center"}
                        >
                          <label
                            className={"text-sm w-1/4 font-medium"}
                            htmlFor={fieldID}
                          >
                            {`Step ${order} ${fieldName}`}
                          </label>
                          <input
                            className={"w-3/4 rounded-full p-2"}
                            name={`${order}.${eventID}.${fieldName}`}
                            type={fieldType}
                            onChange={(e) => {
                              e.preventDefault();
                              // turn the number into a json object of {key: value}
                              // and add it to the parameters
                              const key = `${fieldName}.${order}.${eventID}`;
                              paramRecord.current[key] = e.currentTarget.value;
                            }}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>
              );
            })}
            <button type={"submit"}>
              New session
            </button>
          </Form>
        }
      </Suspense>
    </div>
  );
};