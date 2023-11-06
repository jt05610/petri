import type { FieldConfig } from "@conform-to/react";
import { conform, list, useFieldList, useFieldset } from "@conform-to/react";
import React, { useRef } from "react";
import { MinusCircleIcon } from "@heroicons/react/24/outline";

export function splitCamelCase(str: string) {
  // if all uppercase then return
  if (str === str.toUpperCase()) {
    return str;
  }
  return str.replace(/([A-Z])/g, " $1");
}

export function splitSnakeCase(str?: string) {
  if (!str) {
    return "";
  }
  return str.replace(/_/g, " ");
}

export function split(str: string) {
  return splitCamelCase(splitSnakeCase(str));
}


export function toSentenceCase(str: string) {
  const splitStr = split(str);
  return splitStr.charAt(0).toUpperCase() + splitStr.slice(1).toLowerCase();
}

export function FieldLabel<T>(cfg: FieldConfig<T>) {
  return (
    <label
      htmlFor={cfg.name}
      className={"block text-sm font-medium text-gray-700 dark:text-gray-300"}
    >
      {toSentenceCase(cfg.name)}
    </label>
  );
}

export function FieldInput<T>(config: FieldConfig<T>) {
  return (
    <input
      name={config.name}
      className={"block rounded-full w-full px-3 py-2 border border-gray-300 shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:bg-slate-800 dark:border-slate-400 dark:text-gray-300"}
      defaultValue={config.defaultValue?.toString()}
    />
  );
}

interface FieldSelectInputProps<T> extends FieldConfig<T> {
  options: string[];
}

export function FieldSelectInput<T>(config: FieldSelectInputProps<T>) {
  return (
    <select
      {...conform.select(config)}
      className={"block rounded-full w-full px-3 py-2 border border-gray-300 shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:bg-slate-800 dark:border-slate-400 dark:text-gray-300"}
      defaultValue={config.defaultValue?.toString()}
    >
      {config.options.map((option, index) => (
        <option key={index} value={option}>
          {toSentenceCase(option)}
        </option>
      ))}
    </select>
  );
}

export function SelectField<T>(config: FieldConfig<T> & {
  options: string[]
}) {
  return (
    <>
      <FieldLabel {...config} />
      <FieldSelectInput {...config} />
      <FieldError {...config} />
    </>
  );
}

export function FieldError<T>(config: FieldConfig<T>) {
  return (
    <div
      className={"block text-sm font-italic font-medium text-red-700 dark:text-red-300"}
    >
      {config.error}
    </div>
  );
}

export function Field<T>(config: FieldConfig<T>) {
  return (
    <div className={"p-2"}>
      <FieldLabel {...config} />
      <FieldInput {...config} />
      <FieldError {...config} />
    </div>
  );
}

export function FieldTextArea<T>(config: FieldConfig<T>) {
  return (
    <div className={"p-2"}>
      <FieldLabel {...config} />
      <textarea
        name={config.name}
        className={"block rounded-lg w-full px-3 py-2 border border-gray-300 shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:bg-slate-800 dark:border-slate-400 dark:text-gray-300"}
        defaultValue={config.defaultValue?.toString()}
      />
      <FieldError {...config} />
    </div>
  );
}

export function FieldSet<T>(config: FieldConfig<T> & {
  fields: string[]
}) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const fieldset = useFieldset(ref, config);
  return (
    <fieldset ref={ref}>
      <legend>{config.name}</legend>
      {config.fields.map((field, index) => (
        <div
          key={index}
          className={"flex flex-col space-y-2"}
        >
          <Field {...fieldset[field]} />
        </div>
      ))}
    </fieldset>
  );
}

export function FieldSetRow<T>({ config, fields, parent, idx }: {
  config: FieldConfig<T>,
  fields: (string | { name: string, options: string[] })[],
  parent: FieldConfig<T[]>,
  idx: number
}) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const fieldset = useFieldset(ref, config);
  return (
    <tr>
      {fields.map((field, index) => (
        <td
          key={index}
          className={"border border-slate-600 px-3 py-2"}
        >
          {(typeof field === "string") ?
            <div>
              <input
                name={fieldset[field].name}
                className={"block bg-slate-300 dark:bg-slate-800 rounded-md w-full px-3 py-2 shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:text-gray-300"}
              />
              <FieldError {...fieldset[field]} />
            </div> : (
              <FieldSelectInput {...field} />
            )}
        </td>
      ))}
      <td>
        <button {...list.remove(parent.name, { index: idx })}>
          <MinusCircleIcon className={"h-6 w-6 text-red-400"} />
        </button>
        <FieldError {...parent} />
      </td>
    </tr>
  );
}

export function FieldList(config: FieldConfig<string[] | undefined>) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const items = useFieldList(ref, config);
  return (
    <fieldset ref={ref}>
      <legend>{toSentenceCase(config.name)}</legend>
      {items.map((item, index) => (
          <div key={item.key} className={"flex flex-row align-middle justify-between"}>
            <FieldInput {...item} />
            <button {...list.remove(config.name, { index })}>
              <MinusCircleIcon className={"h-6 w-6 text-red-400"} />
            </button>
            <FieldError {...item} />
          </div>
        )
      )}
      <button {...list.append(config.name, { defaultValue: "" })}>
        Add
      </button>
    </fieldset>
  );
}

export function FieldSetList<T>({ config, fields, fieldList, defaultValue }: {
  config: FieldConfig<T[]>,
  fields: string[],
  fieldList: ({
    key: string
  } & FieldConfig<T>)[],
  defaultValue?: T
}) {
  return (
    <ul>
      {fieldList.map((cfg, index) => (
          <li key={cfg.key} className={"flex flex-row align-middle justify-between"}>
            <FieldSet {...cfg} fields={fields} />
            <button {...list.remove(config.name, { index })}>
              <MinusCircleIcon className={"h-6 w-6 text-red-400"} />
            </button>
          </li>
        )
      )}
      <button {...list.append(config.name, { defaultValue })}>
        Add
      </button>
    </ul>
  );
}

export function FieldSetTable<T>({ config, fields, fieldList, defaultValue }: {
  config: FieldConfig<T[]>,
  fields: (string | { name: string, options: string[] })[],
  fieldList: ({
    key: string
  } & FieldConfig<T>)[],
  defaultValue?: T
}) {
  return (
    <div>
      <table className={"rounded-lg table-auto table border-collapse"}>
        <thead>
        <tr>
          {fields.map((field, index) => (
              <th
                key={index}
                className={"border border-slate-600 px-3 py-2"}
              >
                {typeof field === "string" ? toSentenceCase(field) : toSentenceCase(field.name)}</th>
            )
          )}
        </tr>
        </thead>
        <tbody>
        {fieldList.map((cfg, index) => (
            <FieldSetRow key={cfg.key} config={cfg} fields={fields} idx={index} parent={config} />
          )
        )}
        </tbody>
      </table>
      <button {...list.append(config.name, { defaultValue })}>
        Add
      </button>
    </div>
  );
}
