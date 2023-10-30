import type { ActionArgs, LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { parse } from "@conform-to/zod";
import { useContextSelector } from "use-context-selector";
import { RunContext } from "~/lib/context/run";
import type { RunDetails } from "~/models/net.run.server";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import { Form, useLoaderData, useActionData } from "@remix-run/react";
import JSZip from "jszip";
import { getParameterRecord } from "~/util/parameters";

import type { FieldConfig } from "@conform-to/react";
import { conform, useFieldList, useFieldset, useForm, list } from "@conform-to/react";
import { z } from "zod";
import { useRef, useState } from "react";
import { MinusCircleIcon, PlusCircleIcon } from "@heroicons/react/24/outline";
import { ExperimentDesignResult, ExperimentSample, makeExperiment } from "~/models/experiment.server";
import {
  BatchInput,
  createBatch,
  getSamples,
  SampleDetails,
  SampleInput
} from "~/models/net.run.session.sample.server";
import invariant from "tiny-invariant";
import { createColumnHelper, useReactTable, getCoreRowModel, flexRender } from "@tanstack/react-table";

enum ExperimentType {
  lhs = "lhs",
  fullFactorial = "full",
  boxBehnken = "box",
  plackettBurman = "plackett",
}

const numberAxisSchema = z.object({
  min: z.preprocess((x) => {
    if (typeof x === "string") {
      return parseFloat(x);
    }
    return x;
  }, z.number()),
  max: z.preprocess((x) => {
    if (typeof x === "string") {
      return parseFloat(x);
    }
    return x;
  }, z.number()),
  step: z.preprocess((x) => {
    if (typeof x === "string") {
      return parseFloat(x);
    }
    return x;
  }, z.number())
});

const categoricalAxisSchema = z.array(z.string());

const axisSchema = z.object({
  name: z.string(),
  values: categoricalAxisSchema.optional(),
  axis: numberAxisSchema.optional()
});

const schema = z.object({
  kind: z.nativeEnum(ExperimentType),
  axes: z.array(axisSchema),
  n_samples: z.preprocess((x) => {
      if (typeof x === "string") {
        return parseInt(x);
      }
      return x;
    },
    z.number().int().positive()),
  strength: z.preprocess((x) => {
      if (typeof x === "string") {
        return parseInt(x);
      }
      return x;
    }
    , z.number().int().positive()),
  id_stem: z.string().optional()
});

export const action = async ({ params, request }: ActionArgs) => {
  const authorID = await requireUserId(request);
  const runID = params.sequenceID;
  invariant(runID, "No run ID provided");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const formData = await request.formData();
  const sub = parse(formData, { schema });
  if (sub.intent !== "submit" || !sub.value) {
    return json(sub);
  }
  const exp = await makeExperiment(sub.value);
  const batchName = Date.now().toString();
  await createBatch(authorID, runID, convertBatch(batchName, runID, exp));
  return redirect(".");
};

function getSampleParams(sample: ExperimentSample): Record<string, string | number> {
  const params: Record<string, string | number> = {};
  for (const [key, value] of Object.entries(sample)) {
    if (baseColumns.includes(key)) continue;
    params[key] = value;
  }
  return params;
}

function convertSample(sample: ExperimentSample): SampleInput {
  return {
    name: sample.name,
    group: sample.group,
    params: getSampleParams(sample)
  };
}

function convertBatch(name: string, runID: string, result: ExperimentDesignResult): BatchInput {
  return {
    name: name,
    runID: runID,
    samples: result.map(convertSample)
  };
}

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const runID = params.sequenceID;
  invariant(runID, "No run ID provided");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const samples = await getSamples({
    runID: runID
  });
  console.log(samples);
  return json({ user, samples });
};

interface SamplesTableProps {
  samples: SampleDetails[];
}

const baseColumns = [
  "id",
  "name",
  "group"
];

function extractParameterColumns(sample: SampleDetails) {
  return Object.keys(sample).filter((key) => {
    return !baseColumns.includes(key);
  });
}

function uniqueColumns(samples: SampleDetails[]) {
  const uniqueColumns = new Set<string>();
  samples.forEach((sample) => {
    const columns = extractParameterColumns(sample);
    columns.forEach((column) => {
      uniqueColumns.add(column);
    });
  });
  return Array.from(uniqueColumns);
}


function sentenceCase(str: string) {
  const rg = /(^\w{1}|\.\s*\w{1})/gi;
  return str.replace(rg, (toReplace) => {
    return toReplace.toUpperCase();
  });
}

