import type { ParameterWithValue } from "~/lib/context/session";
import type { RunDetails } from "~/models/net.run.server";
import * as math from "mathjs";

export function toParameterRecord(parameters: Record<string, ParameterWithValue[]>): Record<string, Record<number, Record<string, ParameterWithValue>>> {
  const pr: Record<string, Record<number, Record<string, ParameterWithValue>>> = {};
  Object.entries(parameters).forEach(([deviceID, parameters]) => {
      pr[deviceID] = {};
      parameters.forEach((param) => {
        pr[deviceID][param.parameter.order] = {
          ...pr[deviceID][param.parameter.order],
          [param.parameter.fieldID]: param
        };
      });
    }
  );
  console.log("param record", pr);
  return pr;
}

export function getParameters(run: RunDetails): Record<string, ParameterWithValue[]> {
  // make all EventConstants from the run
  const eventConstants = run.steps.map(({ action, order }) => {
    return {
      order,
      deviceID: action.device.id,
      event: action.event!,
      constants: action.constants
    };
  });
  console.log(eventConstants);
  const params: Record<string, ParameterWithValue[]> = {};
  eventConstants.forEach(({ order, event, constants, deviceID }) => {
    event.fields.forEach((field) => {
      let value = "";
      const constant = constants.find((c) => c.field.id === field.id);
      if (constant) {
        value = constant.value;
      }
      console.log(event);
      if (!params[deviceID]) params[deviceID] = [];
      params[deviceID].push(
        {
          parameter: {
            order,
            eventID: event.id,
            deviceID: deviceID,
            fieldID: field.id,
            fieldName: field.name,
            fieldType: field.type
          },
          value
        });
    });
  });
  console.log("params", params);
  return params;
}

export type ParameterRecord = Record<string, Record<number, Record<string, ParameterWithValue>>>;

export function getParameterRecord(run: RunDetails): ParameterRecord {
  return toParameterRecord(getParameters(run));
}

export interface ExperimentParameter {
  name: string;
  deviceID: string;
  step: number;
  fieldID: string;
  expressions: string[];
}

function evaluate(parser: math.Parser, expressions: string[], value: string): string {
  expressions.forEach((expression) => {
    parser.evaluate(expression);
  });
  return parser.evaluate(value);
}

export interface ExperimentParameterWithValue {
  parameter: ExperimentParameter;
  value: string;
}

function setValue(parser: math.Parser, pr: ParameterRecord, param: ExperimentParameter, value: string) {
  pr[param.deviceID][param.step][param.fieldID].value = evaluate(parser, param.expressions, value);
}

export function evaluateParameters(pr: ParameterRecord, params: ExperimentParameterWithValue[]): ParameterRecord {
  const prCopy = JSON.parse(JSON.stringify(pr));
  const parser = math.parser();
  params.forEach(({ parameter, value }) => {
    setValue(parser, prCopy, parameter, value);
  });
  return prCopy;
}

export function initialParameters(pr: ParameterRecord): ExperimentParameterWithValue[] {
  const params: ExperimentParameterWithValue[] = [];
  Object.values(pr).forEach((device) => {
    Object.values(device).forEach((step) => {
      Object.values(step).forEach((param) => {
        params.push({
          parameter: {
            name: param.parameter.fieldName,
            deviceID: param.parameter.deviceID,
            step: param.parameter.order,
            fieldID: param.parameter.fieldID,
            expressions: []
          },
          value: param.value.toString()
        });
      });
    });
  });
  // remove anywhere ${name}_${step} is duplicated
  const seen: Record<string, boolean> = {};
  return params.filter((param) => {
    const key = `${param.parameter.name}_${param.parameter.step}`;
    if (seen[key]) {
      return false;
    }
    seen[key] = true;
    return true;
  });
}

export function evaluateEntry(parser: math.Parser, entry: string): string {
  // get everything from the entry in between {}
  const expressions = entry.match(/{(.*?)}/g);
  if (!expressions) {
    return entry;
  }

  let evaluatedEntry = entry;

  for (const expression of expressions) {
    const variableName = expression.slice(1, -1); // Remove { and }
    let value;

    try {
      value = parser.get(variableName);
    } catch (e) {
      throw e
    }

    // Replace {variable} with its value in the evaluatedEntry
    evaluatedEntry = evaluatedEntry.replace(expression, value);
  }

  return evaluatedEntry;
}

export function listParameters(parser: math.Parser): string[] {
  return Object.keys(parser.getAll())
}

export function validEntry(parser: math.Parser, entry: string): boolean {
  try {
    evaluateEntry(parser, entry);
  } catch (e) {
    return false;
  }
  return true;
}

