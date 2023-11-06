import { useParserContext } from "~/lib/context/ParserContext";
import { Suspense } from "react";

export default function ParameterEditor() {
  const { state } = useParserContext();
  const { scope } = state;

  return (
    <div className={"flex h-1/4 w-full flex-col space-x-2 p-2"}>
      <Suspense fallback={<div>Loading...</div>}>
        {scope && (
          <table className={"table h-full border-2 border-gray-900 text-left rounded-xl p-2 max-w-lg"}>
            <thead>
            <tr>
              <th className={"text-xl font-black"}>Parameter</th>
              <th className={"text-xl font-black"}>Value</th>
            </tr>
            </thead>
            <tbody>
            {Array.from(scope.entries()).map(([key, value], index) => {
                return (
                  <tr key={index}>
                    <td>{key}</td>
                    <td>{value}</td>
                  </tr>
                );
              }
            )}
            </tbody>
          </table>
        )
        }
      </Suspense>
    </div>
  );
}