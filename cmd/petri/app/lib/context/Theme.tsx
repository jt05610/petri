import type { Dispatch, ReactNode, Reducer } from "react";
import { useReducer } from "react";
import { createContext } from "use-context-selector";

interface Color {
  bg?: string;
  fg?: string;
  text?: string;
  border?: string;
  ring?: string;
}

interface Theme {
  primary: Color;
  secondary: Color;
  accent: Color;
  background: Color;
  text: Color;
}


interface ThemeAction {
  handle: (state: Theme) => Theme;
}

function themeReducer(state: Theme, action: ThemeAction): Theme {
  return action.handle(state);
}

const defaultTheme: Theme = {
  primary: {
    bg: "bg-slate-300 dark:bg-slate-900",
    fg: "bg-slate-400 dark:bg-slate-800",
    text: "text-slate-900 dark:text-slate-400",
    border: "border-slate-400 dark:border-slate-800",
    ring: "ring-slate-400 dark:ring-slate-800"
  },
  secondary: {
    bg: "bg-teal-400 dark:bg-teal-700",
    fg: "bg-teal-500 dark:bg-teal-600",
    text: "text-teal-900 dark:text-teal-400",
    border: "border-teal-500 dark:border-teal-600",
    ring: "ring-teal-500 dark:ring-teal-600"
  },
  accent: {
    bg: "bg-amber-400 dark:bg-amber-700",
    fg: "bg-amber-500 dark:bg-amber-600",
    text: "text-amber-900 dark:text-amber-400",
    border: "border-amber-500 dark:border-amber-600",
    ring: "ring-amber-500 dark:ring-amber-600"
  },
  background: {
    bg: "bg-white dark:bg-slate-900",
    fg: "bg-slate-100 dark:bg-slate-800",
    text: "text-slate-900 dark:text-slate-400",
    border: "border-slate-100 dark:border-slate-800",
    ring: "ring-slate-100 dark:ring-slate-800"
  },
  text: {
    text: "text-slate-900 dark:text-slate-400"
  }
};

type ThemeContextType = {
  theme: Theme;
  dispatch: Dispatch<ThemeAction>;
}
export const ThemeContext = createContext<ThemeContextType | null>(null);

export const ThemeProvider = ({ theme, children }: {
  theme?: Theme,
  children: ReactNode
}) => {
  const [themeState, dispatch] = useReducer<Reducer<Theme, ThemeAction>>(themeReducer, { theme: theme || defaultTheme });
  return (
    <ThemeContext.Provider value={{ theme: themeState, dispatch }}>
      {children}
    </ThemeContext.Provider>
  );
};

export const updateTheme = (theme: Theme): ThemeAction => {
  return {
    handle: (state: Theme) => {
      return {
        ...state,
        ...theme
      };
    }
  };
};