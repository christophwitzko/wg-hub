"use client";

import { ReactNode, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import Loading from "@/app/loading";

export default function Layout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const auth = useAuth();
  useEffect(() => {
    if (auth.isInitialized && !auth.token) {
      router.push("/");
    }
  }, [auth.isInitialized, auth.token, router]);

  if (!auth.isInitialized || !auth.token) {
    return <Loading />;
  }

  //TODO: layout
  return <div>{children}</div>;
}
