import { CodeError } from "~/lib/components/DesignerError";
import { createDataContext } from "~/lib/context/Context";

export enum SequenceDesignActionType {
  EditorChange = "EditorChange",
}

type ActionHandler = (state: EditorState, payload: any) => EditorState;

interface SequenceDesignAction {
  type: SequenceDesignActionType;
  handler: ActionHandler;
  payload: any;
}

export class EditorChangeAction implements SequenceDesignAction {
  type = SequenceDesignActionType.EditorChange;

  payload: any;

  constructor(payload: string) {
    this.payload = payload;
  }

  handler(state: EditorState, payload: string): EditorState {
    return {
      ...state,
      code: payload
    };
  }
}

enum ImplementedLanguage {
  Python = "python",
}

interface EditorState {
  code: string;
  height?: string;
  width?: string;
  theme?: string;
  errors: CodeError[];
  language: ImplementedLanguage;
}

const initialState: EditorState = {
  code: "",
  errors: [],
  height: "100vh",
  width: "100%",
  theme: "vs-dark",
  language: ImplementedLanguage.Python
};

function reducer(state: EditorState, action: SequenceDesignAction): EditorState {
  switch (action.type) {
    case SequenceDesignActionType.EditorChange:
      return action.handler(state, action.payload);
    default:
      return state;
  }
}

export const [RunEditorContext, RunEditorProvider] = createDataContext(initialState, reducer);