"use client";

import { ReactNode, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import Loading from "@/app/loading";
import { Button } from "@/components/ui/button";
import Link from "next/link";

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

  return (
    <div className="flex gap-x-4 min-h-screen">
      <div className="flex flex-col gap-2 p-8 border-r shrink-0">
        <h1 className="text-3xl pb-5 text-center border-b">wg-hub</h1>
        <Link className="bg-secondary rounded-md p-1" href={"/peers"}>
          Peers
        </Link>
        <div className="flex-grow"></div>
        <Button onClick={() => auth.logout()}>Logout</Button>
        <span>Version: {process.env.VERSION}</span>
      </div>
      <div className="flex-grow p-8">{children}</div>
    </div>
  );
}
