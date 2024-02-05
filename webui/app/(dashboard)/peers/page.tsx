"use client";
import { useAuth } from "@/lib/auth";
import { usePeers } from "@/lib/api";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { PeersTable } from "@/app/(dashboard)/peers/table";

export default function Peers() {
  const auth = useAuth();
  const { data, error, isValidating } = usePeers(auth.token);
  return (
    <div>
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="size-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error.toString()}</AlertDescription>
        </Alert>
      )}
      <PeersTable data={data!} isLoading={isValidating} />
    </div>
  );
}
