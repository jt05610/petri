import type { ReactNode } from "react";
import { PetriNetContext, RecordSequenceContext, SequenceActionType, SocketContext } from "~/context";
import { useContextSelector } from "use-context-selector";
import type {
  DeviceEventConstantInputDisplay,
  DeviceInputDisplay,
  SequenceInputDisplay
} from "~/models/sequence.server";
import { useEffect, useRef } from "react";


type GridHeaderItemProps = {
  text?: string
  col: number
}

function GridHeaderItem({ text, col }: GridHeaderItemProps) {
  return (
    <div
      className={`row-start-1 col-start-${col} sticky top-0 z-10 bg-white dark:bg-gradient-to-b dark:from-slate-600 dark:to-slate-700 border-slate-100 dark:border-black/10 bg-clip-padding text-slate-900 dark:text-slate-200 border-b text-sm font-medium py-2`}
    >
      {text || (col > 1 ? `${col - 1}` : "Step")}
    </div>
  );
}

type GridCellProps = {
  row: number
  col: number
  children?: ReactNode
}

function GridCell({ row, col, children }: GridCellProps) {
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

type ColorClasses = {
  border: string
  bg: string
  text: string
}

type GridEventProps = {
  children: ReactNode
  colorClasses: ColorClasses
}

function GridEvent({ children, colorClasses }: GridEventProps) {
  const { bg, border, text } = colorClasses;
  return (
    <div
      className={`rounded-lg m-1 ${bg} ${border} ${text}`}
    >
      {children}
    </div>
  );
}


type DeviceEventConstantViewProps = {
  color: string
  deviceEventConstants: DeviceEventConstantInputDisplay[]
}

function DeviceParameterView({ color, deviceEventConstants }: DeviceEventConstantViewProps) {
  return (
    <table className="table-auto">
      <thead>
      <tr>
        <th>Name</th>
        <th>Value</th>
        <th>Constant</th>
      </tr>
      </thead>
      <tbody>
      {deviceEventConstants.map((deviceEventConstant, i) => {
        return (
          <tr key={i}>
            <td>{deviceEventConstant.fieldName}</td>
            <td>{deviceEventConstant.value}</td>
            <td><input
              type="checkbox"
              className={`focus:ring-${color}-500 h-4 w-4 text-${color}-600 border-gray-300 rounded`}
              defaultChecked={deviceEventConstant.constant}
            /></td>
          </tr>
        );
      })}
      </tbody>
    </table>
  );
}

type DeviceEventDetailsProps = {
  color: deviceEventColors
  eventName: string
  constants: DeviceEventConstantInputDisplay[]
}

function DeviceEventDetails({ color, eventName, constants }: DeviceEventDetailsProps) {
  return (
    <div className="flex flex-col">
      <span className="text-sm">{eventName}</span>
      <DeviceParameterView color={color} deviceEventConstants={constants} />
    </div>
  );
}

// rainbow ordered tailwind colors to cycle through for device events, also including pink, teal and fuchsia in the correct places. values do not have the bg- prefix or value suffix so that these can be styled for border and text colors as well
export enum deviceEventColors {
  red = "red",
  orange = "orange",
  amber = "amber",
  yellow = "yellow",
  lime = "lime",
  green = "green",
  emerald = "emerald",
  teal = "teal",
  cyan = "cyan",
  lightBlue = "lightBlue",
  blue = "blue",
  indigo = "indigo",
  violet = "violet",
  purple = "purple",
  fuchsia = "fuchsia",
  pink = "pink",
  rose = "rose",
}


export const DeviceEventColorsArray = Object.values(deviceEventColors);

const makeBorderClass = (color: string) => `border border-${color}-700/10`;

const makeBgClass = (color: string) => `bg-${color}-400/20`;

const makeTextClass = (color: string) => `text-${color}-600`;

const makeClasses = (color: string) => ({
  border: makeBorderClass(color),
  bg: makeBgClass(color),
  text: makeTextClass(color)
});

type SequenceGridViewProps = {
  nCols: number
  nRows: number
  labels: string[]
  sequence: SequenceInputDisplay
  deviceNames: string[]
}

function SequenceGridView({ nCols, nRows, labels, deviceNames, sequence }: SequenceGridViewProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context!.petriNet);
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
                      <GridHeaderItem key={`${row}.${col}`} col={col + 1} text={labels[col]} />
                    );
                  }
                  if (col === 0) {
                    return <GridHeaderCell key={`${row}.${col}`} title={deviceNames[row - 1]} row={row + 1} />;
                  }
                  if (sequence.events[col - 1] !== undefined) {
                    const event = sequence.events[col - 1];
                    const devEvent = event.events.find((e) => petriNet!.deviceIndexFromID(e.deviceId) === row - 1);
                    if (devEvent) {
                      const color = DeviceEventColorsArray[petriNet!.deviceIndexFromID(devEvent.deviceId)];
                      const colorClasses = makeClasses(color);
                      return (
                        <GridEvent key={`${row}.${col}`} colorClasses={colorClasses}>
                          <DeviceEventDetails
                            color={color}
                            eventName={devEvent.eventName}
                            constants={devEvent.constants}
                          />
                        </GridEvent>
                      );
                    }
                  }
                  return <GridCell key={`${row}.${col}`} row={row + 1} col={col + 1} />;
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

