import type { MutableRefObject, ReactNode } from "react";
import { Suspense } from "react";
import type { ConstantInput, RunDetails, RunInputDisplay } from "~/models/net.run.server";
import { useContextSelector } from "use-context-selector";
import { RecordRunContext, RunActionType } from "~/context";
import { BackspaceIcon } from "@heroicons/react/24/outline";
import { XCircleIcon } from "@heroicons/react/24/solid";
import { PetriNetActionType, PetriNetContext } from "~/lib/context/petrinet";
import type { ParameterWithValue } from "~/lib/context/session";
import { RunSessionActionType, RunSessionContext } from "~/lib/context/session";

type GridHeaderItemProps = {
  col: number
}

function GridHeaderItem({ col }: GridHeaderItemProps) {
  return (
    <div
      className={`
      px-2 
      row-start-1 
      col-start-${col} ${col == 1 ? "text-right max-w-xs" : "max-w-md"}
      sticky top-0 z-10 min-w-xs bg-white dark:bg-gradient-to-b dark:from-slate-600 dark:to-slate-700 border-slate-100 dark:border-black/10 bg-clip-padding text-slate-900 dark:text-slate-200 border-b text-sm font-medium py-2`}
    >
      {(col > 1 ? `${col - 1}` : "Step")}
    </div>
  );
}

type GridCellProps = {
  children?: ReactNode
}

function GridCell({ children }: GridCellProps) {
  return (
    <div
      className={`border-slate-100 dark:border-slate-200/5 border-b border-r min-w-xs max-w-md`}
    >
      {children}
    </div>
  );

}

type GridRowProps = {
  children: ReactNode
  row: number
}

function GridHeaderCell({ children, row }: GridRowProps) {
  return (
    <div
      className={`col-start-1 max-w-xs border-slate-100 dark:border-slate-200/5 border-r text-xs p-1.5 text-right text-slate-400 uppercase sticky left-0 bg-white dark:bg-slate-800 font-medium`}>
      {children}
    </div>
  );
}

type ActionParamViewProps = {
  fields: {
    [key: string]: string
  }
  constants: ConstantInput[]
  parameters?: Record<string, ParameterWithValue>
  paramRef: MutableRefObject<Record<string, Record<number, Record<string, ParameterWithValue>>>>
}

function DeviceEditParamView({ constants, fields, parameters, paramRef }: ActionParamViewProps) {
  const dispatch = useContextSelector(RunSessionContext, v => v?.dispatch);
  return (
    <table className="table-auto w-full">
      <thead>
      <tr className={"text-xs font-medium"}>
        <th>Name</th>
        <th>Value</th>
      </tr>
      </thead>
      <tbody>
      {constants.map((constant, i) => {
        return (
          <tr key={i} className={"text-xs font-medium"}>
            <td>{fields[constant.fieldID]}</td>
            <td>{constant.value}</td>
          </tr>
        );
      })}
      {parameters && dispatch && Object.values(parameters).map((param, i) => {
          console.log("param:", param);
          return (
            <tr key={i} className={"text-xs font-medium"}>
              <td>{param.parameter.fieldName}</td>
              <td>
                {
                  param.parameter.fieldType === "boolean" ?
                    <input
                      type={"checkbox"}
                      value={param.value ? "true" : "false"}
                      onChange={(e) => {
                        paramRef.current[param.parameter.deviceID][param.parameter.order][param.parameter.fieldID].value = e.target.value;
                        dispatch({
                          type: RunSessionActionType.ChangeParameter,
                          payload: {
                            parameter: param.parameter,
                            value: e.target.value
                          }
                        });
                      }
                      }
                    /> :
                    <input
                      className={"w-full bg-gray-900/20 rounded-full pt-0.5 pb-0.5 px-2"}
                      type={"text"}
                      value={param.value as string}
                      onChange={(e) => {
                        dispatch({
                          type: RunSessionActionType.ChangeParameter,
                          payload: {
                            parameter: param.parameter,
                            value: e.target.value
                          }
                        });
                        paramRef.current[param.parameter.deviceID][param.parameter.order][param.parameter.fieldID].value = e.target.value;
                      }
                      }
                    />
                }
              </td>
            </tr>
          );
        }
      )}

      </tbody>
    </table>
  );
}

