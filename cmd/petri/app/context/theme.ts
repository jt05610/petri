import colors from "tailwindcss/colors";

export const DefaultTheme = {
  name: "Default",
  colors: {
    background: {
      lightMode: colors.white,
      darkMode: colors.gray[800]
    },
    text: {
      lightMode: colors.gray[800],
      darkMode: colors.white
    },
    primary: {
      lightMode: colors.blue[500],
    }
  }
};
