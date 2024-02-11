"use client";
import { useAuth } from "@/lib/auth";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfig } from "@/lib/api";
import { Code } from "@/components/code";

export default function Peers() {
  const auth = useAuth();
  const { data, error } = useConfig(auth.token);
  const config = data ? data.config : "";
  return (
    <div className="flex flex-col gap-4 h-full">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="size-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
        </Alert>
      )}
      <h1 className="text-2xl">wireguard-hub.yaml</h1>
      <Code className="flex-grow" lang="yaml" value={config} />
      <Button
        className="ml-auto"
        onClick={() => navigator.clipboard.writeText(config)}
      >
        Copy to clipboard
      </Button>
    </div>
  );
}