type ActionConstantViewProps = {
  actionIndex: number
  fields: {
    [key: string]: string
  }
  constants: ConstantInput[]
}

function DeviceEditConstantView({ actionIndex, constants, fields }: ActionConstantViewProps) {
  const dispatch = useContextSelector(RecordRunContext, v => v?.dispatch);
  return (
    <table className="table-auto">
      <thead>
      <tr className={"text-xs font-medium"}>
        <th>Name</th>
        <th>Value</th>
      </tr>
      </thead>
      <tbody>
      {constants.map((constant, i) => {
        return (
          <tr key={i} className={"text-xs font-medium"}>
            <td>{fields[constant.fieldID]}</td>
            <td>{constant.value}</td>
            <td>
              <button
                onClick={() => {
                  dispatch!({
                    type: RunActionType.ConstantDeleted,
                    payload: {
                      fieldID: constant.fieldID,
                      actionIndex: actionIndex
                    }
                  });
                }
                }
              ><BackspaceIcon className={"h-4 w-4 text-rose-400/50"} /></button>
            </td>
          </tr>
        );
      })}
      </tbody>
    </table>
  );
}

type Color = "teal" | "blue" | "purple" | "green" | "pink" | "yellow" | "red"
const colorFromIndex = (index: number): Color => {
  const colors: Color[] = ["purple", "yellow", "teal", "pink", "green", "blue", "red"];
  return colors[index % colors.length] as Color;
};

type ActionDetailsProps = {
  deviceID: string
  index: number
  color: Color
  name: string,
  fields: {
    [key: string]: string
  }
  constants: ConstantInput[]
  parameters?: Record<string, ParameterWithValue>
  playback?: boolean
  paramRef?: MutableRefObject<Record<string, Record<number, Record<string, ParameterWithValue>>>>
  deleteAction?: () => void
}

function ActionDetails({
                         deviceID,
                         index,
                         color,
                         name,
                         constants,
                         fields,
                         parameters,
                         playback,
                         paramRef,
                         deleteAction
                       }: ActionDetailsProps) {
  const deviceEventColors = {
    teal: "bg-teal-400/20 dark:bg-cyan-600/50 border-teal-700/10 dark:border-cyan-500 text-teal-600 dark:text-cyan-100",
    yellow: "bg-yellow-400/20 dark:bg-amber-600/50 border-yellow-700/10 dark:border-amber-500 text-yellow-600 dark:text-amber-100",
    blue: "bg-blue-400/20 dark:bg-sky-600/50 border-blue-700/10 dark:border-sky-500 text-blue-600 dark:text-sky-100",
    purple: "bg-purple-400/20 dark:bg-violet-600/50 border-purple-700/10 dark:border-violet-500 text-purple-600 dark:text-violet-100",
    pink: "bg-pink-400/20 dark:bg-rose-600/50 border-pink-700/10 dark:border-rose-500 text-pink-600 dark:text-rose-100",
    green: "bg-green-400/20 dark:bg-lime-600/50 border-green-700/10 dark:border-lime-500 text-green-600 dark:text-lime-100",
    red: "bg-red-400/20 dark:bg-rose-600/50 border-red-700/10 dark:border-rose-500 text-red-600 dark:text-rose-100"

  };
  return (
    <div className={`relative m-1 p-2 text-sm rounded-lg flex flex-col max-w-lg ${deviceEventColors[color]}`}>
      <span className="w-full text-sm font-medium">
        <div className="w-full justify-between flex flex-row">
          <div>
          {name}
          </div>
            <div className="text-xs text-gray-500 dark:text-gray-400">
              <button
                onClick={deleteAction}
              >
              <XCircleIcon className={"h-4 w-4 text-rose-400/50"} />
              </button>
          </div>
        </div>

      </span>
      {(constants.length > 0 || parameters) && (
        playback ?
          <DeviceEditParamView fields={fields} constants={constants} parameters={parameters} paramRef={paramRef!} /> :
          <DeviceEditConstantView constants={constants} fields={fields} actionIndex={index} />
      )}
    </div>
  );
}

