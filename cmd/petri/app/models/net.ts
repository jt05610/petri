export namespace Net {
  export interface ReachableState {
    id: string;
    marking: number[];
    children: ReachableState[];
  }

  export interface ReachabilityTree {
    id: string;
    head: ReachableState;
  }

  export interface Node {
    readonly id: string;
    readonly name: string;
  }

  export interface Place extends Node {
    readonly bound: number;
    readonly initialMarking: number;
    readonly marking: number;
  }

  export interface Type extends Node {
    fields?: Field[];
  }

  export interface Field extends Node {
    type: Type;
    required: boolean;
  }

  export interface Event extends Node {
    fields?: Field[];
  }

  export type Handler<S> = (net: PetriNet, payload: S) => PetriNet;

  export interface Transition extends Node {
    events: Event[];
  }

  export interface Arc {
    id: string;
    from: Place | Transition;
    to: Place | Transition;
  }

  export interface PetriNet {
    readonly id: string;
    readonly name: string;
    readonly places: Place[];
    readonly transitions: Transition[];
    readonly arcs: Arc[];
    readonly children: PetriNet[];
  }

  export interface Loader {
    load(id: string): Promise<PetriNet | null>;
  }

  export interface Flusher {
    flush(net: PetriNet): Promise<void | null>;
  }

  export interface Lister {
    list(): Promise<PetriNet[]>;
  }

  export interface Creator<S> {
    create(net: S): Promise<PetriNet>;
  }

  export interface Updater<S> {
    update(id: string, net: S): Promise<PetriNet>;
  }

  export interface Deleter {
    delete(id: string): Promise<void>;
  }

  export interface Service<S, T> extends Loader, Flusher, Lister, Creator<S>, Updater<T>, Deleter {
  }

  export interface Builder {
    newNet(name: string): PetriNet;

    addPlace(net: PetriNet, ...place: Omit<Place, "id">[]): PetriNet;

    addTransition(net: PetriNet, ...transition: Omit<Transition, "id">[]): PetriNet;

    addArc(net: PetriNet, ...arc: Omit<Arc, "id">[]): PetriNet;

    addChild(net: PetriNet, ...children: PetriNet[]): PetriNet;

    copy(net: PetriNet): PetriNet;

    removePlace(net: PetriNet, ...place: (Pick<Place, "name"> | Pick<Place, "id">)[]): PetriNet;

    removeTransition(net: PetriNet, ...transition: (Pick<Transition, "name"> | Pick<Transition, "id">)[]): PetriNet;

    removeArc(net: PetriNet, ...arc: (Pick<Arc, "id">)[]): PetriNet;

    removeChild(net: PetriNet, ...children: Pick<PetriNet, "id">[]): PetriNet;

    updatePlace(net: PetriNet, place: Place): PetriNet;

    updateTransition(net: PetriNet, transition: Transition): PetriNet;

    updateArc(net: PetriNet, arc: Arc): PetriNet;

    rename(net: PetriNet, name: string): PetriNet;

  }

  export interface Controller {
    inputs(net: PetriNet, node: Pick<Place, "id"> | Pick<Transition, "id">): Place[] | Transition[];

    outputs(net: PetriNet, node: Pick<Place, "id"> | Pick<Transition, "id">): Place[] | Transition[];

    fireTransition(net: PetriNet, transitionId: Pick<Transition, "id">): PetriNet;

    availableTransitions(net: PetriNet): Transition[];

    marking(net: PetriNet): Map<string, number>;
  }

  export type EventInput = {
    name: string;
    fields: {
      name: string;
      type: Type;
      required: boolean;
    }[]
  }

  export interface EventService {

    registerEvent(net: PetriNet, transition: Pick<Transition, "id">, event: EventInput): PetriNet;

    registerHandler<S>(net: PetriNet, event: Pick<Event, "id">, handler: Handler<S>): PetriNet;

    canHandleEvent(net: PetriNet, eventId: Pick<Event, "id">): boolean;

    handleEvent<S>(net: PetriNet, eventId: Pick<Event, "id">, payload: S): PetriNet;

    isValidEventSequence(net: PetriNet, sequence: Pick<Event, "id">[]): boolean;

    availableEvents(net: PetriNet): Event[];
  }

  export function isPlace(obj: any): obj is Net.Place {
    return obj.bound !== undefined;
  }

  export function isTransition(obj: any): obj is Net.Transition {
    return obj.events !== undefined;
  }

  export function sameKind(obj1: Net.Node, obj2: Net.Node): boolean {
    return (isPlace(obj1) && isPlace(obj2)) || (isTransition(obj1) && isTransition(obj2));
  }

  export const String: Type = {
    id: "string",
    name: "string",
    fields: []
  };

  export const Number: Type = {
    id: "number",
    name: "number",
    fields: []
  };

  export const Boolean: Type = {
    id: "boolean",
    name: "boolean",
    fields: []
  };

  export interface Array extends Type {
    type: Type;
  }

}
