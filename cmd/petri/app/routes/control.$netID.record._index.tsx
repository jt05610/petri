import {  RecordRunProvider } from "~/context";
import { SystemControl } from "~/lib/components/systemControl";

export default function ControlIndex() {
  return (
    <RecordRunProvider>
      <div className={"max-w-screen"}>
        <SystemControl />
      </div>
    </RecordRunProvider>
  );
};