function SamplesTable({ samples }: SamplesTableProps) {
  const columnHelper = createColumnHelper<SampleDetails>();
  const allColumnNames = baseColumns.concat(uniqueColumns(samples));
  const columns = allColumnNames.map((column) => {
    return columnHelper.accessor(column, {
      header: () => <span>{sentenceCase(column)}</span>,
      cell: info => info.renderValue()
    });
  }).concat();
  const tableOptions = {
    columns: columns,
    data: samples,
    getCoreRowModel: getCoreRowModel()
  };
  const table = useReactTable(tableOptions);
  return (
    <div className="p-2">
      <table>
        <thead>
        {table.getHeaderGroups().map(headerGroup => (
          <tr key={headerGroup.id}>
            {headerGroup.headers.map(header => (
              <th key={header.id}>
                {header.isPlaceholder
                  ? null
                  : flexRender(
                    header.column.columnDef.header,
                    header.getContext()
                  )}
              </th>
            ))}
          </tr>
        ))}
        </thead>
        <tbody>
        {table.getRowModel().rows.map(row => (
          <tr key={row.id}>
            {row.getVisibleCells().map(cell => (
              <td key={cell.id}>
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </td>
            ))}
          </tr>
        ))}
        </tbody>
      </table>
      <div className="h-4" />
    </div>
  );
}

type DeviceInstanceInput = {
  deviceID: string
  instanceID: string
}

function getRunDevices(run: RunDetails): DeviceInstanceInput[] {
  return run.steps.map(({ action }) => {
    return {
      deviceID: action.device.id,
      instanceID: ""
    };
  }).filter((device, index, self) => {
    return self.findIndex((d) => d.deviceID === device.deviceID) === index;
  });
}

