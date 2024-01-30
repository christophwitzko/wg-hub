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

  //TODO: layout
  return (
    <div className="flex gap-x-4 min-h-screen">
      <div className="flex flex-col gap-2 border border-red-600 p-5">
        <h1 className="text-xl pb-5 text-center">wg-hub</h1>
        <Link href={"/peers"}>Peers</Link>
        <Button onClick={() => auth.logout()}>Logout</Button>
        <div className="flex-grow"></div>
        <span>User: {auth.username}</span>
        <span>Version: xxxxxxxx</span>
      </div>
      <div className="border border-green-600 flex-grow p-5">{children}</div>
    </div>
  );
}
