import type { ActionArgs } from "@remix-run/node";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { Form, useLoaderData, useNavigate } from "@remix-run/react";
import { useContextSelector } from "use-context-selector";
import type { FormEvent } from "react";
import { Suspense, useState } from "react";
import { useMutation, useQuery } from "@apollo/client";
import { getRunDetails } from "~/models/net.run.server";
import JSONExport, { newSessionRequest } from "~/util/json_export";
import type { DeviceInstanceInput } from "~/models/__generated__/graphql";
import {
  DevicesDocument,
  NewSessionDocument,
  SessionsDocument
} from "~/models/__generated__/graphql";
import { parse } from "@conform-to/zod";
import { z } from "zod";
import { PetriNetContext } from "~/lib/context/petrinet";

export const loader = async ({ params, request }: ActionArgs) => {
  const user = await requireUser(request);
  const sequenceID = params.sequenceID;
  invariant(sequenceID, "sequenceID is required");
  const run = await getRunDetails({ runID: sequenceID });
  // create an array of Devices with only the unique ones
  const runDevices = run.steps.map(({ action }) => action.device).filter((device, index, self) => {
    return self.findIndex((d) => d.id === device.id) === index;
  });
  return { run, user, runDevices };
};


export default function PlaySequence() {
  const { run, user, runDevices } = useLoaderData<typeof loader>();
  const [reqJSON, setReqJSON] = useState<Object>({});
  const {
    net,
    marking
  } = useContextSelector(PetriNetContext, (ctx) => ctx ? {
    net: ctx.petriNet.net,
    marking: ctx.petriNet.marking
  } || {} : {
    net: undefined,
    marking: undefined
  });
  const sessions = useQuery(
    SessionsDocument, { variables: { runID: run.id } }
  );

  const devices = useQuery(
    DevicesDocument, { variables: { filter: "" }, pollInterval: 1000 }
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
    console.log(formData);
    const submission = parse(formData, {
      schema: z.object({
        sequenceID: z.string(),
        userID: z.string(),
        instances: z.preprocess((instStrings) => {
          // instStrings can either be one or many strings, so we need to check if it is a single string and convert to an array of 1 if so
          const instanceStrings: string[] = [];
          if (typeof instStrings === "string") {
            instanceStrings.push(instStrings);
          } else {
            const strings = instStrings as string[];
            strings.forEach((string) => {
              instanceStrings.push(string);
            });
          }
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
    setReqJSON(newSessionRequest(submission.value));
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
            {runDevices.map(({ id, name }, index) => {
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
                </div>
              );
            })}
            <JSONExport name={"run.json"} json={reqJSON}>
              <button type={"submit"}>
                New session
              </button>
            </JSONExport>

          </Form>
        }
      </Suspense>
    </div>
  );
};