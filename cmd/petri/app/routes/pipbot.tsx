import { Form, useActionData } from "@remix-run/react";
import type { FieldConfig } from "@conform-to/react";
import { list, useFieldList, useFieldset, useForm } from "@conform-to/react";
import { parse } from "@conform-to/zod";
import type { ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import type { PipBotDesignInput, SoluteInput, SolutionInput, StockInput, SolventInput } from "~/models/pipbot";
import { PipBotDesignInputSchema } from "~/models/pipbot";
import { CheckboxField, Field, FieldSetTable } from "~/lib/components/FormFieldSet";
import type { ReactNode } from "react";
import { useRef, useState } from "react";
import { makeTransfers } from "~/models/pipbot.server";

export const action = async ({ request }: ActionArgs) => {
  const formData = await request.formData();
  const submission = parse(formData, { schema: PipBotDesignInputSchema });
  if (submission.intent !== "submit" || !submission.value) {
    return json(submission);
  }
  const transfers = await makeTransfers(submission.value);
  console.log("transfers", transfers);
  return json(transfers);
};

export function FieldHeader({ children }: { children: ReactNode }) {
  return (
    <h2
      className={"text-lg font-semibold text-center text-slate-900 dark:text-slate-100"}
    >
      {children}
    </h2>
  );
}

function FieldSection({ children, visible, toggleFunc, title }: {
  children: ReactNode,
  visible: boolean,
  toggleFunc: () => void,
  title: string
}) {
  return (
    <div className={"flex border rounded-md border-gray-300 flex-col space-y-1 m-2 min-w-full min-h-"}>
      <FieldHeader>
        {title}
      </FieldHeader>
      <button
        type={"button"}
        onClick={toggleFunc}
      >
        {visible ? "Hide" : "Show"}
      </button>
      {visible && children}
    </div>
  );
}

export default function PipBot() {
  const lastSubmission = useActionData<typeof action>();
  const [sectionVisibility, setSectionVisibility] = useState({
    solutes: true,
    solvents: true,
    stocks: true,
    samples: true,
    curve: true,
    grids: true
  });

  const toggleSection = (section: keyof typeof sectionVisibility) => {
    setSectionVisibility({
      ...sectionVisibility,
      [section]: !sectionVisibility[section]
    });
  };

  const toggleHandler = (section: keyof typeof sectionVisibility) => {
    return () => {
      toggleSection(section);
    };
  };
  const [soluteSelections, setSolutes] = useState<string[]>([]);
  const [solventSelections, setSolvents] = useState<string[]>([]);

  const [form, { solutes, solvents, stocks, samples, grids, curve, dest_grid }] = useForm({
    lastSubmission,
    shouldValidate: "onBlur",
    onValidate: ({ formData }) => {
      const sub = parse(formData, { schema: PipBotDesignInputSchema });
      console.log("sub", sub);
      if (sub.payload) {
        if (sub.payload.solutes) {
          setSolutes(sub.payload.solutes.map((solute: SoluteInput) => {
            console.log("solute", solute);
            return solute.name;
          }));
        }
        if (sub.payload.solvents) {
          setSolvents(sub.payload.solvents.map((solvent: SolventInput) => {
            console.log("solvent", solvent);
            return solvent.name;
          }));
        }
      }
      return sub;
    }
  });
  const soluteList = useFieldList(form.ref, solutes);
  const solventList = useFieldList(form.ref, solvents);
  const gridsList = useFieldList(form.ref, grids);
  return (
    <div className={"flex w-full items-center justify-center"}>
      <Form
        method={"post"}
        {...form.props}
        className={"flex flex-col flex-wrap items-center justify-center space-2"}
      >
        <FieldSection visible={sectionVisibility.solutes} toggleFunc={toggleHandler("solutes")} title={"Solutes"}>
          <FieldSetTable config={solutes} fields={["name", "molecular_weight"]} fieldList={soluteList} />
        </FieldSection>
        <FieldSection visible={sectionVisibility.solvents} toggleFunc={toggleHandler("solvents")} title={"Solvents"}>
          <FieldSetTable config={solvents} fields={["name"]} fieldList={solventList} />
        </FieldSection>
        <FieldSection visible={sectionVisibility.stocks} toggleFunc={toggleHandler("stocks")} title={"Stocks"}>
          <StockListField config={stocks} solventChoices={solventSelections} soluteChoices={soluteSelections} />
        </FieldSection>
        <FieldSection visible={sectionVisibility.samples} toggleFunc={toggleHandler("samples")} title={"Samples"}>
          <SampleListField config={samples} solventChoices={solventSelections} soluteChoices={soluteSelections} />
        </FieldSection>
        <FieldSection visible={sectionVisibility.curve} toggleFunc={toggleHandler("curve")} title={"Standard Curve"}>
          <CurveField {...curve} />
        </FieldSection>
        <FieldSection visible={sectionVisibility.grids} toggleFunc={toggleHandler("grids")} title={"Grids"}>
          <FieldSetTable config={grids} fields={[
            "name",
            "n_wells",
            "n_cols",
            "n_rows",
            "row_space",
            "col_space"
          ]} fieldList={gridsList} />
        </FieldSection>
        <Field {...dest_grid} />
        <button
          type={"submit"}
          className={"bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-full"}
        >
          Submit
        </button>
      </Form>
    </div>
  );
}

type SampleListFieldProps = {
  config: FieldConfig<PipBotDesignInput["samples"]>
  soluteChoices: string[]
  solventChoices: string[]
}

function SampleListField({ config, soluteChoices, solventChoices }: SampleListFieldProps) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const sampleList = useFieldList(ref, config);
  return (
    <fieldset ref={ref}>
      {sampleList.map((sample) => {
        return <SolutionField config={sample} key={sample.key} soluteChoices={soluteChoices}
                              solventChoices={solventChoices} />;
      })}
      <button {...list.append(config.name)}>
        Add
      </button>
    </fieldset>
  );
}

