interface ResultOrError<T> {
  result?: T;
  error?: string;
}

type PumpRequest = {
  flowRate: number;
}

type FinishRequest = {
  dispensedVolume: number;
  remainingVolume: number;
}

type PumpResponse = {
  flowRate: number;
}

type FinishResponse = {
  dispensedVolume: number;
  remainingVolume: number;
}

interface PumpActions {
  pump(params: PumpRequest): ResultOrError<PumpResponse>;

  finish(params: FinishRequest): ResultOrError<FinishResponse>;
}

interface PumpSettings {
  syringeDiameter: number;
  syringeLength?: number;
  syringeVolume?: number;
  syringeUnits?: string;
  maxFlowRate?: number;
  minFlowRate?: number;
}

interface Settings<S> {
  get settings(): S;

  set settings(settings: S);
}

export interface Pump extends PumpActions, Settings<PumpSettings> {

}