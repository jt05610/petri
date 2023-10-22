import RunEditor from "~/lib/components/DesignerEditor";
import { RunEditorProvider } from "~/lib/components/DesignerContext";

export default function SequenceEditPage() {
  return (
    <RunEditorProvider>
      <RunEditor />
    </RunEditorProvider>
  );
}
