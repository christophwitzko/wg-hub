"use client";

import { ReactNode, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import Loading from "@/app/loading";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { cn } from "@/lib/utils";

export default function Layout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
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
        <Link
          className={cn(
            "font-medium rounded-md p-2 border-background border hover:border-border",
            pathname === "/peers" && "bg-primary/80 text-primary-foreground",
          )}
          href={"/peers"}
        >
          Peers
        </Link>
        <Link
          className={cn(
            "font-medium rounded-md p-2 border-background border hover:border-border",
            pathname === "/config" && "bg-primary/80 text-primary-foreground",
          )}
          href={"/config"}
        >
          Config
        </Link>
        <div className="flex-grow"></div>
        <Button onClick={() => auth.logout()}>Logout</Button>
        <span>Version: {process.env.VERSION}</span>
      </div>
      <div className="flex-grow p-8">{children}</div>
    </div>
  );
}
