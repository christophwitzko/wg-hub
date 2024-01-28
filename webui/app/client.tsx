"use client";

import { useAuth } from "@/lib/auth";
import { Login } from "@/components/login";
import { useRouter } from "next/navigation";
import Loading from "@/app/loading";

export function Client() {
  const router = useRouter();
  const auth = useAuth();
  if (!auth.isInitialized) {
    return <Loading />;
  }
  if (!auth.token) {
    return <Login />;
  }
  router.push("/peers");
  return null;
}
