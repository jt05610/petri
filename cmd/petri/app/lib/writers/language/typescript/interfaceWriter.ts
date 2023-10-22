import InterfaceWriter, { ImplementedLanguage } from "~/lib/writers/interface";
import { PetriNet } from "~/util/petrinet";
import { EventDetails } from "~/models/net.server";

interface InterfaceMethod {
  methodName: string;
  returnType: string;
  params: Map<string, string>;
}

type InterfaceDefinition = {
  name: string;
  methods: InterfaceMethod[]
};

function convertEvent(event: EventDetails): InterfaceMethod {
  const methodName = event.name;
  const returnType = "ResultOrError<" + pascalCase(event.name) + "Response>";
  const params = new Map<string, string>();

  event.fields.forEach((field) => {
    params.set(field.name, field.type);
  });

  return {
    methodName,
    returnType,
    params
  };
}


export default class Writer implements InterfaceWriter {
  language = ImplementedLanguage.TypeScript;

  writeInterface(settings: InterfaceDefinition): string {
    let output = `export interface ${pascalCase(settings.name)} {`;

    settings.methods.forEach((method) => {

      output += `\n  ${camelCase(method.methodName)}(params: ${pascalCase(method.methodName)}Request): ${method.returnType};\n`;
    });

    output += "}";

    return output;
  }

  writePetriNet(petriNet: PetriNet): string {
    const methods = new Map<string, InterfaceMethod>();
    petriNet.net.children[0].transitions.forEach((transition) => {
      if (!transition.events) {
        return;
      }
      const events = transition.events.map((event) => convertEvent(event));
      events.forEach((event) => {
          methods.set(event.methodName, event);
        }
      );
    });

    const interfaceDefinition: InterfaceDefinition = {
      name: petriNet.net.children[0].name,
      methods: [...methods.values()]
    };
    const contents = [
      resultOrError
    ]
      .concat([...methods.values()].map((method) => requestDefinition(method)))
      .concat([...methods.values()].map((method) => responseDefinition(method))
      );
    return contents.join("\n") + "\n" + this.writeInterface(interfaceDefinition);
  }
}

// language=TypeScript
const resultOrError = `interface ResultOrError<T> {
  result?: T;
  error?: string;
}
`;

export function pascalCase(str: string): string {
  // converts the first character of the string to Uppercase
  // if there are spaces, underscores, or dashes, it removes them and converts the next character to upper case
  // if there are any other characters, it converts them to lower case
  return str.replace(/^\w|[A-Z]|\b\w/g, function(word, index) {
      return word.toUpperCase();
    }
  ).replace(/\s+|_+|-+/g, "");
}

function firstCapital(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function camelCase(str: string): string {
  // converts the first character of the string to lower case
  // if there are spaces, underscores, or dashes, it removes them and converts the next character to upper case
  // if there are any other characters, it converts them to lower case
  return str.replace(/^\w|[A-Z]|\b\w/g, function(word, index) {
      return index === 0 ? word.toLowerCase() : word.toUpperCase();
    }
  ).replace(/\s+|_+|-+/g, "");
}

function removeWhitespace(str: string): string {
  return str.replace(/\s/g, "");
}

enum RequestOrResponse {
  Request = "Request",
  Response = "Response"
}

const typeDefinition = (requestOrResponse: RequestOrResponse, event: InterfaceMethod) => {
  let fields = "";
  event.params.forEach((type, name) => {
    fields += `\n  ${camelCase(name)}: ${type};`;
  });
  return `type ${pascalCase(event.methodName)}${requestOrResponse} = {${fields}
}
`;
};

const requestDefinition = (event: InterfaceMethod) => {
  return typeDefinition(RequestOrResponse.Request, event);
};

const responseDefinition = (event: InterfaceMethod) => {
  return typeDefinition(RequestOrResponse.Response, event);
};