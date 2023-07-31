import type { ActionArgs, LoaderArgs, TypedResponse } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import type { ActionParams } from "~/context/action";
import { defaultActions, InputKind } from "~/context/action";
import { Form } from "~/lib/components/form";
import type { FormInput } from "~/lib/components/formInput";
import { useContext, useEffect, useState } from "react";
import { getPlace } from "~/models/place.server";
import { getTransition } from "~/models/transition.server";
import type { NetActionKind } from "~/context/net";
import { useLoaderData } from "@remix-run/react";
import { FormSetterContext } from "~/context/form";

type ActionData = {
  result?: FormInput
  error?: Error
}
const handleAction = async (actionFn: Function, op: NetActionKind, requiredFields: string[], formInputs: any): Promise<ActionData> => {
    for (const field of requiredFields) {
      if (formInputs[field] === undefined) {
        return { error: new Error(`Missing required field: ${field}`) };
      }
    }
    try {
      const res = await actionFn(formInputs);
      return {
        result: res
      };
    } catch (error) {
      let res = {
        errors: {
          body: "An error occurred",
          title: "Internal server error"
        }
      };
      if (error instanceof Error) {
        res.errors.body = error.message;
      }

      // Handle any errors from your actionFn
      return {
        error: new Error(res.errors.body)
      };
    }
  }
;


export const action = async ({ params, request }: ActionArgs) => {
  invariant(params.netID, "netID not found");
  invariant(params.operation, "op not found");
  const [op, entity] = params.operation.split("-");
  let action = defaultActions.find((action) => action.kind === op && action.inputKind === entity);
  invariant(action, "action not found");
  const netID = params.netID;
  let req = await request.json();
  if (typeof req !== "object") {
    req = action.input;
  }
  req["netID"] = netID;
  const result = await handleAction(action.action, params.operation as NetActionKind, action.requiredFields, req);
  if (result.error) {
    return json(result.error, { status: 400 });
  }
  return json(result.result, { status: 200 });
};


type LoaderData = {
  formAction: ActionParams;
  netID: string;
  data: FormInput;
  op: NetActionKind;
}

export const loader = async ({ params, request }: LoaderArgs): Promise<TypedResponse<LoaderData>> => {
  invariant(params.netID, "netID not found");
  const netID = params.netID;
  invariant(params.operation, "op not found");
  const url = new URL(request.url);
  const [op, entity] = params.operation.split("-");
  let action = defaultActions.find((action) => action.kind === op && action.inputKind === entity);
  invariant(action, "action not found");
  const id = url.searchParams.get("id");
  if (id === null || id === undefined || id === "") {
    if (action.input) {
      return json({
        formAction: action,
        netID: netID,
        data: action.input as FormInput,
        op: params.operation as NetActionKind
      });
    }
    return json({
      formAction: action,
      netID: netID,
      data: {} as FormInput,
      op: params.operation as NetActionKind
    });
  }

  if (action.inputKind == InputKind.PLACE) {
    const place = await getPlace({ id: id });
    return json({
      formAction: action,
      netID: netID,
      data: place as FormInput,
      op: params.operation as NetActionKind
    });
  }
  if (action.inputKind == InputKind.TRANSITION) {
    const transition = await getTransition({ id: id });
    return json({
      formAction: action,
      netID: netID,
      data: transition as FormInput,
      op: params.operation as NetActionKind
    });
  }
  return json({
    formAction: action,
    netID: netID,
    data: {} as FormInput,
    op: params.operation as NetActionKind
  });
};

export default function OperationForm() {
  const { formAction, data, netID } = useLoaderData<typeof loader>();
  const setter = useContext(FormSetterContext);
  useEffect(() => {
    if (setter) {
      setter({ data });
    }
  }, [data, setter])

  const formTitle = formAction.inputKind.charAt(0).toUpperCase() + formAction.inputKind.slice(1);
  const [showForm, setShowForm] = useState(true);
  return (
    <div>
      {showForm ? (
        <div>
          <div className="flex justify-end">
            <button
              className="hover:bg-red-700 font-bold py-2 px-4 rounded"
              onClick={() => setShowForm(false)}
            >
              Close
            </button>
          </div>

          <h3 className="text-2xl font-bold">{formTitle}</h3>
          {formAction && (
            <Form
              netID={netID}
              schema={formAction.schema}
              buttons={formAction.buttons}
              method={"post"}
            />
          )}

        </div>
      ) : (
        <div className="flex justify-end">
          <button
            className="hover:bg-green-700 font-bold py-2 px-4 rounded"
            onClick={() => setShowForm(true)}
          >
            Edit Place
          </button>
        </div>
      )}
    </div>
  );
}
