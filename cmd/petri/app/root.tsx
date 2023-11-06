import { cssBundleHref } from "@remix-run/css-bundle";
import type { LinksFunction, LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import {
  Links,
  LiveReload,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration
} from "@remix-run/react";

import { getUser } from "./session.server";
import stylesheet from "./tailwind.css";
import { ThemeProvider } from "~/lib/context/Theme";

export const links: LinksFunction = () => [
  { rel: "stylesheet", href: stylesheet },
  ...(cssBundleHref ? [{ rel: "stylesheet", href: cssBundleHref }] : [])
];

export const loader = async ({ request }: LoaderArgs) => {
  return json({ user: await getUser(request) });
};

export default function App() {
  return (
    <html lang="en" className="h-full dark">
    <head>
      <meta charSet="utf-8" />
      <meta name="viewport" content="width=device-width,initial-scale=1" />
      <Meta />
      <Links /><title>Petri</title>
    </head>
    <ThemeProvider>
      <body className={`h-full antialiased dark:text-slate-400 bg-white dark:bg-slate-900`}>
      <Outlet />
      <ScrollRestoration />
      <Scripts />
      <LiveReload />
      </body>
    </ThemeProvider>

    </html>
  );
}