type RunGridViewProps = {
  nCols: number;
  nRows: number;
  sequence: RunDetails;
  deviceNames: string[];
  deviceMarkings?: {
    [key: string
      ]:
      number;
  },
  paramRef?: MutableRefObject<Record<string, Record<number, Record<string, ParameterWithValue>>>>
}

export default function RunGridView({ nCols, nRows, deviceNames, sequence, paramRef }: RunGridViewProps) {
  const session = useContextSelector(RunSessionContext, (context) => context!.session);
  const petriNet = useContextSelector(PetriNetContext, (context) => context!.petriNet.net);
  return (
    <Suspense fallback={<div>Loading...</div>}>
      {session && (

        <div className="not-prose relative bg-slate-50 rounded-xl overflow-hidden dark:bg-slate-800/25">
          <h2>Progress: {`${session.currentStep} of ${sequence.steps.length}`}</h2>
          <div
            className="absolute inset-0 bg-grid-slate-100 [mask-image:linear-gradient(0deg,#fff,rgba(255,255,255,0.6))] dark:bg-grid-slate-700/25 dark:[mask-image:linear-gradient(0deg,rgba(255,255,255,0.1),rgba(255,255,255,0.5))]"></div>
          <div className="relative rounded-xl overflow-auto">
            <div className="mx-4 bg-white dark:bg-slate-800 shadow-xl overflow-hidden">
              <div
                className={`overflow-scroll grid  grid-cols-${nCols + 1} grid-rows-${nRows + 1}`}>
                {Array.from({ length: nRows + 1 }).map((_, row) => {
                  return Array.from({ length: nCols + 1 }).map((_, col) => {
                      if (row === 0) {
                        return (
                          <GridHeaderItem key={`${row}.${col}`} col={col + 1} />
                        );
                      }
                      if (col === 0) {
                        return (
                          <GridHeaderCell key={`${row}.${col}`} row={row + 1}>
                            {deviceNames[row - 1]}
                          </GridHeaderCell>
                        );
                      }
                      if (sequence.steps[col - 1] !== undefined) {
                        const { action } = sequence.steps[col - 1];
                        // loop through the session markings keys and if the deviceID matches the key of the marking, then the index of the marking is the row we want
                        const intendedRow = Object.keys(session.markings).findIndex((key) => key === action.device.id) + 1;
                        if (row === intendedRow) {
                          const color = colorFromIndex(petriNet!.deviceIndexFromID(action.device.id));
                          return (
                            <GridCell key={`${row}.${col}`}>
                              <ActionDetails
                                paramRef={paramRef}
                                deviceID={action.device.id}
                                index={col - 1}
                                color={color}
                                name={action.event.name}
                                constants={action.constants}
                                parameters={session.parameters && session.parameters[action.device.id] && session.parameters[action.device.id][col - 1] && session.parameters[action.device.id][col - 1]}
                                playback
                                fields={action.event.fields.map((field) => field.id).reduce((acc, id, i) => {
                                    acc[id] = action.event.fields[i].name;
                                    return acc;
                                  }, {} as {
                                    [key: string]: string
                                  }
                                )}
                              />
                            </GridCell>
                          );
                        }
                      }
                      return <GridCell key={`${row}.${col}`} />;
                    }
                  );
                })
                }
              </div>
            </div>
          </div>
          <div
            className="absolute inset-0 pointer-events-none border border-black/5 rounded-xl dark:border-white/5"></div>
        </div>
      )}
    </Suspense>
  );
}

type RecordRunGridViewProps = {
  nCols: number;
  nRows: number;
  sequence: RunInputDisplay;
  deviceNames: string[];
}

