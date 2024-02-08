"use client";

import { ReactNode, useEffect } from "react";
import { FileCog, LogOut, Network, Sun, Moon } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { useTheme } from "next-themes";

import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import Loading from "@/app/loading";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const navItems = [
  {
    href: "/peers",
    title: "Peers",
    icon: Network,
  },
  {
    href: "/config",
    title: "Config",
    icon: FileCog,
  },
] as const;

export default function Layout({ children }: { children: ReactNode }) {
  const { setTheme } = useTheme();
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
      <nav className="flex flex-col gap-2 p-2 border-r shrink-0">
        <h1 className="text-3xl pt-4 pb-6 text-center font-bold border-b mb-6">
          wg-hub
        </h1>
        {navItems.map((item) => {
          const selected = pathname === item.href;
          return (
            <Link
              key={item.href}
              className={cn(
                buttonVariants({
                  variant: selected ? "default" : "ghost",
                  size: "sm",
                }),
                selected &&
                  "dark:bg-muted dark:text-white dark:hover:bg-muted dark:hover:text-white",
                "justify-start w-44",
              )}
              href={item.href}
            >
              <item.icon className="mr-2 size-4" />
              {item.title}
            </Link>
          );
        })}
        <div className="flex-grow"></div>
        <Button onClick={() => auth.logout()} variant="ghost">
          <LogOut className="mr-2 size-4" />
          Logout
        </Button>
        <div className="flex justify-between gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon">
                <Sun className="size-5 scale-100 dark:scale-0" />
                <Moon className="absolute size-5 scale-0 dark:scale-100" />
                <span className="sr-only">Toggle theme</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setTheme("light")}>
                Light
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme("dark")}>
                Dark
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme("system")}>
                System
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <div className="flex justify-center">
            <span className="text-primary/40 text-center my-auto">
              v{process.env.VERSION}
            </span>
          </div>
        </div>
      </nav>
      <div className="flex-grow p-8">{children}</div>
    </div>
  );
}
