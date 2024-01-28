"use client";

import { useAuth } from "@/lib/auth";
import { Login } from "@/components/login";

export function Client() {
  const auth = useAuth();
  if (!auth.isInitialized) {
    return <div>Loading...</div>;
  }
  if (!auth.token) {
    return <Login />;
  }
  return <div>Logged in!</div>;
}
