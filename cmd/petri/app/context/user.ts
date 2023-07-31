import { createContext } from "react";

const DefaultUser = {
  id: "",
  email: "",
  role: "NONE",
};

export const UserContext = createContext(DefaultUser);