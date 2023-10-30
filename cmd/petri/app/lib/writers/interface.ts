import PetriNetWriter from "~/util/writer";

export enum ImplementedLanguage {
  TypeScript = "TypeScript",
}

export default interface InterfaceWriter extends PetriNetWriter {
  language: ImplementedLanguage;
}