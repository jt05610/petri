import { type Monaco, Editor } from "@monaco-editor/react";
import { useContextSelector } from "use-context-selector";
import { EditorChangeAction, RunEditorContext } from "~/lib/components/DesignerContext";
import { ActivePetriNet, PetriNetContext } from "~/lib/context/petrinet";
import Writer, { pascalCase } from "../writers/language/typescript/interfaceWriter";
import { Component, Suspense, useState } from "react";
import { editor } from "monaco-editor";

const writer = new Writer();

const DEST_DIR = (userID: string) => `public/${userID}/@types`;

type File = {
  content: string
  filePath: string
}

function writeInterfaceTypeFile(userID: string, petriNet: ActivePetriNet): File {
  const str = writer.writePetriNet(petriNet.net);
  const path = `${DEST_DIR(userID)}/index.d.ts`;
  return {
    content: str,
    filePath: path
  };
}


class RunButton extends Component {
  handleClick() {
    console.log("run");
  }

  render() {
    return (
      <button onClick={this.handleClick}>
        Run
      </button>
    );
  }
}

export default function RunEditor() {
  const { state, dispatch } = useContextSelector(RunEditorContext, (context) => context!);
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const [runResult, setRunResult] = useState<string | undefined>(undefined);

  function run() {
    setRunResult("Running...");
  }

  function handleEditorWillMount(monaco: Monaco) {
    writeInterfaceTypeFile(petriNet?.userID!, petriNet!);
    const { content, filePath } = writeInterfaceTypeFile(petriNet?.userID!, petriNet!);

    monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
      target: monaco.languages.typescript.ScriptTarget.ES2016,
      allowNonTsExtensions: true,
      moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
      module: monaco.languages.typescript.ModuleKind.CommonJS,
      noEmit: true,
      typeRoots: ["file:///node_modules/@types"]
    });
    monaco.languages.typescript.typescriptDefaults.addExtraLib(
      content,
      `file:///node_modules/@types/system/index.d.ts`
    );
  }


  function handleEditorDidMount(editor: editor.IStandaloneCodeEditor, monaco: Monaco) {
    const systemName = pascalCase(petriNet!.net.net.children[0].name);
    const model = monaco.editor.createModel(
      `import { ${systemName} } from 'system';\n\nfunction run(system: ${systemName}) {\n\n}\n`,
      "typescript",
      monaco.Uri.parse("file:///main.ts")
    );
    editor.setModel(model);
    editor.addAction({
      id: "run",
      label: "Run",
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
      contextMenuGroupId: "navigation",
      contextMenuOrder: 1.5,
      run: function(ed, ...args) {
        run();
      }
    });
    // Create an empty div

    editor.addOverlayWidget({
      getId: () => "run",
      getDomNode: () => {
        const node = document.createElement("button");
        node.innerText = "Run";
        node.className = "bg-blue-500 hover:bg-blue-700 text-white font-bold py-1 px-4 mt-2 rounded-full";
        node.onclick = () => {
          node.className = "bg-green-500 hover:bg-green-700 text-white font-bold py-1 px-4 mt-2 rounded-full";
          run();
        };
        return node;
      },
      getPosition: () => ({
        preference: monaco.editor.OverlayWidgetPositionPreference.TOP_RIGHT_CORNER
      })
    });
  }

  function handleEditorChange(value: string | undefined) {
    dispatch(new EditorChangeAction(value!));
  }

  return (
    <Suspense fallback={<div>Loading...</div>}>
      <div className={"overlay rounded-md overflow-hidden w-full h-full shadow-2xl"}>
        <Editor
          height={state.height}
          width={state.width}
          language={"typescript"}
          value={state.code}
          theme={"vs-dark"}
          onChange={handleEditorChange}
          onMount={handleEditorDidMount}
          beforeMount={handleEditorWillMount}
          options={{
            minimap: {
              enabled: false
            }
          }}
        />
      </div>
    </Suspense>
  );
}
