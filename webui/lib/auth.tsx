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

type AuthContextValue = {
  token: string;
  isInitialized: boolean;
  isLoading: boolean;
  error: string;
  login: (username: string, password: string) => void;
};

export const AuthContext = createContext({} as AuthContextValue);

async function apiLogin(username: string, password: string) {
  const res = await fetch("/api/auth", {
    method: "POST",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) {
    throw new Error("Invalid username or password.");
  }
  const body = await res.json();
  if (!body.token) {
    throw new Error("Invalid response from server.");
  }
  return body;
}
export function AuthProvider({ children }: { children: ReactNode }) {
  const isSSR = useSyncExternalStore(
    () => {
      return () => {};
    },
    () => false,
    () => true,
  );
  const [token, setToken] = useLocalStorageState("authToken", {
    defaultValue: "",
  });
  const [isInitialized, setIsInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const login = useCallback((username: string, password: string) => {
    setIsLoading(true);
    setToken("");
    setError("");
    apiLogin(username, password)
      .then((res) => {
        setToken(res.token);
      })
      .catch((err) => {
        setError(err.message);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);
  useEffect(() => {
    if (token) {
      // TODO verify token
    }
    if (!isSSR) {
      setIsInitialized(true);
    }
  }, [token, isSSR]);
  const contextValue = useMemo(() => {
    return {
      token,
      isInitialized,
      isLoading,
      error,
      login,
    } as AuthContextValue;
  }, [token, isLoading, isInitialized, error, login]);

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
