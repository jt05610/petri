import type { ActionArgs, LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { defaultActions, InputKind } from "~/context/action";
import { Form } from "~/lib/components/form";
import type { FormInput } from "~/lib/components/formInput";
import { useContext, useEffect } from "react";
import { getPlace } from "~/models/place.server";
import { getTransition } from "~/models/transition.server";
import type { NetActionKind } from "~/context/net";
import { NetDispatchContext } from "~/context/net";
import { useActionData, useLoaderData, useOutletContext } from "@remix-run/react";

const handleAction = async (actionFn: Function, action: NetActionKind, requiredFields: string[], formInputs: any) => {

  for (const field of requiredFields) {
    if (formInputs[field] === undefined) {
      return {
        result: json({
            errors: {
              body: null,
              title: `${field.charAt(0).toUpperCase() + field.slice(1)} is required`
            }
          },
          { status: 400 }
        )
      };
    }
  }
  try {
    const result = await actionFn(formInputs);

    return {
      result: json(result, { status: 200 })
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
      result: json(res, { status: 500 })
    };
  }
};

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

  const { result } = await handleAction(action.action, params.operation as NetActionKind, action.requiredFields, req);
  return result;
};

export const loader = async ({ params, request }: LoaderArgs) => {
  invariant(params.netID, "netID not found");
  const netID = params.netID;
  invariant(params.operation, "op not found");
  const url = new URL(request.url);
  const [op, entity] = params.operation.split("-");
  let action = defaultActions.find((action) => action.kind === op && action.inputKind === entity);
  invariant(action, "action not found");
  const id = url.searchParams.get("id");
  if (action.input) {
    return json({
      action: action,
      netID: netID,
      data: action.input as FormInput,
      buttons: action.buttons,
      op: params.operation as NetActionKind
    });
  }
  if (!id) {
    return json({
      action: action,
      netID: netID,
      data: {} as FormInput,
      buttons: action.buttons,
      op: params.operation as NetActionKind
    });
  }

  if (action.inputKind == InputKind.PLACE) {
    const place = await getPlace({ id: id });
    return json({
      action: action,
      netID: netID,
      data: place,
      buttons: action.buttons,
      op: params.operation as NetActionKind
    });
  }
  if (action.inputKind == InputKind.TRANSITION) {
    const transition = await getTransition({ id: id });
    return json({
      action: action,
      netID: netID,
      data: transition,
      op: params.operation as NetActionKind,
      buttons: action.buttons
    });
  }
  if (action.inputKind == InputKind.ARC) {
    const arc = await getTransition({ id: id });
    return json({
      action: action,
      netID: netID,
      data: arc,
      op: params.operation as NetActionKind,
      buttons: action.buttons
    });
  }
  return json({
    action: action,
    netID: netID,
    data: {} as FormInput,
    op: params.operation as NetActionKind,
    buttons: action.buttons
  });
};

export default function OperationForm() {
  const [selected, formData, setFormData] = useOutletContext<any>();
  const actionData = useActionData<typeof action>();
  const { action, op, netID, buttons } = useLoaderData<typeof loader>();
  const formTitle = action.inputKind.charAt(0).toUpperCase() + action.inputKind.slice(1);
  const dispatch = useContext(NetDispatchContext);

  useEffect(() => {
    if (!selected) {
      return;
    }
    switch (op) {
      case "add-arc":
        if (selected.kind == "transition") {
          setFormData({
            ...formData,
            transitionID: selected.id
          });
        }
        if (selected.kind == "place") {
          setFormData({
            ...formData,
            placeID: selected.id
          });
          return;
        }
        return;

    }
  }, [selected, op, selected.kind, formData, selected.id, setFormData]);
  useEffect(() => {
    if (!actionData) return;
    console.log(actionData);
    if (!dispatch) return;
    dispatch(
      {
        type: op,
        payload: actionData
      }
    );
  });

  return (
    <div>
      <h3 className="text-2xl font-bold">{formTitle}</h3>
      <Form
        netID={netID}
        schema={action.schema}
        data={formData}
        buttons={buttons}
        setData={setFormData}
        method={"post"}
      />
    </div>
  );
}