type SequenceViewProps = {
  minCols: number
  minRows: number
  deviceNames: string[]
  sequence: SequenceInputDisplay
}

function SequenceView({ minCols, minRows, deviceNames, sequence }: SequenceViewProps) {
  const { events } = sequence;
  let labels: string[] = [];
  if (sequence) {
    labels = events.map((s) => s.name);
  }
  return (
    <div className="mt-4 -mb-3">
      <SequenceGridView nCols={minCols} nRows={minRows} labels={labels} deviceNames={deviceNames} sequence={sequence} />
    </div>
  );
}

export default function Timeline() {
  const socket = useContextSelector(SocketContext, (context) => context);
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const petriNetRef = useRef<typeof petriNet>();
  const dispatch = useContextSelector(RecordSequenceContext, (context) => context?.dispatch);
  const dispatchRef = useRef<typeof dispatch>();
  const sequence = useContextSelector(RecordSequenceContext, (context) => context?.sequence);
  const seqRef = useRef<typeof sequence>();

  useEffect(() => {
    console.log("rendering timeline");
  }, []);
  useEffect(() => {
    if (!dispatch) return;
    dispatchRef.current = dispatch;
  }, [dispatch]);

  useEffect(() => {
    if (!sequence) return;
    seqRef.current = sequence;
  }, [sequence]);

  useEffect(() => {
    if (!petriNet) return;
    petriNetRef.current = petriNet;
  }, [petriNet]);

  useEffect(() => {
    if (!socket) return;
    socket.on("command", (data: {
      deviceID: string,
      event: string,
      message: {
        [fieldNme: string]: string
      }
    }) => {
      console.log("command", data);
      // replace whitespace in event name with underscores and make lowercase
      const transformedEventName = (event: string) => event.replace(/\s/g, "_").toLowerCase();
      const event = petriNetRef.current!.events.find((e) => data.event === transformedEventName(e.name));
      // make human-readable timestamp in format DD MMM YYYY HH:MM:SS
      if (!event) {
        console.log("event not found");
        return;
      }
      // map the message fieldNames to the event's field ids
      const constants: DeviceEventConstantInputDisplay[] = event.fields.map((field) => {
        return {
          fieldID: field.id,
          fieldName: field.name,
          constant: false,
          value: data.message[field.name]
        };
      });

      // any data with the message is assumed to be a constant for this event
      const devEventInput: DeviceInputDisplay = {
        eventName: event.name,
        eventID: event.id,
        deviceId: petriNetRef.current!.instanceOf(data.deviceID),
        constants
      };
      console.log("dispatching", devEventInput);

      dispatchRef.current!({
        type: SequenceActionType.EVENT_ADDED,
        payload: devEventInput
      });
    });
  }, [socket]);

  return (
    <div className={"w-full h-3/10 bottom-0 space-x-2"}>
      <h4>Timeline</h4>
      <button
        type="submit"
        className={`rounded-full px-2 py-1 text-white bg-slate-900`}
        onClick={async (e) => {
          e.preventDefault();
          await fetch("sequences/new", {
            method: "POST",
            headers: {
              "Content-Type": "application/json"
            },
            body: JSON.stringify(sequence)
          }).then((res) => res.json()).then((res) => {
            console.log("response", res);
          });
        }
        }
      >
        Save
      </button>
      {
        sequence && <SequenceView
          minCols={10}
          minRows={2}
          deviceNames={petriNet?.devices.map((d) => d.name) ?? []}
          sequence={sequence}
        />
      }
    </div>
  )
    ;
}