function DownloadComponent() {
  const sequence = useContextSelector(RunContext, (context) => context?.run);
  const user = useLoaderData<typeof loader>().user;

  function makeSessionRequest(d: RunDetails) {
    return {
      sequenceID: d.id,
      userID: user.id,
      instances: getRunDevices(d)
    };
  }

  function makeStartRequest(d: RunDetails) {
    return {
      sessionID: "",
      parameters: getParameterRecord(d)
    };
  }

  async function handleDownload() {
    if (!sequence) return;
    let zip = new JSZip();
    const sessionReq = makeSessionRequest(sequence);
    zip.file(".gitignore", "data");
    zip.file("data/start.json", JSON.stringify(makeStartRequest(sequence), null, 2));
    zip.file("data/run.json", JSON.stringify(sessionReq, null, 2));
    const blob = await zip.generateAsync({ type: "blob" });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${sequence.name}.zip`;
    link.click();
  }

  return (
    <button
      type={"button"}
      className={"px-2 py-1 text-white hover:text-sky-500 dark:hover:text-sky-400 bg-gradient-to-b from-sky-700 to-sky-800 rounded-full"}
      onClick={handleDownload}
    >
      Download
    </button>

  );
}

const inputClass = "block rounded-full w-full px-3 py-2 border border-gray-300 shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:bg-slate-800 dark:border-slate-400 dark:text-gray-300";
const labelClass = "block text-sm font-medium text-gray-700 dark:text-gray-300";
const errorClass = "block text-sm font-medium text-red-700 dark:text-red-300";

export default function SequenceIndexPage() {
  const { samples } = useLoaderData<typeof loader>();
  const lastSubmission = useActionData<typeof action>();
  const [form, { id_stem, axes, n_samples, strength, kind }] = useForm({
    lastSubmission,
    shouldValidate: "onBlur",
    onValidate({ formData }) {
      const sub = parse(formData, { schema });
      console.log(sub);
      return sub;
    }
  });
  const axisList = useFieldList(form.ref, axes);
  const [expressions, setExpressions] = useState<string[]>([]);

  return (
    <div>
      <DownloadComponent />
      <Form method="post" {...form.props} className={"w-96"}>
        <div>
          <label className={labelClass} htmlFor={kind.id}>Kind</label>
          <select {...conform.select(kind)} className={inputClass}>
            <option value={ExperimentType.lhs}>Latin Hypercube</option>
            <option value={ExperimentType.fullFactorial}>Full Factorial</option>
            <option value={ExperimentType.plackettBurman}>Plackett Burman</option>
            <option value={ExperimentType.boxBehnken}>Box Behnken</option>
          </select>
          <div className={errorClass}>{kind.error}</div>
        </div>
        <div>
          <label className={labelClass} htmlFor={id_stem.id}>ID Stem</label>
          <input {...conform.input(id_stem, { type: "text" })} className={inputClass} />
          <div className={errorClass}>{id_stem.error}</div>
        </div>
        <div>
          <label htmlFor={n_samples.id}>Samples</label>
          <input className={inputClass} {...conform.input(n_samples, { type: "number" })} />
          <div className={errorClass}>{n_samples.error}</div>
        </div>
        <div>
          <label htmlFor={strength.id}>Strength</label>
          <input className={inputClass} {...conform.input(strength, { type: "number" })} />
          <div className={errorClass}>{strength.error}</div>
        </div>
        <div>
          <div className={"flex items-start space-x-2 align-baseline"}>
            <label htmlFor={axes.name} className={"font-bold text-lg"}>Axes</label>
            <button {...list.append(axes.name)}><PlusCircleIcon className={"h-6 w-6 text-green-400"} /></button>
          </div>
          <div className={"flex flex-col space-y-2"}>
            <ul>

              {axisList.map((item, index) => {
                return (
                  <li key={item.key} className={"flex flex-row space-x-2"}>
                    <AxisFieldset config={item} />
                    <button {...list.remove(axes.name, { index })}><MinusCircleIcon
                      className={"h-6 w-6 text-red-400"} />
                    </button>
                  </li>
                );
              })
              }
            </ul>
          </div>
          <div className={errorClass}>{axes.error}</div>
        </div>
        <button
          type={"submit"}
          className={"m-2 px-2 py-1 text-white hover:text-sky-500 dark:hover:text-sky-400 bg-gradient-to-b from-sky-700 to-sky-800 rounded-full"}>
          Submit
        </button>
      </Form>
      <SamplesTable samples={samples} />
      <p>
        No active session.
      </p>
    </div>
  );
}

function AxisFieldset({ config }: { config: FieldConfig<z.infer<typeof axisSchema>> }) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const options = ["Continuous", "Categorical"];
  const [selectedKind, setSelectedKind] = useState(options[0]);
  const { name, axis, values } = useFieldset(ref, config);
  return (
    <fieldset ref={ref} className={"border-gray-400 flex flex-grow flex-col border-2 p-2 rounded-md"}>
      <div>
        <label className={labelClass}>Kind</label>
        <select className={inputClass} value={selectedKind} onChange={(e) => setSelectedKind(e.target.value)}>
          {options.map((option) => {
            return (
              <option key={option} value={option}>{option}</option>
            );
          })
          }
        </select>
      </div>
      <div>
        <label className={labelClass} htmlFor={name.name}>Name</label>
        <input className={inputClass} name={name.name} />
        <div className={errorClass}>{name.error}</div>
      </div>
      <div>
        {selectedKind === "Continuous" && axis && (
          <NumberAxisFieldSet config={axis} />
        )}
        {selectedKind === "Categorical" && values && values.defaultValue && (
          <CategoricalFieldSet config={values} />
        )}
      </div>
    </fieldset>
  );
}


function NumberAxisFieldSet({ config }: { config: FieldConfig<z.infer<typeof numberAxisSchema>> }) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const { min, max, step } = useFieldset(ref, config);
  return (
    <fieldset ref={ref}>
      <div>
        <label htmlFor={min.name} className={labelClass}>Min</label>
        <input name={min.name} className={inputClass} />
        <div className={errorClass}>{min.error}</div>
      </div>
      <div>
        <label htmlFor={max.name} className={labelClass}>Max</label>
        <input name={max.name} className={inputClass} />
        <div className={errorClass}>{max.error}</div>
      </div>
      <div>
        <label htmlFor={step.name} className={labelClass}>Step</label>
        <input name={step.name} className={inputClass} />
        <div className={errorClass}>{step.error}</div>
      </div>
    </fieldset>
  );
}

function CategoricalFieldSet({ config }: { config: FieldConfig<z.infer<typeof axisSchema>> }) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const { values } = useFieldset(ref, config);
  const categoryList = useFieldList(ref, values);
  return (
    <div>
      <div className={"flex items-start space-x-2 align-baseline"}>
        <label htmlFor={values.name}>Options</label>
        <button {...list.append(values.name)}><PlusCircleIcon className={"h-6 w-6 text-green-400"} /></button>
      </div>
      <div className={"flex flex-col space-y-1"}>
        {categoryList.map((item, index) => {
            return (
              <div key={item.key} className={"flex flex-row align-middle justify-between"}>
                <input
                  name={item.name}
                  className={inputClass}
                />
                <button {...list.remove(values.name, { index })}><MinusCircleIcon className={"h-6 w-6 text-red-400"} />
                </button>
                <div>{item.error}</div>
              </div>
            );
          }
        )}
      </div>
      <div>{values.error}</div>
    </div>
  );
}