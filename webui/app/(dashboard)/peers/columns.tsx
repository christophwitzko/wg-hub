import { Column, ColumnDef, Row } from "@tanstack/react-table";
import { formatDistanceToNow } from "date-fns";
import prettyBytes from "pretty-bytes";

import { Peer } from "@/lib/api";
import { DataTableColumnHeader } from "@/components/ui/data-table";
import { CellActions } from "./cell-actions";
import { Badge } from "@/components/ui/badge";

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

export function getColumns(): ColumnDef<Peer>[] {
  return [
    {
      id: "badges",
      enableHiding: false,
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          {row.original.lastHandshake === 0 ? (
            <Badge className="text-destructive">Offline</Badge>
          ) : null}
          {row.original.isHub ? <Badge>Hub</Badge> : null}
          {row.original.isRequester ? <Badge>You</Badge> : null}
        </div>
      ),
    },
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
      cell: CellActions,
    },
  ];
}
