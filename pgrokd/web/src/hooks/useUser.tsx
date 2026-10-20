import { createContext, useContext } from "react";

export interface UserContextType {
  authed: boolean;
  displayName: string;
  token: string;
  url: string;
  subdomain: string;
  domain: string;
  scheme: string;
  remoteAddr: string;
  tcpHost: string;
  tcpPortStart: number;
  tcpPortEnd: number;
}

const UserContext = createContext<UserContextType>({} as UserContextType);

export function UserProvider({ user, children }: { user: UserContextType; children: JSX.Element }) {
  return <UserContext.Provider value={user}>{children}</UserContext.Provider>;
}

export default function useUser() {
  return useContext(UserContext);
}
