import type { Net } from "~/models/net";
import { createId } from "@paralleldrive/cuid2";


class EventService implements Net.EventService {
  private controller: Net.Controller;
  private handlers: Map<string, Net.Handler<any>[]>;

  constructor(controller: Net.Controller) {
    this.controller = controller;
    this.handlers = new Map<string, Net.Handler<any>[]>();
  }

  availableEvents(net: Net.PetriNet): Net.Event[] {
    return this.controller.availableTransitions(net).flatMap(t => t.events);
  }

  registerHandler<S>(net: Net.PetriNet, event: Pick<Net.Event, "id">, handler: Net.Handler<S>): Net.PetriNet {
    const handlers = this.handlers.get(event.id);
    if (!handlers) {
      this.handlers.set(event.id, [handler]);
      return net;
    }
    this.handlers.set(event.id, [...handlers, handler]);
    return net;
  }

  makeField(name: string, type: Net.Type, required: boolean): Net.Field {
    return {
      id: createId(),
      name,
      type,
      required
    };
  }

  makeEvent(name: string, input: Net.EventInput): Net.Event {
    return {
      id: createId(),
      name,
      fields: input.fields.map(f => this.makeField(f.name, f.type, f.required))
    };
  }

  registerEvent(net: Net.PetriNet, transition: Pick<Net.Transition, "id">, event: Net.EventInput): Net.PetriNet {
    const transitionIndex = net.transitions.findIndex(t => t.id === transition.id);
    if (transitionIndex === -1) {
      throw new Error("Transition not found");
    }
    const transitionCopy = {
      ...net.transitions[transitionIndex],
      events: [
        ...net.transitions[transitionIndex].events,
        this.makeEvent(event.name, event)
      ]
    };
    return {
      ...net,
      transitions: [
        ...net.transitions.slice(0, transitionIndex),
        transitionCopy,
        ...net.transitions.slice(transitionIndex + 1)
      ]
    };
  }

  canHandleEvent(net: Net.PetriNet, eventId: Pick<Net.Event, "id">): boolean {
    const transition = this.controller.availableTransitions(net).filter(t => t.events.some(e => e.id === eventId.id));
    return transition.length > 0;
  }

  handleEvent<S>(net: Net.PetriNet, eventId: Pick<Net.Event, "id">, payload: S): Net.PetriNet {
    if (!this.canHandleEvent(net, eventId)) {
      throw new Error("Cannot handle event");
    }
    const transition = this.controller.availableTransitions(net).filter(t => t.events.some(e => e.id === eventId.id))[0];
    const handlers = this.handlers.get(eventId.id);
    if (!handlers) {
      return this.controller.fireTransition(net, transition);
    }
    let netCopy = {
      ...net
    };
    handlers.forEach(h => {
      netCopy = h(netCopy, payload);
    });
    return this.controller.fireTransition(netCopy, transition);
  }

  isValidEventSequence(net: Net.PetriNet, sequence: Pick<Net.Event, "id">[]): boolean {
    try {
      let netCopy = {
        ...net
      };
      sequence.forEach(e => {
        netCopy = this.handleEvent(netCopy, e, {});
      });

    } catch (e) {
      return false;
    }
    return true;
  }
}

export default EventService;