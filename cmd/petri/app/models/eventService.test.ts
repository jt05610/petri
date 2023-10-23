import { expect } from "vitest";
import EventService from "~/models/eventService";
import { controller } from "~/models/controller";
import { isCuid } from "@paralleldrive/cuid2";
import { makeNet } from "~/models/controller.test";
import { Net } from "~/models/net";

const eventService = new EventService(controller);

function log<S>(net: Net.PetriNet, payload: S): Net.PetriNet {
  console.log(payload);
  return net;
}

describe.concurrent("EventService", async () => {
  it("should register an event", async () => {
    const net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    const netCopy = eventService.registerEvent(net, transition, event);
    const [eventCopy] = netCopy.transitions[0].events;
    expect(eventCopy.name).toBe(event.name);
    expect(isCuid(eventCopy.id)).toBe(true);
  });

  it("should handle an event", async () => {
    let net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    net = eventService.registerEvent(net, transition, event);
    const [eventCopy] = net.transitions[0].events;
    let seenValue = "";
    net = eventService.registerHandler(net, eventCopy, (net, payload: { f1: string }) => {
      seenValue = payload.f1;
      return net;
    });
    net = eventService.handleEvent(net, eventCopy, { f1: "test" });
    expect(net.places[1].marking).toBe(1);
    expect(seenValue).toBe("test");
  });
  it("should register a handler", async () => {
    const net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    const netCopy = eventService.registerEvent(net, transition, event);
    const [eventCopy] = netCopy.transitions[0].events;
    const netCopy2 = eventService.registerHandler(netCopy, eventCopy, log);
    expect(netCopy2).toBe(netCopy);
  });
  it("should return available events", async () => {
    const net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    const netCopy = eventService.registerEvent(net, transition, event);
    const availableEvents = eventService.availableEvents(netCopy);
    expect(availableEvents).toHaveLength(1);
    expect(availableEvents[0].name).toBe(event.name);
  });
  it("should return true if event can be handled", async () => {
    const net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    const netCopy = eventService.registerEvent(net, transition, event);
    const [eventCopy] = netCopy.transitions[0].events;
    const canHandle = eventService.canHandleEvent(netCopy, eventCopy);
    expect(canHandle).toBe(true);
  });
  it("should return false if event can't be handled", async () => {
    const net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    const netCopy = eventService.registerEvent(net, transition, event);
    const [eventCopy] = netCopy.transitions[0].events;
    const canHandle = eventService.canHandleEvent(netCopy, eventCopy);
    expect(canHandle).toBe(true);
  });
  it("should return false if event sequence is invalid", async () => {
    const net = makeNet();
    const transition = net.transitions[0];
    if (!transition) {
      throw new Error("Transition not found");
    }
    const event = {
      name: "e1",
      fields: [{
        name: "f1",
        type: Net.String,
        required: true
      }]
    };
    const netCopy = eventService.registerEvent(net, transition, event);
    const [eventCopy] = netCopy.transitions[0].events;
    const isValid = eventService.isValidEventSequence(netCopy, [eventCopy, eventCopy]);
    expect(isValid).toBe(false);
  });
});
