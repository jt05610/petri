import type { FormInput } from "~/lib/components/formInput";
import { createContext, Dispatch, SetStateAction, useState } from "react";


type FormContext = {
  data: FormInput
}

export const FormContext = createContext<FormContext>({
  data: {} as FormInput
});

export const FormSetterContext = createContext<Dispatch<SetStateAction<FormContext>> | null>(null);

type FormProviderProps = {
  children: React.ReactNode
}

export function FormProvider(props: FormProviderProps) {
  const [form, setForm] = useState<FormContext>({
    data: {} as FormInput
  });
  return (
    <FormContext.Provider value={form}>
      <FormSetterContext.Provider value={setForm}>
        {props.children}
      </FormSetterContext.Provider>
    </FormContext.Provider>
  );

}