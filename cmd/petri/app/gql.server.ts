import type { TypedDocumentNode } from "@graphql-typed-document-node/core";
import type { DocumentNode } from "@apollo/client";
import invariant from "tiny-invariant";

invariant(process.env.GRAPHQL_URI, "GRAPHQL_URI must be set");

const GRAPHQL_URI = process.env.GRAPHQL_URI;

type GqlFetchResult<TData = any> = {
  data?: TData;
  errors?: Error[];
};