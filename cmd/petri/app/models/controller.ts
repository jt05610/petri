import { Net } from "~/models/net";


class Controller implements Net.Controller {
  availableTransitions(net: Net.PetriNet): Net.Transition[] {
    const ret: Net.Transition[] = [];

    for (const transition of net.transitions) {
      if (this.transitionIsHot(net, transition)) {
        ret.push(transition);
      }
    }
    return ret;

  }

  transitionIsHot(net: Net.PetriNet, transitionId: Pick<Net.Transition, "id">): boolean {
    const inputPlaces = this.inputs(net, transitionId) as Net.Place[];
    if (!inputPlaces) {
      return false;
    }
    return inputPlaces.every(p => p.marking > 0);
  }

  fireTransition(net: Net.PetriNet, transitionId: Pick<Net.Transition, "id">): Net.PetriNet {
    if (!this.transitionIsHot(net, transitionId)) {
      throw new Error("Transition is not hot");
    }
    const inputPlaces = this.inputs(net, transitionId) as Net.Place[];
    const outputPlaces = this.outputs(net, transitionId) as Net.Place[];

    const updatedInputs = inputPlaces.map(p => {
      return {
        ...p,
        marking: p.marking - 1
      };
    });

    const updatedOutputs = outputPlaces.map(p => {
      return {
        ...p,
        marking: p.marking + 1
      };
    });
    const updatedPlaces = updatedInputs.concat(updatedOutputs);

    return {
      ...net,
      places: net.places.map(p => {
          const updatedPlace = updatedPlaces.find(up => up.id === p.id);
          if (!updatedPlace) {
            return p;
          }
          return updatedPlace;
        }
      )
    };
  }

  inputs(net: Net.PetriNet, node: Pick<Net.Place, "id"> | Pick<Net.Transition, "id">): Net.Place[] | Net.Transition[] {
    const nodeArcs = net.arcs.filter(a => a.to.id === node.id);
    if (Net.isPlace(node)) {
      return net.transitions.filter(p => nodeArcs.some(a => a.from.id === p.id));
    }
    return net.places.filter(p => nodeArcs.some(a => a.from.id === p.id));
  }

  marking(net: Net.PetriNet): Map<string, number> {
    const ret = new Map<string, number>();
    net.places.forEach(p => ret.set(p.id, p.initialMarking));
    return ret;
  }

  outputs(net: Net.PetriNet, node: Pick<Net.Place, "id"> | Pick<Net.Transition, "id">): Net.Place[] | Net.Transition[] {
    const nodeArcs = net.arcs.filter(a => a.from.id === node.id);
    if (Net.isPlace(node)) {
      return net.transitions.filter(p => nodeArcs.some(a => a.to.id === p.id));
    }
    return net.places.filter(p => nodeArcs.some(a => a.to.id === p.id));
  }

}

export const controller = new Controller();