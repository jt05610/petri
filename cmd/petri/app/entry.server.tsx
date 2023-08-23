/**
 * By default, Remix will handle generating the HTTP Response for you.
 * You are free to delete this file if you'd like to, but if you ever want it revealed again, you can run `npx remix reveal` ✨
 * For more information, see https://remix.run/docs/en/main/file-conventions/entry.server
 */

import { PassThrough } from "node:stream";

import type { EntryContext } from "@remix-run/node";
import { Response } from "@remix-run/node";
import { RemixServer } from "@remix-run/react";
import isbot from "isbot";
import { renderToPipeableStream } from "react-dom/server";
import { ApolloClient, ApolloProvider, createHttpLink, InMemoryCache } from "@apollo/client";
import { loadErrorMessages, loadDevMessages } from "@apollo/client/dev";
import { getDataFromTree } from "@apollo/client/react/ssr";

const ABORT_DELAY = 5_000;

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  return isbot(request.headers.get("user-agent"))
    ? handleBotRequest(
      request,
      responseStatusCode,
      responseHeaders,
      remixContext
    )
    : handleBrowserRequest(
      request,
      responseStatusCode,
      responseHeaders,
      remixContext
    );
}

function handleBotRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  return new Promise((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <RemixServer
        context={remixContext}
        url={request.url}
        abortDelay={ABORT_DELAY}
      />,
      {
        onAllReady() {
          const body = new PassThrough();

          responseHeaders.set("Content-Type", "text/html");

          resolve(
            new Response(body, {
              headers: responseHeaders,
              status: responseStatusCode
            })
          );

          pipe(body);
        },
        onShellError(error: unknown) {
          reject(error);
        },
        onError(error: unknown) {
          responseStatusCode = 500;
          console.error(error);
        }
      }
    );

    setTimeout(abort, ABORT_DELAY);
  });
}

async function handleBrowserRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {

  function makeHeaders(): Record<string, string> {
    const headers: Record<string, string> = {};
    for (const [key, value] of request.headers.entries()) {
      headers[key] = value;
    }
    return headers;
  }

  const client = new ApolloClient({
    ssrMode: true,
    link: createHttpLink({
      uri: "http://localhost:8081/graphql",
      headers: makeHeaders(),
      credentials: request.credentials ?? "same-origin"
    }),
    cache: new InMemoryCache()
  });

  if (process.env.NODE_ENV === "development") {
    loadDevMessages();

    loadErrorMessages();
  }
  const App = (
    <ApolloProvider client={client}>
      <RemixServer
        context={remixContext}
        url={request.url}
        abortDelay={ABORT_DELAY}
      />
    </ApolloProvider>
  );

  return new Promise(async (resolve, reject) => {
    return await getDataFromTree(App).then(() => {
      const initialState = client.extract();
      const { pipe, abort } = renderToPipeableStream(
        <>
          {App}
          <script
            dangerouslySetInnerHTML={{
              __html: `window.__APOLLO_STATE__=${JSON.stringify(
                initialState
              ).replace(/</g, "\\u003c")}` // The replace call escapes the < character to prevent cross-site scripting attacks that are possible via the presence of </script> in a string literal
            }}
          />
        </>,
        {
          onShellReady() {
            const body = new PassThrough();

            responseHeaders.set("Content-Type", "text/html");

            resolve(
              new Response(body, {
                headers: responseHeaders,
                status: responseStatusCode
              })
            );

            pipe(body);
          },
          onShellError(error: unknown) {
            reject(error);
          },
          onError(error: unknown) {
            console.error(error);
            responseStatusCode = 500;
          }
        }
      );
      setTimeout(abort, ABORT_DELAY);
    });
  });
}
