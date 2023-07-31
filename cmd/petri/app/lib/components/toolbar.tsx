import type { ReactNode } from "react";


type ToolbarSectionProps = {
  title: string;
  children: ReactNode;
}

export function ToolbarSection({ title, children }: ToolbarSectionProps) {
  return (
    <div className="flex flex-col gap-2">
      <h2 className="font-bold text-lg">{title}</h2>
      {children}
    </div>
  );
}

type ToolbarProps = {
  children: ReactNode;
}

export default function Toolbar({ children }: ToolbarProps) {
  return (
    <div className="flex flex-col gap-4">
      {children}
    </div>
  );
}

