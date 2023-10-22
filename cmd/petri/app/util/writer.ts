import { PetriNet } from "~/util/petrinet";


export default interface PetriNetWriter {
  writePetriNet(petriNet: PetriNet): string;
}