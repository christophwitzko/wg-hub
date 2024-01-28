"use client";

import { ReactNode } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import Loading from "@/app/loading";

export default function Layout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const auth = useAuth();
  if (!auth.isInitialized) {
    return <Loading />;
  }
  if (!auth.token) {
    router.push("/");
    return null;
  }
  //TODO: layout
  return <div>{children}</div>;
}
