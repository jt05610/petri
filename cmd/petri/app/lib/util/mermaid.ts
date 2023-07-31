import mermaid from "mermaid";
import type { MermaidConfig, RenderResult } from "mermaid";

export async function render(
  config: MermaidConfig,
  code: string,
  id: string,
): Promise<RenderResult> {
  // Should be able to call this multiple times without any issues.
  mermaid.initialize(config);
  return mermaid.render(id, code);
}

export const parse = async (code: string): Promise<unknown> => {
  return await mermaid.parse(code);
};