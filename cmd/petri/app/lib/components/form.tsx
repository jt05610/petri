import {
  materialRenderers,
  materialCells
} from "@jsonforms/material-renderers";

import type { JsonSchema, UISchemaElement } from "@jsonforms/core";
import { JsonForms } from "@jsonforms/react";
import { useSubmit } from "@remix-run/react";
import type { HTMLFormMethod } from "@remix-run/router";

type Props = {
  netID: string
  schema: JsonSchema,
  ui?: UISchemaElement
  method: HTMLFormMethod
  data: any
  buttons: {
    label: string
    route?: string
    color?: string
  }[]

  setData: (data: any) => void
  initialData?: any
}

export function Form(props: Props) {
  const submit = useSubmit();

  function onSubmit(route?: string) {
    if (!route) {
      route = "";
    }
    submit(props.data, { method: props.method, action: route, encType: "application/json" });
  }

  return (
    <div>
      <JsonForms
        schema={props.schema}
        data={props.data}
        uischema={props.ui}
        renderers={materialRenderers}
        cells={materialCells}
        onChange={({ data }) => props.setData(data)}
      />
      <div className="flex flex-row justify-between">
        {props.buttons.map(({ label, route, color }, i) => (
          <button
            key={i}
            className={`rounded bg-${color ? color : "blue-500"} px-4 py-2 text-white hover:bg-blue-600 focus:bg-blue-400`}
            onClick={() => {
              onSubmit(route);
            }}
          >
            {label}
          </button>
        ))}
      </div>
    </div>
  );
}
