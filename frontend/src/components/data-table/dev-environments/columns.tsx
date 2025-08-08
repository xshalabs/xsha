import { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal, Edit, Trash2, Monitor } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { DevEnvironmentDisplay } from "@/types/dev-environment";

interface DevEnvironmentColumnsProps {
  onEdit: (environment: DevEnvironmentDisplay) => void;
  onDelete: (id: number) => void;
}

export const createDevEnvironmentColumns = ({
  onEdit,
  onDelete,
}: DevEnvironmentColumnsProps): ColumnDef<DevEnvironmentDisplay>[] => [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue("name")}</div>
    ),
  },
  {
    accessorKey: "description",
    header: "Description",
    cell: ({ row }) => {
      const description = row.getValue("description") as string;
      return (
        <div className="max-w-[300px] truncate text-muted-foreground">
          {description || "No description"}
        </div>
      );
    },
  },
  {
    accessorKey: "cpu_limit",
    header: "CPU",
    cell: ({ row }) => {
      const cores = row.getValue("cpu_limit") as number;
      const coreCount = cores || 0;
      return (
        <Badge variant="secondary">
          {coreCount} {coreCount === 1 ? "core" : "cores"}
        </Badge>
      );
    },
  },
  {
    accessorKey: "memory_limit",
    header: "Memory",
    cell: ({ row }) => {
      const memoryMb = row.getValue("memory_limit") as number;
      const memory = memoryMb || 0;
      const formatMemory = (mb: number) => {
        if (mb >= 1024) {
          return `${(mb / 1024).toFixed(1)} GB`;
        }
        return `${mb} MB`;
      };
      return (
        <Badge variant="secondary">
          {formatMemory(memory)}
        </Badge>
      );
    },
  },
  {
    id: "env_vars_count",
    header: "Env Vars",
    cell: ({ row }) => {
      const envVarsMap = row.original.env_vars_map || {};
      const count = Object.keys(envVarsMap).length;
      return (
        <Badge variant="outline">
          {count}
        </Badge>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: "Created",
    cell: ({ row }) => {
      const date = new Date(row.getValue("created_at"));
      return (
        <div className="text-sm text-muted-foreground">
          {date.toLocaleDateString()}
        </div>
      );
    },
  },
  {
    id: "actions",
    header: "Actions",
    cell: ({ row }) => {
      const environment = row.original;

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
            <DropdownMenuItem onClick={() => onEdit(environment)}>
              <Edit className="mr-2 h-4 w-4" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => onDelete(environment.id)}
              className="text-destructive"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
    enableSorting: false,
    enableHiding: false,
  },
];
