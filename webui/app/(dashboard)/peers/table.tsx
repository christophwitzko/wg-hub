"use client";

import * as React from "react";
import {
  Row,
  Column,
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import {
  ArrowUpDown,
  ChevronDown,
  MoreHorizontal,
  ArrowDownAZ,
  ArrowUpAZ,
  Loader2,
  Plus,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import prettyBytes from "pretty-bytes";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Peer } from "@/lib/api";

function SortIcon({
  sortDirection,
}: {
  sortDirection: "asc" | "desc" | false;
}) {
  switch (sortDirection) {
    case "asc":
      return <ArrowDownAZ className="ml-2 size-4" />;
    case "desc":
      return <ArrowUpAZ className="ml-2 size-4" />;
    default:
      return <ArrowUpDown className="ml-2 size-4" />;
  }
}

const columnNames = {
  publicKey: "Public Key",
  allowedIP: "Allowed IP",
  endpoint: "Endpoint",
  lastHandshake: "Last Handshake",
  txBytes: "Transmitted Bytes",
  rxBytes: "Received Bytes",
} as Record<string, string>;

function SortButton({ column }: { column: Column<Peer> }) {
  return (
    <Button
      variant="ghost"
      onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
    >
      {columnNames[column.id]}
      <SortIcon sortDirection={column.getIsSorted()} />
    </Button>
  );
}

function stringSort(a: Row<Peer>, b: Row<Peer>, columnId: string) {
  return (a.getValue(columnId) as string).localeCompare(b.getValue(columnId));
}

export const columns: ColumnDef<Peer>[] = [
  {
    accessorKey: "publicKey",
    accessorFn: (peer) => peer.publicKey.slice(0, 16) + "...",
    sortingFn: stringSort,
    header: ({ column }) => <SortButton column={column} />,
    cell: ({ row }) => (
      <div className="font-mono whitespace-nowrap">
        {row.getValue("publicKey")}
      </div>
    ),
  },
  {
    accessorKey: "allowedIP",
    header: ({ column }) => <SortButton column={column} />,
    cell: ({ row }) => <div>{row.getValue("allowedIP")}</div>,
  },
  {
    accessorKey: "endpoint",
    header: ({ column }) => <SortButton column={column} />,
    cell: ({ row }) => <div>{row.getValue("endpoint")}</div>,
  },
  {
    accessorKey: "lastHandshake",
    accessorFn: (peer) => new Date(peer.lastHandshake * 1000),
    sortingFn: "datetime",
    header: ({ column }) => <SortButton column={column} />,
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
    header: ({ column }) => <SortButton column={column} />,
    cell: ({ row }) => (
      <div className="text-right">{prettyBytes(row.getValue("txBytes"))}</div>
    ),
  },
  {
    accessorKey: "rxBytes",
    sortingFn: "basic",
    header: ({ column }) => <SortButton column={column} />,
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
            <DropdownMenuItem className="text-destructive">
              Delete Peer
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  },
];

export function PeersTable({
  data,
  isLoading,
}: {
  data: Peer[];
  isLoading?: boolean;
}) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    initialState: {
      sorting: [{ id: "publicKey", desc: false }],
    },
  });

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Button variant="outline">
          Add Peer <Plus className="ml-2 size-4" />
        </Button>
        {isLoading ? (
          <Loader2 className="ml-6 size-6 animate-spin text-primary/50" />
        ) : null}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              Columns <ChevronDown className="ml-2 size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    className="capitalize"
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) => column.toggleVisibility(value)}
                  >
                    {columnNames[column.id]}
                  </DropdownMenuCheckboxItem>
                );
              })}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext(),
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
