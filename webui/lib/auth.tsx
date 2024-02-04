"use client";

import {
  createContext,
  useContext,
  useState,
  ReactNode,
  useMemo,
  useCallback,
  useEffect,
  useSyncExternalStore,
} from "react";
import useLocalStorageState from "use-local-storage-state";
import { createToken, getUser } from "@/lib/api";

export type AuthContextType = {
  token: string;
  isInitialized: boolean;
  isLoading: boolean;
  error: string;
  login: (username: string, password: string) => void;
  logout: () => void;
  username: string;
};

export const AuthContext = createContext({} as AuthContextType);

export function AuthProvider({ children }: { children: ReactNode }) {
  const isSSR = useSyncExternalStore(
    () => {
      return () => {};
    },
    () => false,
    () => true,
  );
  const [token, setToken, { removeItem }] = useLocalStorageState("authToken", {
    defaultValue: "",
  });
  const [isInitialized, setIsInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [username, setUsername] = useState("");

  const login = useCallback((username: string, password: string) => {
    setIsLoading(true);
    setToken("");
    setError("");
    createToken(username, password)
      .then((token) => {
        setToken(token);
      })
      .catch((err) => {
        setError(err.message);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  const logout = useCallback(() => {
    removeItem();
  }, []);

  useEffect(() => {
    if (isSSR) {
      return;
    }
    if (!token) {
      setIsInitialized(true);
      return;
    }
    getUser(token)
      .then((user) => {
        setUsername(user.username);
      })
      .catch((err) => {
        logout();
      })
      .finally(() => {
        setIsInitialized(true);
      });
  }, [token, isSSR, logout]);

  const contextValue = useMemo(() => {
    return {
      token,
      isInitialized,
      isLoading,
      error,
      login,
      logout,
      username,
    } as AuthContextType;
  }, [token, isLoading, isInitialized, error, login, logout, username]);

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
