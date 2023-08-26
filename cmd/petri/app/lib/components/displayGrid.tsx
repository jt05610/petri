import type { ReactNode } from "react";
import type { ConstantInput, RunDetails, RunInputDisplay } from "~/models/net.run.server";
import { useContextSelector } from "use-context-selector";
import { RecordRunContext, RunActionType, RunSessionContext } from "~/context";
import { BackspaceIcon } from "@heroicons/react/24/outline";
import { PetriNetContext } from "~/lib/context/petrinet";

type GridHeaderItemProps = {
  col: number
}

function GridHeaderItem({ col }: GridHeaderItemProps) {
  return (
    <div
      className={`px-2 row-start-1 col-start-${col} ${col == 0 && "text-right"} sticky top-0 z-10 bg-white dark:bg-gradient-to-b dark:from-slate-600 dark:to-slate-700 border-slate-100 dark:border-black/10 bg-clip-padding text-slate-900 dark:text-slate-200 border-b text-sm font-medium py-2`}
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
      className={`border-slate-100 dark:border-slate-200/5 border-b border-r`}
    >
      {children}
    </div>
  );

}

type GridRowProps = {
  title?: string
  row: number
}

function GridHeaderCell({ title, row }: GridRowProps) {
  return (
    <div
      className={`row-start-${row} col-start-1 border-slate-100 dark:border-slate-200/5 border-r text-xs p-1.5 text-right text-slate-400 uppercase sticky left-0 bg-white dark:bg-slate-800 font-medium`}>
      {title || " "}
    </div>
  );
}

type ActionConstantViewProps = {
  actionIndex: number
  fields: {
    [key: string]: string
  }
  constants: ConstantInput[]
}

function DeviceParameterView({ actionIndex, constants, fields }: ActionConstantViewProps) {
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
                  if (confirm("Are you sure you want to delete this constant?")) {
                    dispatch!({
                      type: RunActionType.ConstantDeleted,
                      payload: {
                        fieldID: constant.fieldID,
                        actionIndex: actionIndex
                      }
                    });
                  }
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
  index: number
  color: Color
  name: string,
  fields: {
    [key: string]: string
  }
  constants: ConstantInput[]
}

function ActionDetails({ index, color, name, constants, fields }: ActionDetailsProps) {
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
    <div className={`m-1 p-2 text-sm rounded-lg flex flex-col ${deviceEventColors[color]}`}>
      <span className="text-sm font-medium">{name}</span>
      {constants.length > 0 && (
        <DeviceParameterView constants={constants} fields={fields} actionIndex={index} />
      )}
    </div>
  );
}


type RunGridViewProps = {
  nCols: number
  nRows: number
  sequence: RunDetails
  deviceNames: string[]
}

export default function RunGridView({ nCols, nRows, deviceNames, sequence }: RunGridViewProps) {

  const session = useContextSelector(RunSessionContext, (context) => context!.session);

  const petriNet = useContextSelector(PetriNetContext, (context) => context!.petriNet.net);
  return (
    <div className="not-prose relative bg-slate-50 rounded-xl overflow-hidden dark:bg-slate-800/25">
      <h2>Progress: {`${session.activeIndex} of ${sequence.steps.length}`}</h2>
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
                    return <GridHeaderCell key={`${row}.${col}`} title={deviceNames[row - 1]} row={row + 1} />;
                  }
                  if (sequence.steps[col - 1] !== undefined) {
                    const { action } = sequence.steps[col - 1];
                    const intendedRow = petriNet!.deviceIndexFromID(action.device.id) + 1;
                    if (row === intendedRow) {
                      const color = colorFromIndex(petriNet!.deviceIndexFromID(action.device.id));
                      return (
                        <GridCell key={`${row}.${col}`}>
                          <ActionDetails
                            index={col - 1}
                            color={color}
                            name={action.event.name}
                            constants={action.constants}
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
      <div className="absolute inset-0 pointer-events-none border border-black/5 rounded-xl dark:border-white/5"></div>
    </div>
  );
}

type RecordRunGridViewProps = {
  nCols: number
  nRows: number
  sequence: RunInputDisplay
  deviceNames: string[]
}

export function RecordRunGridView({ nCols, nRows, deviceNames, sequence }: RecordRunGridViewProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context!.petriNet.net);
  return (
    <div className="not-prose relative bg-slate-50 rounded-xl overflow-hidden dark:bg-slate-800/25">
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
                    return <GridHeaderCell key={`${row}.${col}`} title={deviceNames[row - 1]} row={row + 1} />;
                  }
                  if (sequence.actions[col - 1] !== undefined) {
                    const event = sequence.actions[col - 1];
                    const intendedRow = petriNet.deviceIndexFromID(event.deviceId) + 1;
                    if (row === intendedRow) {
                      const color = colorFromIndex(petriNet.deviceIndexFromID(event.deviceId));
                      return (
                        <GridCell key={`${row}.${col}`}>
                          <ActionDetails
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
