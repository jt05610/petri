import { RefObject, useEffect, useRef, useState } from "react";
import type { FC } from "react";
import { PlusCircleIcon, XCircleIcon } from "@heroicons/react/24/outline";
import type { z } from "zod";
import { parse } from "@conform-to/zod";

type FormProps = {
  noButton?: boolean;
  activeButton?: boolean;
  fields: {
    name: string;
    content?: string | string[] | number | boolean | object[];
    type: "text" | "textarea" | "select" | "checkbox" | "radio" | "number" | "multiselect" | "array";
    options?: { value: string, display: string }[];
    arraySchema?: z.Schema
    arrayFields?: FormProps["fields"]
    error?: string;
  }[]
}

type ArrayFieldProps = {
  field: FormProps["fields"][0]
  arrayFields: FormProps["fields"]
  arraySchema: z.Schema
}

const ArrayField: FC<ArrayFieldProps> = ({ field, arraySchema, arrayFields }: ArrayFieldProps) => {
  const [items, setItems] = useState(field.content as object[]);

  const handleRemoveItem = (index: number) => {
    // Remove item logic goes here.
    // Delete the item at the given index
    const newItems = [...items];
    newItems.splice(index, 1);
    setItems(newItems);
  };

  const handleAddItem = (item: object) => {
    // Add item logic goes here.
    // Add the item to the items array
    setItems([...items, item]);
  }

  return (
    <div className={"flex flex-col h-full space-y-2 p-2"}>
      <table className={"table-auto"}>
        {/* Table logic to display items */}
        <thead>
        <tr>
          {arrayFields.map((field, index) => (
            <th key={index}>{field.name}</th>
          ))}
          <th>Actions</th>
        </tr>
        </thead>
        <tbody>
        {items && items.map((item, index) => (
          <tr key={index}>
            {arrayFields.map((field, index) => (
              <td key={index}>
                <FieldComponent field={field} />
              </td>
            ))}
            <td>
              <button onClick={() => handleRemoveItem(index)}><XCircleIcon className={"h-5 w-5"} /></button>
            </td>
          </tr>
        ))}
        </tbody>
      </table>
      <form onSubmit={(e) => {
        e.preventDefault();
        // parse the form data with the array schema
        const formData = new FormData(e.target as HTMLFormElement);
        const parsed = parse(formData, { schema: arraySchema });
        if (parsed.intent === "submit") {
          // Add the item to the items array
          setItems([...items, parsed.value]);
        }
      }}>
        {/* Iterate through arrayFields and create a form */}
        {arrayFields.map((field, index) => (
          // create a ref in the arrayFieldRefs object for each field
          <FieldComponent key={index} field={field}/>
        ))}
        <button
          type={"button"}
          onClick={() => handleAddItem}
        ><PlusCircleIcon className={"h-5 w-5"}/></button>
      </form>
    </div>
  );
};

const FieldComponent = ({ field, ref }: { field: FormProps["fields"][0], ref?: RefObject<any> }) => {
  const baseProps = {
    name: field.name,
    "aria-invalid": Boolean(field.error),
    "aria-errormessage": field.error ? `${field.name}-error` : undefined,
    className: "dark:bg-slate-700 rounded-full px-2 py-2",
    ref,
  };

  switch (field.type) {
    case "textarea":
      baseProps.className = "dark:bg-slate-700 rounded-xl px-2 py-2";
      return <textarea defaultValue={field.content as string} {...baseProps}/>;
    case "select":
      return (
        <select defaultValue={field.content as string} {...baseProps}>
          {field.options?.map(({ value, display }, i) => (
            <option key={i} value={value}>
              {display}
            </option>
          ))}
        </select>
      );
    case "checkbox":
      return <input defaultChecked={field.content === "true"} type={field.type} {...baseProps} />;
    case "radio":
      return <input defaultChecked={field.content === "true"} type={field.type} {...baseProps} />;
    case "number":
      return <input defaultValue={field.content as number} type={field.type} {...baseProps} />;
    case "array":
      return <ArrayField field={field} arrayFields={field.arrayFields!} arraySchema={field.arraySchema!}/>;
    case "multiselect":
      return (
        <select defaultValue={field.content as string[]} multiple {...baseProps}>
          {field.options?.map(({ value, display }, i) => (
            <option key={i} value={value}>
              {display}
            </option>
          ))}
        </select>
      );
    default:
      return <input defaultValue={field.content as string} type={field.type} {...baseProps} />;
  }
};

const FormContent = (props: FormProps) => {
  return (
    <div className={`flex flex-col space-y-2`}>
      {props.fields.map((field, i) => (
        <div
          key={i}
          className={`flex flex-col space-y-1 form-field ${field.type === "checkbox" ? "form-field-checkbox" : ""}`}
        >
          <div className={`flex flex-row justify-between`}>
            <label htmlFor={field.name}>
              {field.name}:{" "}
            </label>
            {field.error ? (
              <p
                className="form-validation-error text-amber-600"
                id={`${field.name}-error`}
                role="alert"
              >
                {field.error}
              </p>
            ) : null}
          </div>
          <FieldComponent field={field} />
        </div>
      ))}
      {props.noButton ? null : (
        <button
          type="submit"
          className={`rounded-full text-white py-2 px-4 bg-teal-700 hover:bg-teal-600 ${(props.activeButton !== undefined && !props.activeButton) ? "disabled cursor-not-allowed" : ""}`}
        >Submit
        </button>
      )}
    </div>
  );
};

export default FormContent;