export function RecordRunGridView({ nCols, nRows, deviceNames, sequence }: RecordRunGridViewProps) {
  const [petriNet, petriNetDispatch] = useContextSelector(PetriNetContext, (context) => [context!.petriNet.net, context!.dispatch]);
  const dispatch = useContextSelector(RecordRunContext, (context) => context?.dispatch);
  return (
    <div className="not-prose relative bg-slate-50 rounded-xl overflow-hidden dark:bg-slate-800/25">
      <div
        className="absolute inset-0 bg-grid-slate-100 [mask-image:linear-gradient(0deg,#fff,rgba(255,255,255,0.6))] dark:bg-grid-slate-700/25 dark:[mask-image:linear-gradient(0deg,rgba(255,255,255,0.1),rgba(255,255,255,0.5))]"></div>
      <div className="relative rounded-xl overflow-auto">
        <div className="mx-4 bg-white dark:bg-slate-800 shadow-xl overflow-hidden">
          <div
            className={`overflow-scroll grid`}>
            {Array.from({ length: nRows + 1 }).map((_, row) => {
              return Array.from({ length: nCols + 1 }).map((_, col) => {
                  if (row === 0) {
                    return (
                      <GridHeaderItem key={`${row}.${col}`} col={col + 1} />
                    );
                  }
                  if (col === 0) {
                    return (
                      <GridHeaderCell key={`${row}.${col}`} row={row + 1}>
                        {deviceNames[row - 1]}
                      </GridHeaderCell>
                    );
                  }
                  if (sequence.actions[col - 1] !== undefined) {
                    const event = sequence.actions[col - 1];
                    const intendedRow = petriNet.deviceIndexFromID(event.deviceId) + 1;
                    if (row === intendedRow) {
                      const color = colorFromIndex(petriNet.deviceIndexFromID(event.deviceId));
                      return (
                        <GridCell key={`${row - 1}.${col}`}>
                          <ActionDetails
                            deviceID={event.deviceId}
                            index={col - 1}
                            color={color}
                            name={event.eventName}
                            constants={event.constants}
                            fields={event.eventFields.map((field) => field.id).reduce((acc, id, i) => {
                                acc[id] = event.eventFields[i].name;
                                return acc;
                              }
                              , {} as {
                                [key: string]: string
                              }
                            )
                            }
                            deleteAction={() => {
                              if (!dispatch) {
                                console.log("no dispatch");
                                return;
                              }
                              dispatch({
                                type: RunActionType.ActionRemoved,
                                payload: {
                                  index: col - 1
                                }
                              });
                              petriNetDispatch({
                                  type: PetriNetActionType.RevertMarking,
                                  payload: {
                                    markingIndex: col - 1
                                  }
                                }
                              );
                            }
                            }
                          />
                        </GridCell>
                      );
                    }
                  }
                  return <GridCell key={`${row}.${col}`} />;
                }
              );
            })
            }
          </div>
        </div>
      </div>
      <div className="absolute inset-0 pointer-events-none border border-black/5 rounded-xl dark:border-white/5"></div>
    </div>
  );
}

const rowStart = (row: number) => {
  // have to use switch statement because tailwind doesn't support dynamic classes. supporting 20 rows, then adding scroll if more than 20
  if (row > 20) row = row % 20;
  switch (row) {
    case 1:
      return "row-start-1";
    case 2:
      return "row-start-2";
    case 3:
      return "row-start-3";
    case 4:
      return "row-start-4";
    case 5:
      return "row-start-5";
    case 6:
      return "row-start-6";
    case 7:
      return "row-start-7";
    case 8:
      return "row-start-8";
    case 9:
      return "row-start-9";
    case 10:
      return "row-start-10";
    case 11:
      return "row-start-11";
    case 12:
      return "row-start-12";
    case 13:
      return "row-start-13";
    case 14:
      return "row-start-14";
    case 15:
      return "row-start-15";
    case 16:
      return "row-start-16";
    case 17:
      return "row-start-17";
    case 18:
      return "row-start-18";
    case 19:
      return "row-start-19";
    case 20:
      return "row-start-20";
  }
};
