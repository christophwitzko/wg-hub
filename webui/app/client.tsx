"use client";

import { useAuth } from "@/lib/auth";
import { Login } from "@/components/login";
import { useRouter } from "next/navigation";
import Loading from "@/app/loading";
import { useEffect } from "react";

export function Client() {
  const router = useRouter();
  const auth = useAuth();
  useEffect(() => {
    if (auth.token) {
      router.push("/peers");
    }
  }, [auth.token, router]);

  if (auth.isInitialized && !auth.token) {
    router.prefetch("/peers");
    return <Login />;
  }

  return <Loading />;
}