type SolutionFieldProps = {
  config: FieldConfig<SolutionInput>
  soluteChoices: string[]
  solventChoices: string[]
}

function SolutionField({ config, soluteChoices, solventChoices }: SolutionFieldProps) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const { name, solutes, solvents } = useFieldset(ref, config);
  const soluteList = useFieldList(ref, solutes);
  const solventList = useFieldList(ref, solvents);
  return (
    <fieldset ref={ref}>
      <Field {...name} />
      <FieldHeader>
        Solutes
      </FieldHeader>
      <FieldSetTable config={solutes} fields={[{
        name: "name",
        options: soluteChoices
      }, "amount"]} fieldList={soluteList} />
      <FieldHeader>
        Solvents
      </FieldHeader>
      <FieldSetTable config={solvents} fields={[{
        name: "name",
        options: solventChoices
      }, "amount"]} fieldList={solventList} />
    </fieldset>
  );
}

type StockFieldProps = {
  config: FieldConfig<StockInput>
  index: number
  parent: FieldConfig<StockInput[]>
  soluteChoices: string[]
  solventChoices: string[]
}

function StockField({ config, index, parent, soluteChoices, solventChoices }: StockFieldProps) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const { name, grid, pos, volume, solution } = useFieldset(ref, config);
  return (
    <div className={"flex flex-col space-y-2"}>
      <Field {...name} />
      <Field {...grid} />
      <Field {...pos} />
      <Field {...volume} />
      <SolutionField config={solution} soluteChoices={soluteChoices} solventChoices={solventChoices} />
      <button {...list.remove(parent.name, { index: index })}>
        Remove
      </button>
    </div>
  );
}

type StockListFieldProps = {
  config: FieldConfig<PipBotDesignInput["stocks"]>
  soluteChoices: string[]
  solventChoices: string[]
}

function StockListField({ config, soluteChoices, solventChoices }: StockListFieldProps) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const stockList = useFieldList(ref, config);
  return (
    <fieldset ref={ref}>
      {stockList.map((stock, index) => {
        return <StockField config={stock} key={stock.key} parent={config} index={index} solventChoices={solventChoices}
                           soluteChoices={soluteChoices} />;
      })}
      <button {...list.append(config.name)}>
        Add
      </button>
    </fieldset>
  );
}

function CurveField(config: FieldConfig<PipBotDesignInput["curve"]>) {
  const ref = useRef<HTMLFieldSetElement>(null);
  const {
    stock,
    diluent,
    volume,
    start_row,
    start_col,
    increment_along_row,
    serial_dilution,
    standards
  } = useFieldset(ref, config);

  const standardList = useFieldList(ref, standards);

  return (
    <fieldset ref={ref}
    >
      <Field {...stock} />
      <Field {...diluent} />
      <Field {...volume} />
      <Field {...start_row} />
      <Field {...start_col} />
      <CheckboxField {...increment_along_row} />
      <CheckboxField {...serial_dilution} />
      <FieldHeader>
        Standards
      </FieldHeader>
      <FieldSetTable config={standards} fields={["dilution_factor", "source"]} fieldList={standardList} />
    </fieldset>
  );
}