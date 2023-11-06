import type { ActionArgs, LoaderArgs, V2_MetaFunction } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { Form, Link, useActionData, useSearchParams } from "@remix-run/react";
import type { ReactNode } from "react";
import { useId } from "react";

import type { VerifyLoginInput } from "~/models/user.server";
import { verifyLogin, VerifyLoginSchema } from "~/models/user.server";
import { createUserSession, getUserId } from "~/session.server";
import { safeRedirect } from "~/utils";
import { parse } from "@conform-to/zod";
import type { Submission } from "@conform-to/react";
import { conform, useForm } from "@conform-to/react";

export const loader = async ({ request }: LoaderArgs) => {
  console.log("request", request);
  const userId = await getUserId(request);
  if (userId) return redirect("/");
  return json({});
};

function validateForm(form: FormData): Submission<VerifyLoginInput> {
  return parse(
    form,
    { schema: VerifyLoginSchema }
  );
}

function validateJSON(json: any): VerifyLoginInput | unknown {
  try {
    return VerifyLoginSchema.parse(json);
  } catch (e) {
    return e;
  }
}

export const action = async ({ request }: ActionArgs) => {
  let submission: VerifyLoginInput;
  if (request.headers.get("Content-Type")?.includes("application/json")) {
    const sub = validateJSON(await request.json());
    if (sub instanceof Error) {
      return json({ errors: sub });
    }
    submission = sub as VerifyLoginInput;
  } else {
    const sub = validateForm(await request.formData());
    if (!sub.value || sub.intent !== "submit") {
      return sub;
    }
    submission = sub.value;
  }
  if (!submission) {
    return json({ errors: { email: "Email is required", password: "Password is required" } });
  }
  if (!submission.redirectTo) {
    submission.redirectTo = "/";
  }

  if (submission.remember === undefined) {
    submission.remember = false;
  }

  const user = await verifyLogin(submission);

  if (!user) {
    return json({ errors: { email: "Invalid email or password", password: "Invalid email or password" } });
  }

  console.log("user", user);
  return await createUserSession({
    redirectTo: safeRedirect(submission.redirectTo),
    remember: submission.remember,
    request,
    userId: user.id
  });
};

export const meta: V2_MetaFunction = () => [{ title: "Login" }];

enum FormLayoutType {
  Horizontal = "horizontal",
  Vertical = "vertical"
}

type FormLayoutProps = {
  children: ReactNode
  type: FormLayoutType
}

function FormLayout({ children, type }: FormLayoutProps) {
  switch (type) {
    case FormLayoutType.Horizontal:
      return (
        <div className="flex items-center justify-between">
          {children}
        </div>
      );
    case FormLayoutType.Vertical:
      return (
        <div className="flex flex-col space-y-4">
          {children}
        </div>
      );

  }
}

export default function LoginPage() {
  const [searchParams] = useSearchParams();
  const redirect = searchParams.get("redirectTo") || "/nets";
  const actionData = useActionData<typeof action>();
  const id = useId();
  const [form, { email, password, remember, redirectTo }] = useForm({
    id,
    lastSubmission: actionData as Submission<VerifyLoginInput>
  });

  return (
    <div className="flex min-h-full flex-col justify-center">
      <div className="mx-auto w-full max-w-md px-8">
        <Form method="post" className="space-y-6" {...form.props}>
          <input {...conform.input(redirectTo, { type: "hidden" })} value={redirect} />
          <div>
            <label
              htmlFor={email.id}
              className="block text-sm font-medium text-gray-700 dark:text-gray-200"
            >
              Email
            </label>
            <input
              {...conform.input(email, { type: "email" })}
              className="w-full rounded-full border border-gray-500 px-2 py-1 text-lg dark:bg-slate-500 dark:text-gray-200"
            />
            <div id={email.errorId}>{email.errors}</div>
          </div>
          <div>
            <label
              htmlFor={password.id}
              className="block text-sm font-medium text-gray-700 dark:text-gray-200"
            >
              Password
            </label>
            <input
              {...conform.input(password, { type: "password" })}
              className="w-full rounded-full border border-gray-500 px-2 py-1 text-lg dark:bg-slate-500 dark:text-gray-200"
            />
            <div id={password.errorId}>{password.errors}</div>
          </div>
          <button
            type="submit"
            className="w-full rounded-full bg-blue-500 dark:bg-teal-700 px-4 py-2 text-white hover:bg-blue-600 focus:bg-blue-400"
          >
            Log in
          </button>
          <FormLayout type={FormLayoutType.Horizontal}>
            <div className="flex items-center">
              <input
                {...conform.input(remember, { type: "checkbox" })}
                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                value={"true"}
              />
              <label
                htmlFor="remember"
                className="ml-2 block text-sm text-gray-900 dark:text-gray-200"
              >
                Remember me
              </label>
            </div>
            <div className="text-center text-sm text-gray-500 dark:text-gray-400">
              Don't have an account?{" "}
              <Link
                className="text-blue-500 underline"
                to={{
                  pathname: "/join",
                  search: searchParams.toString()
                }}
              >
                Sign up
              </Link>
            </div>
          </FormLayout>
        </Form>
      </div>
    </div>
  );
}
