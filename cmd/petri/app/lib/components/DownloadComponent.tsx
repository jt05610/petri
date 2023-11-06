import { useContextSelector } from "use-context-selector";
import { RunContext } from "~/lib/context/run";
import { useLoaderData } from "@remix-run/react";
import type { RunDetails } from "~/models/net.run";
import { getParameterRecord } from "~/util/parameters";
import JSZip from "jszip";
import type { loader } from "~/routes/control.$netID.sequences.$sequenceID._index";
import type { DeviceInstanceInput } from "~/models/__generated__/graphql";

function getRunDevices(run: RunDetails): DeviceInstanceInput[] {
  return run.steps.map(({ action }) => {
    return {
      deviceID: action.device.id,
      instanceID: ""
    };
  }).filter((device, index, self) => {
    return self.findIndex((d) => d.deviceID === device.deviceID) === index;
  });
}

export default function DownloadComponent() {
  const sequence = useContextSelector(RunContext, (context) => context?.run);
  const user = useLoaderData<typeof loader>().user;

  function makeSessionRequest(d: RunDetails) {
    return {
      sequenceID: d.id,
      userID: user.id,
      instances: getRunDevices(d)
    };
  }

  function makeStartRequest(d: RunDetails) {
    return {
      sessionID: "",
      parameters: getParameterRecord(d)
    };
  }

  async function handleDownload() {
    if (!sequence) return;
    let zip = new JSZip();
    const sessionReq = makeSessionRequest(sequence);
    zip.file(".gitignore", "data");
    zip.file("data/start.json", JSON.stringify(makeStartRequest(sequence), null, 2));
    zip.file("data/run.json", JSON.stringify(sessionReq, null, 2));
    const blob = await zip.generateAsync({ type: "blob" });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${sequence.name}.zip`;
    link.click();
  }

  return (
    <button
      type={"button"}
      className={"px-2 py-1 text-white hover:text-sky-500 dark:hover:text-sky-400 bg-gradient-to-b from-sky-700 to-sky-800 rounded-full"}
      onClick={handleDownload}
    >
      Download
    </button>

  );
}
