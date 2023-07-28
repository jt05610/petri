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
  setData: (data: any) => void
  route: string
  initialData?: any
  onDone: () => void
}

export function Creator(props: Props) {
  const submit = useSubmit();

  function onSubmit() {
    let formData = new FormData();
    if (!props.schema.properties) {
      return;
    }
    for (const [key, value] of Object.entries(props.schema.properties)) {
      if (typeof value !== "object" || !value.type) {
        continue;
      }
      if (value.type === "array") {
        formData.append(key, JSON.stringify(props.data[key]));
      } else {
        formData.append(key, props.data[key]);
      }
    }

    submit(formData, { method: props.method, action: props.route });
    props.onDone();
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
      <button
        className="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 focus:bg-blue-400"
        onClick={onSubmit}
      >
        Create Place
      </button>
    </div>
  );
}
