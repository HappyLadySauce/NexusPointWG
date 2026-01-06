import { zodResolver } from "@hookform/resolvers/zod";
import { Database, Edit, Plus, Trash2 } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../components/ui/dialog";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { useAuth } from "../context/AuthContext";
import { api, CreateIPPoolRequest, IPPoolResponse, UpdateIPPoolRequest } from "../services/api";

// Form schema for creating/editing IP pool
const ipPoolSchema = z.object({
  name: z.string().min(1, "Name is required").max(64, "Name must be at most 64 characters"),
  cidr: z.string().min(1, "CIDR is required").regex(/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/, "Invalid CIDR format (e.g., 100.100.100.0/24)"),
  routes: z.string().optional().refine((val) => {
    if (!val || val === "") return true;
    return /^((\d{1,3}\.){3}\d{1,3}\/\d{1,2})(,\s*((\d{1,3}\.){3}\d{1,3}\/\d{1,2}))*$/.test(val);
  }, "Invalid CIDR list format (e.g., 0.0.0.0/0, 192.168.1.0/24)"),
  dns: z.string().optional().refine((val) => {
    if (!val || val === "") return true;
    return /^((\d{1,3}\.){3}\d{1,3})(,\s*((\d{1,3}\.){3}\d{1,3}))*$/.test(val);
  }, "Invalid DNS IP format (e.g., 1.1.1.1, 8.8.8.8)"),
  endpoint: z.string().optional().refine((val) => {
    if (!val || val === "") return true;
    return /^(\d{1,3}\.){3}\d{1,3}:\d{1,5}$/.test(val);
  }, "Invalid Endpoint format (e.g., 10.10.10.10:51820)"),
  description: z.string().max(255, "Description must be at most 255 characters").optional().or(z.literal("")),
  status: z.enum(["active", "disabled"]).optional(),
});

type IPPoolFormValues = z.infer<typeof ipPoolSchema>;

