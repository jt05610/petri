import { useQuery } from "@apollo/client";
import { GetTokenSchemaDocument } from "./gql/graphql";
import React from "react";

interface TokenProps {
  id: string;
}

export function TokenSchema({ id }: TokenProps) {
  const { loading, error, data } = useQuery(GetTokenSchemaDocument, {
    variables: { id },
  });
  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error :(</p>;
  return (
    <>
      <div>
        <h3>Token Schema</h3>
        <p>{JSON.stringify(data)}</p>
      </div>
    </>
  );
}
