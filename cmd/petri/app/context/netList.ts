import { createContext } from "react";
import type { NetListItem } from "~/models/net.server";

export const NetListContext= createContext<NetListItem[]>([]);