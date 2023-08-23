import type { ActionArgs } from "@remix-run/node";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { Form, useLoaderData } from "@remix-run/react";
import { useContextSelector } from "use-context-selector";
import { Suspense, useState } from "react";
import { PetriNetContext } from "~/context";
import { useMutation, useQuery } from "@apollo/client";
import { NewSessionDocument, SessionsDocument } from "~/models/__generated__/graphql";
import type { RunDetails, ConstantDetails, ActionDetails } from "~/models/net.run.server";
import { getRunDetails } from "~/models/net.run.server";

type Devices = {
  id: string
  instanceID: string
}[]

export const loader = async ({ params, request }: ActionArgs) => {
  await requireUser(request);
  const sequenceID = params.sequenceID;
  invariant(sequenceID, "sequenceID is required");
  const run = await getRunDetails({ runID: sequenceID });
  const runParams = getParameters(run);
  return { run, runParams };
};

type EventConstant = {
  event: ActionDetails["event"]
  constants: ConstantDetails[]
}

type Parameter = {
  eventID: string
  fieldID: string
  fieldName: string
}


function getParameters(run: RunDetails): Parameter[] {
  // make all EventConstants from the run
  const eventConstants: EventConstant[] = run.steps.map(({ action }) => {
    return {
      event: action.event!,
      constants: action.constants
    };
  });
  const params: Parameter[] = [];
  eventConstants.forEach(({ event, constants }) => {
    event.fields.forEach((field) => {
      const constant = constants.find((constant) => constant.fieldID === field.id);
      if (constant) return;
      params.push({
        eventID: event.id,
        fieldID: field.id,
        fieldName: field.name
      });
    });
  });
  return params;
}

export default function PlaySequence() {
  const { run, runParams } = useLoaderData<typeof loader>();
  const net = useContextSelector(PetriNetContext, (ctx) => ctx?.petriNet);
  const marking = useContextSelector(PetriNetContext, (ctx) => ctx?.marking);
  const [devices, setDevices] = useState<Devices>([]);
  // set inintial parameters with blank values from each field in runParams
  const [parameters, setParameters] = useState<Record<string, any>>(
    runParams.reduce((acc, { eventID, fieldID, fieldName }) => ({
        ...acc,
        [`${eventID}.${fieldID}`]: ""
      }
    ), {})
  );

  const { loading, data } = useQuery(
    SessionsDocument, { variables: { runID: run.id } }
  );

  const [newSession] = useMutation(
    NewSessionDocument,
    {
      variables: {
        input: {
          sequenceID: run.id,
          instances: devices.map(({ id, instanceID }) => ({ deviceID: id, instanceID })),
          parameters: parameters
        }
      }
    }
  );

  const [selectedInstances, setSelectedInstances] = useState<{
    [deviceID: string]: string
  }>({} as {
    [deviceID: string]: string
  });

  const handleSelectChanged = (deviceID: string, instanceID: string) => {
    setSelectedInstances({ ...selectedInstances, [deviceID]: instanceID });
    setDevices([...devices, { id: deviceID, instanceID }]);
  };

  const handleSubmit = async () => {
    const res = await newSession();
    console.log("newSession", res);
  };

  const handleParameterChanged = (eventID: string, fieldID: string, value: string) => {
    setParameters({ ...parameters, [`${eventID}.${fieldID}`]: value });
  };

  return (
    <div>
      <h3>Active sessions</h3>
      <Suspense fallback={<div>Loading...</div>}>
        {!loading && data && data.sessions.map((session) => {
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
          <Form method={"post"} onSubmit={handleSubmit}>
            {net.childDeviceEvents(marking!).map(({ id, name, instances, events }) => {
              return (
                <div
                  key={id}
                  className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                >
                  <h2 className={"text-xl font-bold"}>{name}</h2>
                  <select
                    defaultValue={""}
                    className={"rounded-full p-2"}
                    onChange={(e) => {
                      const instance = instances.find((instance) => instance.id === e.target.value);
                      if (!instance) return;
                      handleSelectChanged(id, instance.id);
                      const data = { data: {}, deviceID: instance.id, command: "get" };
                      console.log("systemControl sent", data);
                    }
                    }>
                    <option value={""}>Select a device</option>
                    {instances?.map((instance) => {
                        return (
                          <option key={instance.id} value={instance.id}>{instance.name}</option>
                        );
                      }
                    )}
                  </select>
                  {events.map((event) => {
                    return (
                      <div
                        className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                        key={event.id}
                      >
                        <h2 className={"text-xl font-bold"}>{event.name}</h2>

                        {event.fields!.map((field, i) => {
                          return (
                            <div
                              key={i}
                              className={"m-2 flex flex-row space-x-2 items-center"}
                            >
                              <label
                                className={"w-1/4"}
                                htmlFor={field.name}
                              >
                                {field.name}
                              </label>
                              <input
                                className={"w-3/4"}
                                name={field.name}
                                type={field.type}
                                defaultValue={""}
                                value={parameters[`${event.id}.${field.id}`]}
                                onChange={(e) => handleParameterChanged(event.id, field.id, e.target.value)}
                              />
                            </div>
                          );
                        })}
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