import { createContext, useContextSelector } from "use-context-selector";
import type { ReactNode } from "react";
import type { UserDetails } from "~/models/user";


type UserContextType = {
  user: UserDetails;
}

const UserContext = createContext<UserContextType | null>(null);

export const UserProvider = ({ user, children }: {
  user: UserDetails,
  children: ReactNode
}) => {
  return (
    <UserContext.Provider value={{ user }}>
      {children}
    </UserContext.Provider>
  );
};

export const useUser = () => {
  const context = useContextSelector(UserContext, (state) => state ? state.user : null);
  if (context === null) {
    throw new Error("useUser must be used within a UserProvider");
  }
  return context;
};