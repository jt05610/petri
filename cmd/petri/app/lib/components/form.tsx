import {
  materialRenderers,
  materialCells
} from "@jsonforms/material-renderers";

import type { JsonSchema, UISchemaElement } from "@jsonforms/core";
import { JsonForms } from "@jsonforms/react";
import { useSubmit } from "@remix-run/react";
import type { HTMLFormMethod } from "@remix-run/router";
import { useContext } from "react";
import { FormContext, FormSetterContext } from "~/context/form";

type Props = {
  netID: string
  schema: JsonSchema,
  ui?: UISchemaElement
  method: HTMLFormMethod
  buttons: {
    label: string
    route?: string
    color?: string
  }[]
}

export function Form(props: Props) {
  const data = useContext(FormContext);
  const setter = useContext(FormSetterContext);
  const submit = useSubmit();

  function onSubmit(route?: string) {
    if (!route) {
      route = "";
    }
    submit(data, { method: props.method, action: route, encType: "application/json" });
  }

  return (
    <div>
      <JsonForms
        schema={props.schema}
        data={data}
        uischema={props.ui}
        renderers={materialRenderers}
        cells={materialCells}
        onChange={({ data }) => setter!(data)}
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
