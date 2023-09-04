import { createRef, forwardRef, useState } from "react";
import type { ChangeEventHandler, FC, Ref, RefObject } from "react";
import { PlusCircleIcon, XCircleIcon } from "@heroicons/react/24/outline";
import type { z } from "zod";
import type { FieldConfig } from "@conform-to/react";
import { list } from "@conform-to/react";

type FormProps = {
  noButton?: boolean;
  activeButton?: boolean;
  fields: {
    name: string;
    content?: string | string[] | number | boolean | object[];
    type: "text" | "textarea" | "select" | "checkbox" | "radio" | "number" | "multiselect" | "array";
    value?: ({ key: string } & FieldConfig<string>)[]
    options?: { value: string, display: string }[];
    arraySchema?: z.Schema
    arrayFields?: ({ key: string } & FieldConfig<{ type: string, name: string, description?: string | undefined }>)[]
    error?: string;
  }[]
}

type ArrayFieldProps = {
  field: FieldConfig<any>
  arrayFields: ({ key: string } & FieldConfig<{ type: string, name: string, description?: string | undefined }>)[]
  arraySchema: z.Schema,
  onAddItem?: (item: Record<string, any>) => void
  onRemoveItem?: (index: number) => void
}

const ArrayField: FC<ArrayFieldProps> = forwardRef(({
                                                      field,
                                                      arrayFields
                                                    }: ArrayFieldProps, ref: Ref<HTMLFieldSetElement>) => {
  const refs: Array<RefObject<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>> = arrayFields.map(() => createRef());
  const [arrayFieldValues, setArrayFieldValues] = useState<string[]>(
    arrayFields.map(() => "")
  );

  return (
    <fieldset ref={ref as Ref<HTMLFieldSetElement>} className={"flex flex-col h-full space-y-2 p-2"}>
      <table className={"table-auto text-sm items-start"}>
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
        {arrayFields.map((item, index) => (
          <tr key={index}>
            {}
            <td>
              {item.form}
            </td>
            <td>
              {item.form}
            </td>
            <td>
              {item.form}
            </td>
            <td>
              <button type={"button"} {...list.remove(field.name, { index })}><XCircleIcon className={"h-7 w-7"} />
              </button>
            </td>
          </tr>
        ))}
        <tr>
          {/* Iterate through arrayFields and create a form */}
          {arrayFields.map((field, index) => (
            // create a ref in the arrayFieldRefs object for each field
            <td
              key={index}
              className={"align-top "}
            >
              <FieldComponent
                field={{ ...field, type: "text" }}
                inputRef={refs[index]}
                value={arrayFieldValues[index]}
                onChange={(e) => {
                  // update the value in the arrayFieldValues array
                  const newValues = [...arrayFieldValues];
                  newValues[index] = e.target.value;
                  setArrayFieldValues(newValues);
                }}
              />
            </td>
          ))}
          <td
            className={"align-top"}
          >
            <button
              type={"button"}
              {...list.append(field.name)}
            ><PlusCircleIcon className={"h-7 w-7"} /></button>
          </td>
        </tr>
        </tbody>
      </table>
    </fieldset>
  );
});

ArrayField.displayName = "ArrayField";

type FieldComponentProps = {
  field: FormProps["fields"][0]
  value?: string | number | readonly string[]
  onChange?: ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  inputRef?: RefObject<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement | HTMLFieldSetElement>
}

const FieldComponent = forwardRef((props: FieldComponentProps, ref: Ref<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
  const [items, setItems] = useState([] as Record<string, any>[]);
  const handleRemoveItem = (index: number) => {
    // Remove item logic goes here.
    // Delete the item at the given index
    const newItems = [...items];
    newItems.splice(index, 1);
    setItems(newItems);
  };

  const handleAddItem = (item: Object) => {
    setItems([...items, item]);
  };

  const { field, value, onChange, inputRef } = props;
  const baseProps = {
    name: field.name,
    "aria-invalid": Boolean(field.error),
    "aria-errormessage": field.error ? `${field.name}-error` : undefined,
    className: "dark:bg-slate-700 rounded-full px-2 py-2",
    value: value,
    onChange: onChange
  };

  switch (field.type) {
    case "textarea":
      baseProps.className = "dark:bg-slate-700 rounded-xl px-2 py-2";
      return <textarea defaultValue={field.content as string} {...baseProps}
                       ref={inputRef && inputRef as Ref<HTMLTextAreaElement>} />;
    case "select":
      return (
        <select {...baseProps}
                ref={inputRef && inputRef as Ref<HTMLSelectElement>}
        >
          <option value={""} disabled>Select...</option>
          {field.options?.map(({ value, display }, i) => (
            <option key={i} value={value}>
              {display}
            </option>
          ))}
        </select>
      );
    case "checkbox":
      return <input  type={field.type} {...baseProps} ref={
        inputRef && inputRef as Ref<HTMLInputElement>
      } />;
    case "radio":
      return <input type={field.type} {...baseProps} ref={
        inputRef && inputRef as Ref<HTMLInputElement>} />;
    case "number":
      return <input defaultValue={field.content as number} type={field.type} {...baseProps} ref={
        inputRef && inputRef as Ref<HTMLInputElement>} />;
    case "array":
      return <ArrayField
        field={field}
        arrayFields={field.arrayFields!}
        arraySchema={field.arraySchema!}
        onAddItem={handleAddItem}
        onRemoveItem={handleRemoveItem}
      />;
    case "multiselect":
      baseProps.className = "dark:bg-slate-700 rounded-xl px-2 py-2";
      return (
        <fieldset ref={inputRef && inputRef as Ref<HTMLFieldSetElement>}>
          <select defaultValue={field.content as string[]} multiple {...baseProps} >
            <option disabled>Select an option</option>
            {field.options?.map(({ value, display }, i) => (
              <option key={i} value={value}>
                {display}
              </option>
            ))}
          </select>
        </fieldset>

      );
    default:
      return <input defaultValue={field.content as string} type={field.type} {...baseProps} />;
  }
});

FieldComponent.displayName = "FieldComponent";

const FormContent = (props: FormProps) => {
  return (
    <div className={`flex flex-col space-y-2 p-2`}>
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
