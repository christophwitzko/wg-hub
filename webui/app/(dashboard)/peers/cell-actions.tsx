"use client";

import { useCallback, useEffect, useState } from "react";
import { AlertCircle, MoreHorizontal } from "lucide-react";
import { Row } from "@tanstack/react-table";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Peer, removePeer } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function CellActions({ row }: { row: Row<Peer> }) {
  const peer = row.original;
  const auth = useAuth();
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const deletePeer = useCallback(() => {
    setError("");
    setIsLoading(true);
    removePeer(auth.token, peer.publicKey)
      .then(() => {
        setOpen(false);
      })
      .catch((error) => {
        setError(error.message);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [auth.token, peer.publicKey]);

  useEffect(() => {
    if (open) {
      setError("");
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="h-8 w-8 p-0">
            <span className="sr-only">Open menu</span>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>Actions</DropdownMenuLabel>
          <DropdownMenuItem
            onClick={() => navigator.clipboard.writeText(peer.publicKey)}
          >
            Copy Public Key
          </DropdownMenuItem>
          <DialogTrigger asChild>
            <DropdownMenuItem className="text-destructive">
              Delete Peer
            </DropdownMenuItem>
          </DialogTrigger>
        </DropdownMenuContent>
      </DropdownMenu>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Peer</DialogTitle>
          <DialogDescription className="grid gap-4 pt-3">
            Do you really want to delete the peer with the following public key?
            <div className="font-mono">{peer.publicKey}</div> This action cannot
            be undone.
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="submit" variant="destructive" onClick={deletePeer}>
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
