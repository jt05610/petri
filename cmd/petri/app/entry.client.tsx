/**
 * By default, Remix will handle hydrating your app on the client for you.
 * You are free to delete this file if you'd like to, but if you ever want it revealed again, you can run `npx remix reveal` ✨
 * For more information, see https://remix.run/docs/en/main/file-conventions/entry.client
 */

import { RemixBrowser } from "@remix-run/react";
import { startTransition, StrictMode } from "react";
import { hydrateRoot } from "react-dom/client";
import { ApolloClient, ApolloProvider, InMemoryCache } from "@apollo/client";

startTransition(() => {
  const client = new ApolloClient({
    uri: "http://localhost:4000/api/",
    cache: new InMemoryCache().restore(window.__APOLLO_STATE__)
  });

  hydrateRoot(
    document,
    <StrictMode>
      <ApolloProvider client={client}>
        <RemixBrowser />
      </ApolloProvider>
    </StrictMode>
  );
});
