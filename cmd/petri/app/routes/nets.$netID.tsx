import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import {
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from "@remix-run/react";
import invariant from "tiny-invariant";

import type { ArcInput, PlaceInput, TransitionInput } from "~/models/net.server";
import { getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { MermaidDiagram } from "~/lib/components/mermaid";
import { Creator } from "~/lib/components/creator";
import { useState, useEffect } from "react";
import PISchema from "~/forms/PlaceInput.schema.json";
import TISchema from "~/forms/TransitionInput.schema.json";
import AISchema from "~/forms/ArcInput.schema.json";
import type { JsonSchema, UISchemaElement } from "@jsonforms/core";
import ButtonBox from "~/lib/components/buttonbox";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");

  const net = await getNet({ id: params.netID, authorID: authorID });
  if (!net) {
    throw new Response("Not Found", { status: 404 });
  }
  return json({ net: net });
};

let forms: {
  create: {
    name: string
    schema: JsonSchema | {
      schema: JsonSchema,
      ui: UISchemaElement
    }
    route: string
    data: TransitionInput | PlaceInput | ArcInput
  }[]
} = {
  create: [
    {
      name: "Place",
      schema: PISchema,
      route: "add-place",
      data: {
        name: "",
        description: "",
        bound: 1
      }
    },
    {
      name: "Transition",
      schema: TISchema,
      route: "add-transition",
      data: {
        name: "",
        description: "",
        condition: ""
      }
    },
    {
      name: "Arc",
      schema: AISchema,
      route: "add-arc",
      data: {
        placeID: "",
        transitionID: "",
        fromPlace: true
      }
    }
  ]
};

type FormProps = JsonSchema & { route: string, netID: string, name: string };

enum NodeKind {
  Place = "Place",
  Transition = "Transition"
}

type Node = {
  id: string,
  name: string,
  kind: NodeKind,
  inputs: string[]
  outputs: string[]
}

type InspectorProps = {
  node: Node
  show: boolean,
  handleClose: () => void,
  persistOpen: boolean,
  setPersistOpen: (persistOpen: boolean) => void,
  handleSubmit: (e: React.FormEvent<HTMLFormElement>) => void,
  handleDelete: () => void,
}

export function Inspector(p: InspectorProps) {
  const [name, setName] = useState(p.node.name);
  const [inputs, setInputs] = useState(p.node.inputs);
  const [outputs, setOutputs] = useState(p.node.outputs);

  const handleDeleteInput = (inputToDelete: string) => {
    setInputs(inputs.filter(input => input !== inputToDelete));
  };
  const handleDeleteOutput = (outputToDelete: string) => {
    setOutputs(outputs.filter(output => output !== outputToDelete));
  };


  return (
    <form className={"flex flex-col"} onSubmit={p.handleSubmit}>
      <label>Name:</label>
      <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
      <div>

        <h3>Inputs:</h3>
        {inputs.map((input, i) =>
          <p key={i}>{input}
            <button type="button" onClick={() => handleDeleteInput(input)}>Delete</button>
          </p>
        )}

      </div>

      <h3>Outputs:</h3>

      <div>
        {outputs.map((output, i) =>
          <p key={i}>{output}
            <button type="button" onClick={() => handleDeleteOutput(output)}>Delete</button>
          </p>
        )}
      </div>

      <input
        type="checkbox"
        checked={p.persistOpen}
        onChange={(e) => p.setPersistOpen(e.target.checked)}
      />
      <label>Keep open</label>

      <button type="submit">Update</button>
      <button type="button" onClick={p.handleDelete}>Delete</button>
    </form>
  );
}

export default function NetDetailsPage() {
  const data = useLoaderData<typeof loader>();
  const [form, setForm] = useState<FormProps>({ ...forms.create[0].schema, name: "", route: "", netID: "" });
  const [formData, setFormData] = useState<PlaceInput | TransitionInput | ArcInput>(forms.create[0].data);
  const [persistOpen, setPersistOpen] = useState(false);
  const [modalPos, setModalPos] = useState({ left: "0px", top: "0px" });
  const [showForm, setShowForm] = useState(false);
  const [selectedNode, setSelectedNode] = useState<Node>({
    id: "",
    name: "",
    kind: NodeKind.Place,
    inputs: [],
    outputs: []
  });
  const [listening, setListening] = useState(false);
  const netID = data.net.id;
  const [showInspector, setShowInspector] = useState(false);
  let idx: { [key: string]: Node } = {};

  let nodeClickCount = 0;

  data.net.places.forEach((place) => {
    idx[place.id] = {
      id: place.id,
      name: place.name,
      kind: NodeKind.Place,
      inputs: [],
      outputs: []
    };
  });
  data.net.transitions.forEach((transition) => {
    idx[transition.id] = {
      id: transition.id,
      name: transition.name,
      kind: NodeKind.Transition,
      inputs: [],
      outputs: []
    };
  });
  data.net.arcs.forEach((arc) => {
    if (arc.fromPlace) {
      idx[arc.placeID].outputs.push(arc.transitionID);
      idx[arc.transitionID].inputs.push(arc.placeID);
    } else {
      idx[arc.transitionID].outputs.push(arc.placeID);
      idx[arc.placeID].inputs.push(arc.transitionID);
    }
  });
  const show = (name: string) => {
    if (!listening && name === "Arc") {
      setListening(true);
      nodeClickCount = 0;
    }
    const clickedForm = forms.create.find((f) => f.name === name);
    if (!clickedForm) {
      return;
    }
    setForm({ ...clickedForm.schema, name: clickedForm.name, route: clickedForm.route, netID: netID });
    setFormData(clickedForm.data);
    setShowForm(true);
  };

  const createClicked = async (e: React.MouseEvent<HTMLButtonElement>) => {
    await show(e.currentTarget.name);
  };
  const code = (): string => {
    if (data.net.places.length === 0) {
      return "";
    }

    const places = data.net.places.map((place) => {
      return `${place.id}((${place.name}))\n  click ${place.id} clicked "${place.name}"`;
    });
    const transitions = data.net.transitions.map((transition) => {
      return `${transition.id}[${transition.name}] \n  click ${transition.id} clicked "${transition.name}"`;
    });
    const arcs = data.net.arcs.map((arc) => {
      console.log(arc);
      if (arc.fromPlace) {
        return `${arc.placeID} --> ${arc.transitionID}`;

      } else {
        return `${arc.transitionID} --> ${arc.placeID}`;
      }
    });
    let out = `graph TD
  ${places.join("\n  ")}
  ${transitions.join("\n  ")}
  ${arcs.join("\n  ")}
  `;
    console.log(out);
    return out;
  };
  const nodeClicked = (e: string) => {
    console.log(e);
    let n = idx[e];
    if (!n) {
      return;
    }
    if (!listening) {
      setSelectedNode(n);
      setShowInspector(true);
    }
    if (listening) {
      let upd = { ...formData, [n.kind === NodeKind.Place ? "placeID" : "transitionID"]: n.id };
      if (nodeClickCount === 0) {
        upd = { ...upd, fromPlace: n.kind === NodeKind.Place };
      }
      if (nodeClickCount === 1) {
        setShowForm(false);
        setListening(false);
        return;
      }
      nodeClickCount++;
      setFormData(upd);
    }
  };

  function closeInspector() {
    setShowInspector(false);
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    setShowForm(false);
  }

  function handleDelete() {
  }

  const handleMouseMove = (event: MouseEvent) => {
    setModalPos({ left: `${event.pageX}px`, top: `${event.pageY}px` });
  };


  useEffect(() => {
    if (showInspector) {
      document.addEventListener("mousemove", handleMouseMove);
      if (!persistOpen) {
        setTimeout(() => {
          closeInspector();
        }, 5000); // close after 5 seconds
      }
    }
    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
    };
  }, [showInspector, persistOpen]);

  return (
    <div>
      <h3 className="text-2xl font-bold">{data.net.name}</h3>
      <p className="py-6">{data.net.description}</p>
      <div className="flex flex-row w-full items-center">
        <MermaidDiagram code={code()} nodeClicked={nodeClicked} />
      </div>
      <ButtonBox
        name={"Create"}
        handlers={
          forms.create.map((form) => form.name)
        }
        onClick={createClicked} />
      {showForm
        ? <Creator
          netID={form.netID}
          schema={form}
          data={formData}
          setData={setFormData}
          method={"post"}
          route={form.route}
          onDone={() => {
            setShowForm(false);
          }}
        />
        : null
      }
      <div className="flex flex-row w-full items-center" style={modalPos}>
        <Inspector node={selectedNode} show={showInspector} handleClose={closeInspector} handleSubmit={handleSubmit}
                   handleDelete={handleDelete} persistOpen={persistOpen} setPersistOpen={setPersistOpen} />
      </div>
    </div>
  );
};

export function ErrorBoundary() {
  const error = useRouteError();

  if (error instanceof Error) {
    return <div>An unexpected error occurred: {error.message}</div>;
  }

  if (!isRouteErrorResponse(error)) {
    return <h1>Unknown Error</h1>;
  }

  if (error.status === 404) {
    return <div>Note not found</div>;
  }

  return <div>An unexpected error occurred: {error.statusText}</div>;
}
