import { zodResolver } from "@hookform/resolvers/zod";
import { Database, Edit, Plus, Trash2 } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation('ipPools');
  const { t: tCommon } = useTranslation('common');
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
      toast.error(t('messages.loadFailed'));
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
        toast.success(t('messages.updateSuccess'));
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
        toast.success(t('messages.createSuccess'));
        setIsCreateOpen(false);
      }
      reset();
      fetchPools();
    } catch (error: any) {
      const errorMessage = error?.message || (editingPool ? t('messages.updateFailed') : t('messages.createFailed'));
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
      toast.success(t('messages.deleteSuccess'));
      setIsDeleteOpen(false);
      setDeletingPool(null);
      fetchPools();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.deleteFailed');
      toast.error(errorMessage);
    }
  };

  return (
    <div className="space-y-6 p-8 bg-slate-50/50">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
        {isAdmin && (
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" /> {t('create.createButton')}
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>{t('create.title')}</DialogTitle>
                <DialogDescription>
                  {t('create.description')}
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">{t('create.name')} *</Label>
                  <Input
                    id="name"
                    placeholder={t('create.namePlaceholder')}
                    {...register("name")}
                  />
                  {errors.name && (
                    <p className="text-sm text-red-500">{errors.name.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="cidr">{t('create.cidr')} *</Label>
                  <Input
                    id="cidr"
                    placeholder={t('create.cidrPlaceholder')}
                    {...register("cidr")}
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('create.cidrHint')}
                  </p>
                  {errors.cidr && (
                    <p className="text-sm text-red-500">{errors.cidr.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="routes">{t('create.routes')}</Label>
                  <Input
                    id="routes"
                    placeholder={t('create.routesPlaceholder')}
                    {...register("routes")}
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('create.routesHint')}
                  </p>
                  {errors.routes && (
                    <p className="text-sm text-red-500">{errors.routes.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="dns">{t('create.dns')}</Label>
                  <Input
                    id="dns"
                    placeholder={t('create.dnsPlaceholder')}
                    {...register("dns")}
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('create.dnsHint')}
                  </p>
                  {errors.dns && (
                    <p className="text-sm text-red-500">{errors.dns.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="endpoint">{t('create.endpoint')}</Label>
                  <Input
                    id="endpoint"
                    placeholder={t('create.endpointPlaceholder')}
                    {...register("endpoint")}
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('create.endpointHint')}
                  </p>
                  {errors.endpoint && (
                    <p className="text-sm text-red-500">{errors.endpoint.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="description">{t('create.description')}</Label>
                  <Input
                    id="description"
                    placeholder={t('create.descriptionPlaceholder')}
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
                    {tCommon('buttons.cancel')}
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? tCommon('buttons.creating') : t('create.createButton')}
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
              <DialogTitle>{t('edit.title')}</DialogTitle>
              <DialogDescription>
                {t('edit.editDescription')}
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">{t('edit.name')} *</Label>
                <Input
                  id="edit-name"
                  placeholder={t('edit.namePlaceholder')}
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-sm text-red-500">{errors.name.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-cidr">{t('edit.cidr')} *</Label>
                <Input
                  id="edit-cidr"
                  placeholder={t('edit.cidrPlaceholder')}
                  {...register("cidr")}
                  readOnly
                  className="bg-muted"
                />
                <p className="text-xs text-muted-foreground">
                  {t('edit.cidrHint')}
                </p>
                {errors.cidr && (
                  <p className="text-sm text-red-500">{errors.cidr.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-routes">{t('edit.routes')}</Label>
                <Input
                  id="edit-routes"
                  placeholder={t('edit.routesPlaceholder')}
                  {...register("routes")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('edit.routesHint')}
                </p>
                {errors.routes && (
                  <p className="text-sm text-red-500">{errors.routes.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-dns">{t('edit.dns')}</Label>
                <Input
                  id="edit-dns"
                  placeholder={t('edit.dnsPlaceholder')}
                  {...register("dns")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('edit.dnsHint')}
                </p>
                {errors.dns && (
                  <p className="text-sm text-red-500">{errors.dns.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-endpoint">{t('edit.endpoint')}</Label>
                <Input
                  id="edit-endpoint"
                  placeholder={t('edit.endpointPlaceholder')}
                  {...register("endpoint")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('edit.endpointHint')}
                </p>
                {errors.endpoint && (
                  <p className="text-sm text-red-500">{errors.endpoint.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-description">{t('edit.description')}</Label>
                <Input
                  id="edit-description"
                  placeholder={t('edit.descriptionPlaceholder')}
                  {...register("description")}
                />
                {errors.description && (
                  <p className="text-sm text-red-500">{errors.description.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-status">{t('edit.status')}</Label>
                <select
                  id="edit-status"
                  {...register("status")}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="active">{t('edit.statusActive')}</option>
                  <option value="disabled">{t('edit.statusDisabled')}</option>
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
                  {tCommon('buttons.cancel')}
                </Button>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? t('edit.updating') : t('edit.updateButton')}
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
              <TableHead>{t('table.name')}</TableHead>
              <TableHead>{t('table.cidr')}</TableHead>
              <TableHead>{t('table.routes')}</TableHead>
              <TableHead>{t('table.dns')}</TableHead>
              <TableHead>{t('table.endpoint')}</TableHead>
              <TableHead>{t('table.status')}</TableHead>
              {isAdmin && <TableHead>{t('table.actions')}</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 7 : 6} className="text-center py-8">
                  {t('table.loading')}
                </TableCell>
              </TableRow>
            ) : pools.length === 0 ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 7 : 6} className="text-center py-8">
                  {t('table.noPools')}
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
                    {pool.routes || tCommon('common.na')}
                  </TableCell>
                  <TableCell className="font-mono text-sm text-xs">
                    {pool.dns || tCommon('common.na')}
                  </TableCell>
                  <TableCell className="font-mono text-sm text-xs">
                    {pool.endpoint || tCommon('common.na')}
                  </TableCell>
                  <TableCell>
                    {pool.status === "active" ? (
                      <Badge variant="default">{t('edit.statusActive')}</Badge>
                    ) : (
                      <Badge variant="secondary">{t('edit.statusDisabled')}</Badge>
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
            {t('messages.totalPools', { total: pools.length, active: pools.filter((p) => p.status === "active").length })}
          </p>
        </div>
      )}
      {isAdmin && (
        <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>{t('messages.deleteTitle')}</DialogTitle>
              <DialogDescription>
                {t('messages.deleteDescription', { name: deletingPool?.name || '' })}
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
                {tCommon('buttons.cancel')}
              </Button>
              <Button
                type="button"
                variant="destructive"
                onClick={confirmDelete}
              >
                {t('messages.deleteButton')}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
