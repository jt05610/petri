export type CodeError = {
  message: string;
  line: number;
  column: number;
  severity: "error" | "warning";
  source: string;
}
