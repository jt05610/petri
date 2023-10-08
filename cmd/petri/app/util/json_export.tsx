import { ReactNode } from "react";
import { NewSessionInput, StartSessionInput } from "~/models/__generated__/graphql";

type JSONExportProps = {
  json: Object
  name: string
  children: ReactNode
}

const JsonBase = "data:text/json;charset=utf-8,";

function makeJsonStr(json: Object) {
  return JsonBase + encodeURIComponent(JSON.stringify(json));
}

export default function JSONExport({ json, name, children }: JSONExportProps) {
  return (
    <a
      id="downloadAnchorElem"
      href={makeJsonStr(json)}
      download={`${name}.json`}
    >
      {children}
    </a>
  );
}

type GQLRequest = {
  operationName: string
  variables: {
    [key: string]: any
  }
  query: string
}

export function newSessionRequest(input: NewSessionInput): GQLRequest {
  return {
    operationName: "NewSession",
    variables: { input },
    query: "mutation NewSession($input: NewSessionInput!) {\n  newSession(input: $input) {\n    id\n    createdAt\n    updatedAt\n    __typename\n  }\n}"
  };
}

export function startSessionRequest(input: StartSessionInput): GQLRequest {
  return {
    operationName: "Start",
    variables: { input },
    query: "mutation Start($input: StartSessionInput!) {\n  startSession(input: $input) {\n    name\n    timestamp\n    data\n    __typename\n  }\n}"
  };
}