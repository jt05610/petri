import { Net } from "~/models/net";
import { createId } from "@paralleldrive/cuid2";

export const ArcFromNotDefined = new Error("Arc from not defined");
export const ArcToNotDefined = new Error("Arc not found");
export const PlaceNotDefined = new Error("Place not defined");
export const PlaceNotFound = new Error("Place not found");
export const TransitionNotDefined = new Error("Transition not defined");
export const TransitionNotFound = new Error("Transition not found");
export const ArcBetweenSameType = new Error("Arc from and to cannot be of the same type");

export class Builder implements Net.Builder {

  addArc(net: Net.PetriNet, ...arc: Omit<Net.Arc, "id">[]): Net.PetriNet {

    return {
      ...net,
      arcs: [
        ...net.arcs,
        ...arc.map(a => {
            if (!a.from) {
              throw ArcFromNotDefined;
            }
            if (!a.to) {
              throw ArcToNotDefined;
            }
            if (Net.sameKind(a.from, a.to)) {
              throw ArcBetweenSameType;
            }
            return {
              ...a,
              id: createId()
            };
          }
        )
      ]
    };
  }

  isUniqueName(net: Net.PetriNet, name: string): boolean {
    return !net.places.some(p => p.name === name) && !net.transitions.some(t => t.name === name) && !net.children.some(c => c.name === name);
  }

  addChild(net: Net.PetriNet, ...children: Net.PetriNet[]): Net.PetriNet {
    return {
      ...net,
      children: [
        ...net.children,
        ...children
      ]
    };
  }

  addPlace(net: Net.PetriNet, ...place: Omit<Net.Place, "id">[]): Net.PetriNet {
    return {
      ...net,
      places: [
        ...net.places,
        ...place.map(p => ({
            ...p,
            id: createId()
          })
        )
      ]
    };
  }

  addTransition(net: Net.PetriNet, ...transition: Omit<Net.Transition, "id">[]): Net.PetriNet {
    return {
      ...net,
      transitions: [
        ...net.transitions,
        ...transition.map(t => ({
            ...t,
            id: createId()
          })
        )
      ]
    };
  }

  incrementNewName(net: Net.PetriNet, name: string): string {
    let i = 1;
    while (!this.isUniqueName(net, `${name} (${i})`)) {
      i++;
    }
    return `${name} (${i})`;
  }

  copy(net: Net.PetriNet): Net.PetriNet {
    return {
      id: createId(),
      name: this.incrementNewName(net, net.name),
      places: net.places.map(p => ({
        ...p,
        id: createId()
      })),
      transitions: net.transitions.map(t => ({
        ...t,
        id: createId()
      })),
      arcs: net.arcs.map(a => ({
        ...a,
        id: createId()
      })),
      children: net.children.map(c => this.copy(c))
    };
  }

  newNet(name: string): Net.PetriNet {
    return {
      id: createId(),
      name,
      places: [],
      transitions: [],
      arcs: [],
      children: []
    };
  }

  removeArc(net: Net.PetriNet, ...arc: Pick<Net.Arc, "id">[]): Net.PetriNet {
    return {
      ...net,
      arcs: net.arcs.filter(a => !arc.some(r => r.id === a.id))
    };
  }

  removeChild(net: Net.PetriNet, ...children: Pick<Net.PetriNet, "id">[]): Net.PetriNet {
    return {
      ...net,
      children: net.children.filter(c => !children.some(r => r.id === c.id))
    };
  }

  removePlace(net: Net.PetriNet, ...place: (Pick<Net.Place, "name"> | Pick<Net.Place, "id">)[]): Net.PetriNet {
    return {
      ...net,
      places: net.places.filter(p => !place.some(r => {
        if ("id" in r) {
          return r.id === p.id;
        }
        return r.name === p.name;
      }))
    };
  }

  removeTransition(net: Net.PetriNet, ...transition: (Pick<Net.Transition, "name"> | Pick<Net.Transition, "id">)[]): Net.PetriNet {
    return {
      ...net,
      transitions: net.transitions.filter(t => !transition.some(r => {
        if ("id" in r) {
          return r.id === t.id;
        }
        return r.name === t.name;
      }))
    };
  }

  rename(net: Net.PetriNet, name: string): Net.PetriNet {
    return {
      ...net,
      name
    };
  }

  updateArc(net: Net.PetriNet, arc: Net.Arc): Net.PetriNet {
    if (!arc.id) {
      throw new Error("Arc id is not defined");
    }
    const arcToUpdate = net.arcs.find(a => a.id === arc.id);
    if (!arcToUpdate) {
      throw new Error("Arc not found");
    }
    if (arc.from) {
      arcToUpdate.from = arc.from;
    }
    if (arc.to) {
      arcToUpdate.to = arc.to;
    }

    return {
      ...net,
      arcs: [
        ...net.arcs.filter(a => a.id !== arc.id),
        arcToUpdate
      ]
    };
  }

  updatePlace(net: Net.PetriNet, place: Net.Place): Net.PetriNet {
    if (!place.id) {
      throw new Error("Place id is not defined");
    }
    const placeToUpdate = net.places.find(p => p.id === place.id);
    if (!placeToUpdate) {
      throw new Error("Place not found");
    }
    return {
      ...net,
      places: [
        ...net.places.filter(p => p.id !== place.id),
        place
      ]
    };
  }

  updateTransition(net: Net.PetriNet, transition: Net.Transition): Net.PetriNet {
    if (!transition.id) {
      throw new Error("Transition id is not defined");
    }
    const transitionToUpdate = net.transitions.find(t => t.id === transition.id);
    if (!transitionToUpdate) {
      throw new Error("Transition not found");
    }
    return {
      ...net,
      transitions: [
        ...net.transitions.filter(t => t.id !== transition.id),
        transition
      ]
    };
  }
}

export const builder = new Builder();