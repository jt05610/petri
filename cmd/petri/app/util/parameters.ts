import type { ParameterWithValue } from "~/lib/context/session";
import type { RunDetails } from "~/models/net.run.server";

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

export function getParameterRecord(run: RunDetails): Record<string, Record<number, Record<string, ParameterWithValue>>> {
  return toParameterRecord(getParameters(run));
}