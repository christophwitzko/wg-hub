"use client";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";

export default function Peers() {
  const auth = useAuth();
  return (
    <div>
      <h1>Peers</h1>
      <Button onClick={() => auth.logout()}>Logout</Button>
    </div>
  );
}
