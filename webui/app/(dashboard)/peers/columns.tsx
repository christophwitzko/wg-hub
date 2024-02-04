import { Column, ColumnDef, Row } from "@tanstack/react-table";
import { formatDistanceToNow } from "date-fns";
import prettyBytes from "pretty-bytes";
import { MoreHorizontal } from "lucide-react";

import { Peer } from "@/lib/api";
import { AuthContextType } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { DataTableColumnHeader } from "@/components/ui/data-table";

export const columnNames = {
  publicKey: "Public Key",
  allowedIP: "Allowed IP",
  endpoint: "Endpoint",
  lastHandshake: "Last Handshake",
  txBytes: "Transmitted Bytes",
  rxBytes: "Received Bytes",
} as Record<string, string>;

function stringSort(a: Row<Peer>, b: Row<Peer>, columnId: string) {
  return (a.getValue(columnId) as string).localeCompare(b.getValue(columnId));
}
function header({ column }: { column: Column<Peer> }) {
  return (
    <DataTableColumnHeader column={column} title={columnNames[column.id]} />
  );
}
export function getColumns({
  auth,
}: {
  auth: AuthContextType;
}): ColumnDef<Peer>[] {
  return [
    {
      id: "publicKey",
      accessorKey: "publicKey",
      accessorFn: (peer) => peer.publicKey.slice(0, 16) + "...",
      sortingFn: stringSort,
      header,
      cell: ({ row }) => (
        <div className="font-mono whitespace-nowrap">
          {row.getValue("publicKey")}
        </div>
      ),
    },
    {
      accessorKey: "allowedIP",
      header,
      cell: ({ row }) => <div>{row.getValue("allowedIP")}</div>,
    },
    {
      accessorKey: "endpoint",
      header,
      cell: ({ row }) => <div>{row.getValue("endpoint")}</div>,
    },
    {
      id: "lastHandshake",
      accessorKey: "lastHandshake",
      accessorFn: (peer) => new Date(peer.lastHandshake * 1000),
      sortingFn: "datetime",
      header,
      cell: ({ row }) => {
        const lastHandshake = row.getValue("lastHandshake") as Date;
        return (
          <div className="text-right">
            {lastHandshake.getTime() === 0
              ? "Never"
              : formatDistanceToNow(lastHandshake) + " ago"}
          </div>
        );
      },
    },
    {
      accessorKey: "txBytes",
      sortingFn: "basic",
      header,
      cell: ({ row }) => (
        <div className="text-right">{prettyBytes(row.getValue("txBytes"))}</div>
      ),
    },
    {
      accessorKey: "rxBytes",
      sortingFn: "basic",
      header,
      cell: ({ row }) => (
        <div className="text-right">{prettyBytes(row.getValue("rxBytes"))}</div>
      ),
    },
    {
      id: "actions",
      enableHiding: false,
      cell: ({ row }) => {
        const peer = row.original;

        return (
          <Dialog>
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
                <DialogTitle>Are you absolutely sure?</DialogTitle>
                <DialogDescription>
                  This action cannot be undone. Are you sure you want to
                  permanently delete this file from our servers?
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button type="submit">Confirm</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        );
      },
    },
  ];
}
