type FormProps = {
  noButton?: boolean;
  activeButton?: boolean;
  fields: {
    name: string;
    content?: string;
    type: "text" | "textarea" | "select" | "checkbox" | "radio" | "number";
    options?: { value: string, display: string }[];
    error?: string;
  }[]
}

export default function FormContent(props: FormProps) {
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
          {field.type === "textarea" ? (
            <textarea
              defaultValue={field.content}
              name={field.name}
              aria-invalid={Boolean(field.error)}
              aria-errormessage={
                field.error
                  ? `${field.name}-error`
                  : undefined
              }
              className={`rounded-lg px-2`}
            />
          ) : field.type === "select" ? (
            <select
              defaultValue={field.content}
              name={field.name}
              aria-invalid={Boolean(field.error)}
              aria-errormessage={
                field.error
                  ? `${field.name}-error`
                  : undefined
              }
              className={`rounded-full px-2`}
            >
              {field.options?.map(({ value, display }, i) => (
                <option key={i} value={value}>
                  {display}
                </option>
              ))}
            </select>
          ) : field.type === "checkbox" ? (
            <input
              defaultChecked={field.content === "true"}
              name={field.name}
              type={field.type}
              aria-invalid={Boolean(field.error)}
              aria-errormessage={
                field.error
                  ? `${field.name}-error`
                  : undefined
              }
              className={`rounded-lg`}
            />
          ) : field.type === "radio" ? (
            <input
              defaultChecked={field.content === "true"}
              name={field.name}
              type={field.type}
              aria-invalid={Boolean(field.error)}
              aria-errormessage={
                field.error
                  ? `${field.name}-error`
                  : undefined
              }
            />
          ) : field.type === "number" ? (
            <input
              defaultValue={field.content}
              name={field.name}
              type={field.type}
              aria-invalid={Boolean(field.error)}
              aria-errormessage={
                field.error
                  ? `${field.name}-error`
                  : undefined
              }
              className={`rounded-full px-2`}
            />
          ) : (
            <input
              defaultValue={field.content}
              name={field.name}
              type={field.type}
              aria-invalid={Boolean(field.error)}
              aria-errormessage={
                field.error
                  ? `${field.name}-error`
                  : undefined
              }
              className={`rounded-full px-2`}
            />
          )}
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
}