export function IPPools() {
  const { user: currentUser } = useAuth();
  const [pools, setPools] = useState<IPPoolResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editingPool, setEditingPool] = useState<IPPoolResponse | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingPool, setDeletingPool] = useState<IPPoolResponse | null>(null);

  const { register, handleSubmit, reset, formState: { errors, isSubmitting }, setValue } = useForm<IPPoolFormValues>({
    resolver: zodResolver(ipPoolSchema),
  });

  const isAdmin = currentUser?.role === "admin";

  const fetchPools = async () => {
    setLoading(true);
    try {
      const response = await api.wg.listIPPools();
      setPools(response.items || []);
    } catch (e) {
      toast.error("Failed to load IP pools");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPools();
  }, []);

  const onSubmit = async (data: IPPoolFormValues) => {
    try {
      if (editingPool) {
        // Update existing pool - only include fields that have changed
        const request: UpdateIPPoolRequest = {};
        if (data.name !== editingPool.name) {
          request.name = data.name;
        }
        if (data.routes !== (editingPool.routes || "")) {
          request.routes = data.routes || undefined;
        }
        if (data.dns !== (editingPool.dns || "")) {
          request.dns = data.dns || undefined;
        }
        if (data.endpoint !== (editingPool.endpoint || "")) {
          request.endpoint = data.endpoint || undefined;
        }
        if (data.description !== (editingPool.description || "")) {
          request.description = data.description || undefined;
        }
        if (data.status !== editingPool.status) {
          request.status = data.status;
        }

        await api.wg.updateIPPool(editingPool.id, request);
        toast.success("IP pool updated successfully");
        setIsEditOpen(false);
        setEditingPool(null);
      } else {
        // Create new pool
        const request: CreateIPPoolRequest = {
          name: data.name,
          cidr: data.cidr,
          routes: data.routes || undefined,
          dns: data.dns || undefined,
          endpoint: data.endpoint || undefined,
          description: data.description || undefined,
        };

        await api.wg.createIPPool(request);
        toast.success("IP pool created successfully");
        setIsCreateOpen(false);
      }
      reset();
      fetchPools();
    } catch (error: any) {
      const errorMessage = error?.message || (editingPool ? "Failed to update IP pool" : "Failed to create IP pool");
      toast.error(errorMessage);
    }
  };

  const handleEdit = (pool: IPPoolResponse) => {
    setEditingPool(pool);
    setValue("name", pool.name);
    setValue("cidr", pool.cidr);
    setValue("routes", pool.routes || "");
    setValue("dns", pool.dns || "");
    setValue("endpoint", pool.endpoint || "");
    setValue("description", pool.description || "");
    setValue("status", pool.status as "active" | "disabled");
    setIsEditOpen(true);
  };

  const handleEditCancel = () => {
    setIsEditOpen(false);
    setEditingPool(null);
    reset();
  };

  const handleDelete = (pool: IPPoolResponse) => {
    setDeletingPool(pool);
    setIsDeleteOpen(true);
  };

  const confirmDelete = async () => {
    if (!deletingPool) return;

    try {
      await api.wg.deleteIPPool(deletingPool.id);
      toast.success("IP pool deleted successfully");
      setIsDeleteOpen(false);
      setDeletingPool(null);
      fetchPools();
    } catch (error: any) {
      const errorMessage = error?.message || "Failed to delete IP pool";
      toast.error(errorMessage);
    }
  };

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">IP Address Pools</h1>
        {isAdmin && (
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" /> Create IP Pool
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Create New IP Pool</DialogTitle>
                <DialogDescription>
                  Create a new IP address pool for WireGuard peer allocation. All fields with * are required.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name *</Label>
                  <Input
                    id="name"
                    placeholder="e.g. Main Pool"
                    {...register("name")}
                  />
                  {errors.name && (
                    <p className="text-sm text-red-500">{errors.name.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="cidr">CIDR *</Label>
                  <Input
                    id="cidr"
                    placeholder="e.g. 100.100.100.0/24"
                    {...register("cidr")}
                  />
                  <p className="text-xs text-muted-foreground">
                    CIDR range for the IP pool (e.g., 100.100.100.0/24)
                  </p>
                  {errors.cidr && (
                    <p className="text-sm text-red-500">{errors.cidr.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="routes">Routes</Label>
                  <Input
                    id="routes"
                    placeholder="e.g. 100.100.100.0/24, 192.168.1.0/24"
                    {...register("routes")}
                  />
                  <p className="text-xs text-muted-foreground">
                    Comma-separated CIDRs for client AllowedIPs (e.g., 100.100.100.0/24, 192.168.1.0/24)
                  </p>
                  {errors.routes && (
                    <p className="text-sm text-red-500">{errors.routes.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="dns">DNS</Label>
                  <Input
                    id="dns"
                    placeholder="e.g. 1.1.1.1, 223.5.5.5"
                    {...register("dns")}
                  />
                  <p className="text-xs text-muted-foreground">
                    Comma-separated DNS servers (e.g., 1.1.1.1, 223.5.5.5)
                  </p>
                  {errors.dns && (
                    <p className="text-sm text-red-500">{errors.dns.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="endpoint">Endpoint</Label>
                  <Input
                    id="endpoint"
                    placeholder="e.g. 10.10.10.10:51820"
                    {...register("endpoint")}
                  />
                  <p className="text-xs text-muted-foreground">
                    Server endpoint (e.g., 10.10.10.10:51820)
                  </p>
                  {errors.endpoint && (
                    <p className="text-sm text-red-500">{errors.endpoint.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Input
                    id="description"
                    placeholder="Optional description"
                    {...register("description")}
                  />
                  {errors.description && (
                    <p className="text-sm text-red-500">{errors.description.message}</p>
                  )}
                </div>

                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setIsCreateOpen(false);
                      reset();
                    }}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Creating..." : "Create IP Pool"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>
      {isAdmin && (
        <Dialog open={isEditOpen} onOpenChange={setIsEditOpen}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Edit IP Pool</DialogTitle>
              <DialogDescription>
                Update the IP address pool configuration. CIDR can only be modified when no IPs are allocated from this pool.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">Name *</Label>
                <Input
                  id="edit-name"
                  placeholder="e.g. Main Pool"
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-sm text-red-500">{errors.name.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-cidr">CIDR *</Label>
                <Input
                  id="edit-cidr"
                  placeholder="e.g. 100.100.100.0/24"
                  {...register("cidr")}
                />
                <p className="text-xs text-muted-foreground">
                  CIDR can only be modified when no IPs are allocated from this pool
                </p>
                {errors.cidr && (
                  <p className="text-sm text-red-500">{errors.cidr.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-routes">Routes</Label>
                <Input
                  id="edit-routes"
                  placeholder="e.g. 100.100.100.0/24, 192.168.1.0/24"
                  {...register("routes")}
                />
                <p className="text-xs text-muted-foreground">
                  Comma-separated CIDRs for client AllowedIPs (e.g., 100.100.100.0/24, 192.168.1.0/24)
                </p>
                {errors.routes && (
                  <p className="text-sm text-red-500">{errors.routes.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-dns">DNS</Label>
                <Input
                  id="edit-dns"
                  placeholder="e.g. 1.1.1.1, 223.5.5.5"
                  {...register("dns")}
                />
                <p className="text-xs text-muted-foreground">
                  Comma-separated DNS servers (e.g., 1.1.1.1, 223.5.5.5)
                </p>
                {errors.dns && (
                  <p className="text-sm text-red-500">{errors.dns.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-endpoint">Endpoint</Label>
                <Input
                  id="edit-endpoint"
                  placeholder="e.g. 10.10.10.10:51820"
                  {...register("endpoint")}
                />
                <p className="text-xs text-muted-foreground">
                  Server endpoint (e.g., 10.10.10.10:51820)
                </p>
                {errors.endpoint && (
                  <p className="text-sm text-red-500">{errors.endpoint.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-description">Description</Label>
                <Input
                  id="edit-description"
                  placeholder="Optional description"
                  {...register("description")}
                />
                {errors.description && (
                  <p className="text-sm text-red-500">{errors.description.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-status">Status</Label>
                <select
                  id="edit-status"
                  {...register("status")}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="active">Active</option>
                  <option value="disabled">Disabled</option>
                </select>
                {errors.status && (
                  <p className="text-sm text-red-500">{errors.status.message}</p>
                )}
              </div>

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleEditCancel}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "Updating..." : "Update IP Pool"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      )}
      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>CIDR</TableHead>
              <TableHead>Routes</TableHead>
              <TableHead>DNS</TableHead>
              <TableHead>Endpoint</TableHead>
              <TableHead>Status</TableHead>
              {isAdmin && <TableHead>Actions</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 7 : 6} className="text-center py-8">
                  Loading...
                </TableCell>
              </TableRow>
            ) : pools.length === 0 ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 7 : 6} className="text-center py-8">
                  No IP pools found
                </TableCell>
              </TableRow>
            ) : (
              pools.map((pool) => (
                <TableRow key={pool.id}>
                  <TableCell className="font-medium flex items-center gap-2">
                    <Database className="h-4 w-4 text-blue-500" />
                    {pool.name}
                  </TableCell>
                  <TableCell className="font-mono text-sm">{pool.cidr}</TableCell>
                  <TableCell className="font-mono text-sm text-xs">
                    {pool.routes || "N/A"}
                  </TableCell>
                  <TableCell className="font-mono text-sm text-xs">
                    {pool.dns || "N/A"}
                  </TableCell>
                  <TableCell className="font-mono text-sm text-xs">
                    {pool.endpoint || "N/A"}
                  </TableCell>
                  <TableCell>
                    {pool.status === "active" ? (
                      <Badge variant="default">Active</Badge>
                    ) : (
                      <Badge variant="secondary">Disabled</Badge>
                    )}
                  </TableCell>
                  {isAdmin && (
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(pool)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(pool)}
                        >
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
      {pools.length > 0 && (
        <div className="text-sm text-muted-foreground">
          <p>
            Total pools: {pools.length} | Active:{" "}
            {pools.filter((p) => p.status === "active").length}
          </p>
        </div>
      )}
      {isAdmin && (
        <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Delete IP Pool</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete the IP pool "{deletingPool?.name}"?
                This action cannot be undone. The pool can only be deleted when no IPs are allocated from it.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsDeleteOpen(false);
                  setDeletingPool(null);
                }}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="destructive"
                onClick={confirmDelete}
              >
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
