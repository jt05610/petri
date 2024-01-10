import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { MockedProvider } from "@apollo/client/testing";
import { TokenSchema } from "./src/token";
import { GetTokenSchemaDocument } from "./src/gql/graphql";
import { expect, test } from "vitest";

import React from "react";

const mocks = [
  {
    request: {
      query: GetTokenSchemaDocument,
      variables: {
        id: "person",
      },
    },
    result: {
      data: {
        tokenSchema: {
          id: "person",
          name: "Person",
          type: "OBJECT",
          properties: {
            name: {
              type: "STRING",
            },
            age: {
              type: "INTEGER",
            },
            hometown: {
              type: "OBJECT",
              properties: {
                city: {
                  type: "STRING",
                },
                state: {
                  type: "STRING",
                },
              },
            },
          },
        },
      },
    },
  },
];

test("renders without error", async () => {
  render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <TokenSchema id={"person"} />
    </MockedProvider>
  );
  expect(screen.getByText("Loading...")).toBeInTheDocument();
});
