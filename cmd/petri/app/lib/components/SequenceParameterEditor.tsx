import { useParserContext } from "~/lib/context/ParserContext";

export default function SequenceParameterEditor() {
  const { state } = useParserContext();
  const { scope } = state;

  return (
    <table className={"table table-auto"}>
      <thead>
      <tr>
        <th>Parameter</th>
        <th>Value</th>
      </tr>
      </thead>
      <tbody>
      {Object.entries(scope).map(([key, value], index) => {
        return (
          <tr key={index}>
            <td>{key}</td>
            <td>{value}</td>
          </tr>
        );
      })}
      </tbody>
    </table>
  );
}