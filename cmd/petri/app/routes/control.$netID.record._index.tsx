import {  RecordRunProvider } from "~/context";
import { SystemControl } from "~/lib/components/systemControl";

export default function ControlIndex() {
  return (
    <RecordRunProvider>
      <SystemControl />
    </RecordRunProvider>
  );
};