import { zodResolver } from "@hookform/resolvers/zod";
import { Copy, Download, Edit, Eye, EyeOff, FileText, MoreHorizontal, Plus, Search, Trash2, X } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import * as z from "zod";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Checkbox } from "../components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "../components/ui/dropdown-menu";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "../components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { useAuth } from "../context/AuthContext";
import { api, BatchCreateWGPeersRequest, BatchUpdateWGPeersRequest, CreateWGPeerRequest, IPPoolResponse, UserResponse, WGPeerResponse } from "../services/api";

// Enhanced form schema
const peerSchema = z.object({
  username: z.string().optional(), // Admin can specify
  device_name: z.string().min(1, "Device name is required").max(64, "Device name must be at most 64 characters"),
  ip_pool_id: z.string().optional(),
  client_ip: z.string().optional().refine((val) => {
    if (!val || val === "") return true;
    return /^(\d{1,3}\.){3}\d{1,3}$/.test(val);
  }, "Invalid IPv4 address format"),
  allowed_ips: z.string().optional().refine((val) => {
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
  persistent_keepalive: z.number().min(0).max(65535).optional(),
}).refine((data) => {
  // Manual mode (no IP Pool or None selected) requires allowed_ips
  const isManualMode = !data.ip_pool_id || data.ip_pool_id === "__none__";
  if (isManualMode && !data.allowed_ips) {
    return false;
  }
  return true;
}, {
  message: "Allowed IPs is required when no IP Pool is selected",
  path: ["allowed_ips"],
});

type PeerFormValues = z.infer<typeof peerSchema>;

// Edit peer form schema
const editPeerSchema = z.object({
  device_name: z.string().min(1, "Device name is required").max(64, "Device name must be at most 64 characters"),
  client_ip: z.string().optional().refine((val) => {
    if (!val || val === "") return true;
    return /^(\d{1,3}\.){3}\d{1,3}$/.test(val);
  }, "Invalid IPv4 address format"),
  ip_pool_id: z.string().optional(),
  client_private_key: z.string().optional(),
  allowed_ips: z.string().optional().refine((val) => {
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
  persistent_keepalive: z.number().min(0).max(65535).optional(),
  status: z.enum(["active", "disabled"]).optional(),
  username: z.string().optional(),
});

type EditPeerFormValues = z.infer<typeof editPeerSchema>;

export function Peers() {
  const { user: currentUser } = useAuth();
  const { t } = useTranslation('peers');
  const { t: tCommon } = useTranslation('common');
  const [peers, setPeers] = useState<WGPeerResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isViewConfigOpen, setIsViewConfigOpen] = useState(false);
  const [editingPeer, setEditingPeer] = useState<WGPeerResponse | null>(null);
  const [viewingConfig, setViewingConfig] = useState<{ peerId: string; peerName: string; config: string } | null>(null);
  const [showPrivateKey, setShowPrivateKey] = useState(false);
  const [ipPoolModified, setIpPoolModified] = useState(false);

  // Batch operation states
  const [selectedPeers, setSelectedPeers] = useState<Set<string>>(new Set());
  const [isBatchCreateOpen, setIsBatchCreateOpen] = useState(false);
  const [isBatchEditOpen, setIsBatchEditOpen] = useState(false);
  const [isBatchDeleteOpen, setIsBatchDeleteOpen] = useState(false);
  const [batchItems, setBatchItems] = useState<CreateWGPeerRequest[]>([]);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  // Enhanced state management
  const [pools, setPools] = useState<IPPoolResponse[]>([]);
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [selectedPool, setSelectedPool] = useState<string>("");

  const { register, handleSubmit, reset, formState: { errors, isSubmitting }, watch, setValue } = useForm<PeerFormValues>({
    resolver: zodResolver(peerSchema),
    defaultValues: {
      persistent_keepalive: 25,
    },
  });

  // Edit form
  const { register: registerEdit, handleSubmit: handleEditSubmit, reset: resetEdit, formState: { errors: editErrors, isSubmitting: isEditSubmitting }, watch: watchEdit, setValue: setEditValue } = useForm<EditPeerFormValues>({
    resolver: zodResolver(editPeerSchema),
  });

  const isAdmin = currentUser?.role === "admin";
  const ipPoolId = watch("ip_pool_id");

  const fetchPeers = async () => {
    setLoading(true);
    try {
      const offset = (currentPage - 1) * pageSize;
      const options: any = {
        offset,
        limit: pageSize,
      };

      // Use backend search if search term is provided
      if (search.trim()) {
        options.device_name = search.trim();
      }

      const response = await api.wg.listPeers(options);
      setPeers(response.items || []);
      setTotal(response.total || 0);
    } catch (error) {
      toast.error(t('messages.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPeers();
  }, [currentPage, search, isAdmin]);

  useEffect(() => {
    // Only admins need to fetch IP pools (for creating peers with pool selection)
    if (isAdmin) {
      fetchPools();
      fetchUsers();
    }
  }, [isAdmin]);

  // Reset to first page when search changes
  useEffect(() => {
    if (currentPage !== 1) {
      setCurrentPage(1);
    }
  }, [search]);

  const fetchPools = async () => {
    try {
      const response = await api.wg.listIPPools({ status: "active" });
      setPools(response.items || []);
    } catch (error) {
      toast.error(t('messages.loadPoolsFailed'));
    }
  };

  const fetchUsers = async () => {
    try {
      const response = await api.users.list();
      setUsers(response.items || []);
    } catch (error) {
      toast.error(t('messages.loadUsersFailed'));
    }
  };

  const handlePoolChange = (poolID: string) => {
    setSelectedPool(poolID);
    setValue("ip_pool_id", poolID);

    // Auto-fill Pool configuration
    const pool = pools.find(p => p.id === poolID);
    if (pool) {
      if (pool.routes) setValue("allowed_ips", pool.routes);
      if (pool.dns) setValue("dns", pool.dns);
      if (pool.endpoint) setValue("endpoint", pool.endpoint);
    }
  };


  const onSubmit = async (data: PeerFormValues) => {
    // Admin must select user
    if (isAdmin && !data.username) {
      toast.error("User is required for admin");
      return;
    }

    try {
      const request: CreateWGPeerRequest = {
        device_name: data.device_name,
      };

      // Admin can specify username
      if (isAdmin && data.username) {
        request.username = data.username;
      }

      // Determine mode based on ip_pool_id
      const hasIPPool = data.ip_pool_id && data.ip_pool_id !== "__none__";

      if (hasIPPool) {
        // Auto mode: IP Pool selected
        request.ip_pool_id = data.ip_pool_id;
        // Use Pool config, but allow manual override
        if (data.allowed_ips) request.allowed_ips = data.allowed_ips;
        if (data.dns) request.dns = data.dns;
        if (data.endpoint) request.endpoint = data.endpoint;
      } else {
        // Manual mode: No IP Pool selected
        // All fields need to be manually filled
        if (data.allowed_ips) request.allowed_ips = data.allowed_ips;
        if (data.dns) request.dns = data.dns;
        if (data.endpoint) request.endpoint = data.endpoint;
      }

      // Optional fields
      if (data.client_ip) request.client_ip = data.client_ip;
      if (data.persistent_keepalive) {
        request.persistent_keepalive = data.persistent_keepalive;
      }

      await api.wg.createPeer(request);
      toast.success(t('messages.createSuccess'));
      setIsCreateOpen(false);
      reset();
      setSelectedPool("");
      // Refresh current page
      fetchPeers();
    } catch (error: any) {
      toast.error(error?.message || t('messages.createFailed'));
    }
  };

  const handleDelete = async (id: string) => {
    if (confirm(t('messages.deleteConfirm'))) {
      try {
        await api.wg.deletePeer(id);
        toast.success(t('messages.deleteSuccess'));

        // If we deleted the last item on the current page and it's not the first page, go to previous page
        if (peers.length === 1 && currentPage > 1) {
          setCurrentPage(currentPage - 1);
        } else {
          fetchPeers();
        }
      } catch (error) {
        toast.error(t('messages.deleteFailed'));
      }
    }
  };

  const handleViewConfig = async (id: string) => {
    try {
      const peer = peers.find(p => p.id === id);
      const peerName = peer ? peer.device_name : `Peer ${id}`;

      toast.info(t('messages.loadingConfig'));
      const configText = await api.wg.downloadConfig(id);

      setViewingConfig({ peerId: id, peerName, config: configText });
      setIsViewConfigOpen(true);
    } catch (error) {
      toast.error(t('messages.loadConfigFailed'));
      console.error(error);
    }
  };

  const handleCopyConfig = async () => {
    if (!viewingConfig) return;

    try {
      await navigator.clipboard.writeText(viewingConfig.config);
      toast.success(t('messages.copyConfigSuccess'));
    } catch (error) {
      toast.error(t('messages.copyConfigFailed'));
      console.error(error);
    }
  };

  const handleDownloadConfig = async (id: string) => {
    try {
      toast.info(t('messages.downloadingConfig'));
      const configText = await api.wg.downloadConfig(id);

      // Find peer to create a good filename
      const peer = peers.find(p => p.id === id);
      const filename = peer
        ? `${peer.device_name.replace(/[^a-zA-Z0-9_-]/g, '_')}.conf`
        : `wg-peer-${id}.conf`;

      const blob = new Blob([configText], { type: "text/plain" });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();

      // Cleanup
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast.success(t('messages.downloadConfigSuccess'));
    } catch (error) {
      toast.error(t('messages.downloadConfigFailed'));
      console.error(error);
    }
  };

  const handleEdit = (peer: WGPeerResponse) => {
    setEditingPeer(peer);
    setShowPrivateKey(false); // Reset show/hide state
    setIpPoolModified(false); // Reset IP Pool modification flag
    // Extract IP from CIDR format (e.g., "100.100.100.1/32" -> "100.100.100.1")
    const clientIP = peer.client_ip.split("/")[0];
    resetEdit({
      device_name: peer.device_name,
      client_ip: clientIP,
      ip_pool_id: peer.ip_pool_id || undefined,
      client_private_key: peer.client_private_key || "", // Show current private key
      allowed_ips: peer.allowed_ips || "",
      dns: peer.dns || "",
      endpoint: peer.endpoint || "",
      persistent_keepalive: peer.persistent_keepalive,
      status: peer.status as "active" | "disabled",
      username: peer.username || undefined,
    });
    setIsEditOpen(true);
  };

  const onEditSubmit = async (data: EditPeerFormValues) => {
    if (!editingPeer) return;

    try {
      const request: any = {};
      const existingIP = editingPeer.client_ip.split("/")[0];

      // Only include fields that are provided (partial update)
      if (data.device_name !== editingPeer.device_name) {
        request.device_name = data.device_name;
      }
      if (data.client_ip && data.client_ip !== existingIP) {
        request.client_ip = data.client_ip;
      }
      // Handle ip_pool_id: allow clearing (empty string) or changing
      const currentPoolId = editingPeer.ip_pool_id || "";
      // Convert undefined to empty string to explicitly clear IP Pool
      const newPoolId = data.ip_pool_id === undefined ? "" : (data.ip_pool_id || "");
      if (newPoolId !== currentPoolId) {
        // Explicitly send empty string to clear IP Pool, or the new pool ID
        request.ip_pool_id = newPoolId === "" ? "" : newPoolId;
      }
      // Only admins can update private key
      if (isAdmin && data.client_private_key && data.client_private_key !== "") {
        // Only send if it's different from current (or if current is empty)
        const currentKey = editingPeer.client_private_key || "";
        if (data.client_private_key !== currentKey) {
          request.client_private_key = data.client_private_key;
        }
      }
      if (data.allowed_ips !== (editingPeer.allowed_ips || "")) {
        request.allowed_ips = data.allowed_ips || undefined;
      }
      if (data.dns !== (editingPeer.dns || "")) {
        request.dns = data.dns || undefined;
      }
      if (data.endpoint !== (editingPeer.endpoint || "")) {
        request.endpoint = data.endpoint || undefined;
      }
      if (data.persistent_keepalive !== editingPeer.persistent_keepalive) {
        request.persistent_keepalive = data.persistent_keepalive;
      }
      if (data.status !== editingPeer.status) {
        request.status = data.status;
      }
      // Only admins can update username (user binding)
      if (isAdmin && data.username && data.username !== editingPeer.username) {
        request.username = data.username;
      }

      await api.wg.updatePeer(editingPeer.id, request);
      toast.success(t('messages.updateSuccess'));
      setIsEditOpen(false);
      setEditingPeer(null);
      setShowPrivateKey(false);
      setIpPoolModified(false);
      resetEdit();
      // Refresh current page
      fetchPeers();
    } catch (error: any) {
      toast.error(error?.message || t('messages.updateFailed'));
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  // Generate page numbers to display
  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const maxPages = 7;

    if (totalPages <= maxPages) {
      // Show all pages if total pages is less than max
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show first page, ellipsis, current page range, ellipsis, last page
      if (currentPage <= 3) {
        // Near the beginning
        for (let i = 1; i <= 4; i++) {
          pages.push(i);
        }
        pages.push('ellipsis');
        pages.push(totalPages);
      } else if (currentPage >= totalPages - 2) {
        // Near the end
        pages.push(1);
        pages.push('ellipsis');
        for (let i = totalPages - 3; i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        // In the middle
        pages.push(1);
        pages.push('ellipsis');
        for (let i = currentPage - 1; i <= currentPage + 1; i++) {
          pages.push(i);
        }
        pages.push('ellipsis');
        pages.push(totalPages);
      }
    }

    return pages;
  };

  // Batch operation handlers
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      const allIds = new Set(peers.map(p => p.id));
      setSelectedPeers(allIds);
    } else {
      setSelectedPeers(new Set());
    }
  };

  const handleSelectPeer = (id: string, checked: boolean) => {
    const newSelected = new Set(selectedPeers);
    if (checked) {
      newSelected.add(id);
    } else {
      newSelected.delete(id);
    }
    setSelectedPeers(newSelected);
  };

  const isAllSelected = peers.length > 0 && peers.every(p => selectedPeers.has(p.id));
  const isIndeterminate = selectedPeers.size > 0 && selectedPeers.size < peers.length;

  const handleBatchDelete = () => {
    if (selectedPeers.size === 0) return;
    setIsBatchDeleteOpen(true);
  };

  const confirmBatchDelete = async () => {
    if (selectedPeers.size === 0) return;

    try {
      const ids = Array.from(selectedPeers);
      await api.wg.batchDeletePeers({ ids });
      toast.success(t('messages.batchDeleteSuccess', { count: ids.length }));
      setIsBatchDeleteOpen(false);
      setSelectedPeers(new Set());
      fetchPeers();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.batchDeleteFailed');
      toast.error(errorMessage);
    }
  };

  const handleBatchEdit = () => {
    if (selectedPeers.size === 0) return;
    setIsBatchEditOpen(true);
  };

  const handleBatchCreate = () => {
    setBatchItems([{
      device_name: '',
    }]);
    setIsBatchCreateOpen(true);
  };

  const addBatchItem = () => {
    if (batchItems.length >= 50) {
      toast.error(t('messages.batchMaxItems', { max: 50 }));
      return;
    }
    setBatchItems([...batchItems, {
      device_name: '',
    }]);
  };

  const removeBatchItem = (index: number) => {
    setBatchItems(batchItems.filter((_, i) => i !== index));
  };

  const updateBatchItem = (index: number, field: keyof CreateWGPeerRequest, value: string | undefined) => {
    const newItems = [...batchItems];
    newItems[index] = { ...newItems[index], [field]: value };
    setBatchItems(newItems);
  };

  const handleBatchCreateSubmit = async () => {
    if (batchItems.length === 0) {
      toast.error(t('messages.batchEmpty'));
      return;
    }

    // Validate all items
    for (let i = 0; i < batchItems.length; i++) {
      const item = batchItems[i];
      if (!item.device_name) {
        toast.error(t('messages.batchInvalidItem', { index: i + 1 }));
        return;
      }
      if (isAdmin && !item.username) {
        toast.error(t('messages.batchInvalidItem', { index: i + 1 }));
        return;
      }
    }

    try {
      const request: BatchCreateWGPeersRequest = { items: batchItems };
      await api.wg.batchCreatePeers(request);
      toast.success(t('messages.batchCreateSuccess', { count: batchItems.length }));
      setIsBatchCreateOpen(false);
      setBatchItems([]);
      fetchPeers();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.batchCreateFailed');
      toast.error(errorMessage);
    }
  };

  const handleBatchEditSubmit = async (data: { status?: string }) => {
    if (selectedPeers.size === 0) return;

    try {
      const items = Array.from(selectedPeers).map(id => ({
        id,
        ...(data.status && { status: data.status as "active" | "disabled" }),
      }));

      const request: BatchUpdateWGPeersRequest = { items };
      await api.wg.batchUpdatePeers(request);
      toast.success(t('messages.batchUpdateSuccess', { count: items.length }));
      setIsBatchEditOpen(false);
      setSelectedPeers(new Set());
      fetchPeers();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.batchUpdateFailed');
      toast.error(errorMessage);
    }
  };

  return (
    <div className="space-y-6 p-8 bg-slate-50/50">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
        {isAdmin && (
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleBatchCreate}>
              <Plus className="mr-2 h-4 w-4" /> {t('batch.createButton')}
            </Button>
            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" /> {tCommon('buttons.create')} {t('title')}
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[90vh] flex flex-col">
                <DialogHeader className="flex-shrink-0">
                  <DialogTitle>{t('create.title')}</DialogTitle>
                  <DialogDescription>
                    {t('create.description')}
                  </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit(onSubmit)} className="flex-1 flex flex-col min-h-0">
                  <div className="flex-1 overflow-y-auto space-y-4">
                    {/* Basic Information */}
                    <div className="space-y-4">
                      <div className="space-y-2">
                        <Label htmlFor="device_name">{t('create.deviceName')} *</Label>
                        <Input id="device_name" placeholder={t('create.deviceNamePlaceholder')} {...register("device_name")} />
                        {errors.device_name && <p className="text-sm text-red-500">{errors.device_name.message}</p>}
                      </div>

                      {isAdmin && (
                        <div className="space-y-2">
                          <Label htmlFor="username">{t('create.user')} *</Label>
                          <Select
                            value={watch("username") || ""}
                            onValueChange={(value) => setValue("username", value)}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder={t('create.selectUser')} />
                            </SelectTrigger>
                            <SelectContent>
                              {users.map((user) => (
                                <SelectItem key={user.username} value={user.username}>
                                  {user.username} {user.nickname && `(${user.nickname})`}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          {errors.username && <p className="text-sm text-red-500">{errors.username.message}</p>}
                        </div>
                      )}
                    </div>

                    {/* Configuration Fields */}
                    <div className="space-y-4 border-t pt-4">
                      <div className="space-y-2">
                        <Label htmlFor="ip_pool_id">{t('create.ipPool')}</Label>
                        <Select
                          value={watch("ip_pool_id") || "__none__"}
                          onValueChange={(value) => {
                            const poolId = value === "__none__" ? undefined : value;
                            setValue("ip_pool_id", poolId);
                            if (poolId) {
                              handlePoolChange(poolId);
                            } else {
                              // Clear auto-filled fields when selecting None
                              setValue("allowed_ips", "");
                              setValue("dns", "");
                              setValue("endpoint", "");
                            }
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder={t('create.selectIPPool')} />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="__none__">{tCommon('common.none')}</SelectItem>
                            {pools.map((pool) => (
                              <SelectItem key={pool.id} value={pool.id}>
                                {pool.name} ({pool.cidr})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        {errors.ip_pool_id && <p className="text-sm text-red-500">{errors.ip_pool_id.message}</p>}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="client_ip">{t('create.clientIP')}</Label>
                        <Input
                          id="client_ip"
                          placeholder={t('create.clientIPPlaceholder')}
                          {...register("client_ip")}
                        />
                        {errors.client_ip && <p className="text-sm text-red-500">{errors.client_ip.message}</p>}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="allowed_ips">
                          {(!ipPoolId || ipPoolId === "__none__") ? t('create.allowedIPsRequired') : t('create.allowedIPsAutoFilled')}
                        </Label>
                        <Input
                          id="allowed_ips"
                          placeholder={t('create.allowedIPsPlaceholder')}
                          {...register("allowed_ips")}
                        />
                        {errors.allowed_ips && <p className="text-sm text-red-500">{errors.allowed_ips.message}</p>}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="dns">{t('create.dns')}</Label>
                        <Input
                          id="dns"
                          placeholder={t('create.dnsPlaceholder')}
                          {...register("dns")}
                        />
                        {errors.dns && <p className="text-sm text-red-500">{errors.dns.message}</p>}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="endpoint">{t('create.endpoint')}</Label>
                        <Input
                          id="endpoint"
                          placeholder={t('create.endpointPlaceholder')}
                          {...register("endpoint")}
                        />
                        {errors.endpoint && <p className="text-sm text-red-500">{errors.endpoint.message}</p>}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="persistent_keepalive">{t('create.persistentKeepalive')}</Label>
                        <Input
                          id="persistent_keepalive"
                          type="number"
                          min="0"
                          max="65535"
                          placeholder={t('create.persistentKeepalivePlaceholder')}
                          {...register("persistent_keepalive", { valueAsNumber: true })}
                        />
                        {errors.persistent_keepalive && <p className="text-sm text-red-500">{errors.persistent_keepalive.message}</p>}
                      </div>
                    </div>
                  </div>
                  <DialogFooter className="flex-shrink-0">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        setIsCreateOpen(false);
                        reset();
                        setSelectedPool("");
                      }}
                    >
                      {tCommon('buttons.cancel')}
                    </Button>
                    <Button type="submit" disabled={isSubmitting}>
                      {isSubmitting ? tCommon('status.loading') : t('create.createButton')}
                    </Button>
                  </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        )}
      </div>

      {/* Batch operation toolbar */}
      {selectedPeers.size > 0 && (
        <div className="flex items-center justify-between p-4 bg-blue-50 border border-blue-200 rounded-md">
          <div className="flex items-center gap-4">
            <span className="text-sm font-medium">
              {t('batch.selected', { count: selectedPeers.size })}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setSelectedPeers(new Set())}
            >
              <X className="mr-2 h-4 w-4" /> {t('batch.clearSelection')}
            </Button>
          </div>
          <div className="flex gap-2">
            {isAdmin && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleBatchEdit}
                  disabled={selectedPeers.size > 50}
                >
                  {t('batch.editButton')}
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={handleBatchDelete}
                  disabled={selectedPeers.size > 50}
                >
                  <Trash2 className="mr-2 h-4 w-4" /> {t('batch.deleteButton')}
                </Button>
              </>
            )}
          </div>
        </div>
      )}

      <div className="flex items-center space-x-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('table.searchPlaceholder')}
            className="pl-9"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={isAllSelected}
                  onCheckedChange={handleSelectAll}
                  ref={(el) => {
                    if (el) {
                      el.indeterminate = isIndeterminate;
                    }
                  }}
                />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>IP Address</TableHead>
              <TableHead>Endpoint</TableHead>
              <TableHead>DNS</TableHead>
              {isAdmin && <TableHead>User</TableHead>}
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 8 : 7} className="text-center py-8">{t('table.loading')}</TableCell>
              </TableRow>
            ) : peers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 8 : 7} className="text-center py-8">{t('table.noPeers')}</TableCell>
              </TableRow>
            ) : (
              peers.map((peer) => (
                <TableRow key={peer.id}>
                  <TableCell>
                    <Checkbox
                      checked={selectedPeers.has(peer.id)}
                      onCheckedChange={(checked) => handleSelectPeer(peer.id, checked as boolean)}
                    />
                  </TableCell>
                  <TableCell className="font-medium">
                    <div className="flex flex-col">
                      <span>{peer.device_name}</span>
                      <span className="text-xs text-muted-foreground font-mono">{peer.client_public_key.substring(0, 10)}...</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="text-sm">{peer.client_ip}</div>
                    {peer.allowed_ips && (
                      <div className="text-xs text-muted-foreground">{peer.allowed_ips}</div>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{peer.endpoint || tCommon('common.na')}</TableCell>
                  <TableCell className="text-muted-foreground text-sm">{peer.dns || tCommon('common.na')}</TableCell>
                  {isAdmin && (
                    <TableCell>
                      <span className="text-sm">{peer.username || tCommon('common.na')}</span>
                    </TableCell>
                  )}
                  <TableCell>
                    {peer.status === 'active' ? (
                      <Badge variant="default">{tCommon('status.active')}</Badge>
                    ) : (
                      <Badge variant="secondary">{tCommon('status.inactive')}</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>{tCommon('common.actions')}</DropdownMenuLabel>
                        <DropdownMenuItem onClick={() => handleViewConfig(peer.id)}>
                          <FileText className="mr-2 h-4 w-4" /> {t('table.viewConfig')}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleDownloadConfig(peer.id)}>
                          <Download className="mr-2 h-4 w-4" /> {tCommon('buttons.download')} {tCommon('buttons.view')}
                        </DropdownMenuItem>
                        {isAdmin && (
                          <>
                            <DropdownMenuItem onClick={() => handleEdit(peer)}>
                              <Edit className="mr-2 h-4 w-4" /> {t('table.edit')}
                            </DropdownMenuItem>
                            <DropdownMenuItem className="text-red-600" onClick={() => handleDelete(peer.id)}>
                              <Trash2 className="mr-2 h-4 w-4" /> {t('table.delete')}
                            </DropdownMenuItem>
                          </>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {totalPages > 0 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            {tCommon('pagination.total')} {total} {tCommon('pagination.items')}
          </div>
          <Pagination>
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  href="#"
                  onClick={(e) => {
                    e.preventDefault();
                    if (currentPage > 1) {
                      setCurrentPage(currentPage - 1);
                    }
                  }}
                  className={currentPage === 1 ? "pointer-events-none opacity-50" : ""}
                >
                  {tCommon('pagination.previous')}
                </PaginationPrevious>
              </PaginationItem>
              {getPageNumbers().map((page, index) => (
                <PaginationItem key={index}>
                  {page === 'ellipsis' ? (
                    <PaginationEllipsis />
                  ) : (
                    <PaginationLink
                      href="#"
                      onClick={(e) => {
                        e.preventDefault();
                        setCurrentPage(page as number);
                      }}
                      isActive={currentPage === page}
                    >
                      {page}
                    </PaginationLink>
                  )}
                </PaginationItem>
              ))}
              <PaginationItem>
                <PaginationNext
                  href="#"
                  onClick={(e) => {
                    e.preventDefault();
                    if (currentPage < totalPages) {
                      setCurrentPage(currentPage + 1);
                    }
                  }}
                  className={currentPage === totalPages ? "pointer-events-none opacity-50" : ""}
                >
                  {tCommon('pagination.next')}
                </PaginationNext>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      )}

      {/* Edit Peer Dialog */}
      <Dialog open={isEditOpen} onOpenChange={(open) => {
        setIsEditOpen(open);
        if (!open) {
          setEditingPeer(null);
          setShowPrivateKey(false);
          setIpPoolModified(false);
          resetEdit();
        }
      }}>
        <DialogContent className="max-w-2xl max-h-[90vh] flex flex-col">
          <DialogHeader className="flex-shrink-0">
            <DialogTitle>{t('edit.title')}</DialogTitle>
            <DialogDescription>
              {t('edit.description')}
            </DialogDescription>
          </DialogHeader>
          {editingPeer && (
            <form onSubmit={handleEditSubmit(onEditSubmit)} className="flex-1 flex flex-col min-h-0">
              <div className="flex-1 overflow-y-auto space-y-4">
                {/* Read-only fields */}
                <div className="space-y-4 border-b pb-4">
                  <div className="space-y-2">
                    <Label>{t('edit.clientPublicKey')}</Label>
                    <Input value={editingPeer.client_public_key} readOnly className="bg-muted font-mono text-xs" />
                  </div>
                </div>

                {/* Editable fields */}
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="edit_device_name">{t('edit.deviceName')} *</Label>
                    <Input
                      id="edit_device_name"
                      placeholder={t('edit.deviceNamePlaceholder')}
                      {...registerEdit("device_name")}
                    />
                    {editErrors.device_name && (
                      <p className="text-sm text-red-500">{editErrors.device_name.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_client_ip">{t('edit.clientIP')}</Label>
                    <Input
                      id="edit_client_ip"
                      placeholder={t('edit.clientIPPlaceholder')}
                      {...registerEdit("client_ip")}
                    />
                    <p className="text-xs text-muted-foreground">
                      {t('edit.clientIPHint')}
                    </p>
                    {editErrors.client_ip && (
                      <p className="text-sm text-red-500">{editErrors.client_ip.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_ip_pool_id">{t('edit.ipPool')}</Label>
                    <Select
                      value={ipPoolModified ? (watchEdit("ip_pool_id") || "__none__") : (editingPeer?.ip_pool_id || "__none__")}
                      onValueChange={(value) => {
                        setIpPoolModified(true);
                        setEditValue("ip_pool_id", value === "__none__" ? undefined : value);
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t('edit.selectIPPool')} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__none__">{tCommon('common.none')}</SelectItem>
                        {pools.map((pool) => (
                          <SelectItem key={pool.id} value={pool.id}>
                            {pool.name} ({pool.cidr})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      {t('edit.ipPoolHint')}
                    </p>
                    {editErrors.ip_pool_id && (
                      <p className="text-sm text-red-500">{editErrors.ip_pool_id.message}</p>
                    )}
                  </div>

                  {isAdmin && (
                    <div className="space-y-2">
                      <Label htmlFor="edit_username">{t('edit.user')}</Label>
                      <Select
                        value={watchEdit("username") || editingPeer?.username || ""}
                        onValueChange={(value) => {
                          setEditValue("username", value === "" ? undefined : value);
                        }}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder={t('edit.userPlaceholder')} />
                        </SelectTrigger>
                        <SelectContent>
                          {users.map((user) => (
                            <SelectItem key={user.id} value={user.username}>
                              {user.username} {user.nickname ? `(${user.nickname})` : ""}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <p className="text-xs text-muted-foreground">
                        {t('edit.userHint')}
                      </p>
                      {editErrors.username && (
                        <p className="text-sm text-red-500">{editErrors.username.message}</p>
                      )}
                    </div>
                  )}

                  <div className="space-y-2">
                    <Label htmlFor="edit_client_private_key">{t('edit.clientPrivateKey')}</Label>
                    <div className="relative">
                      <Input
                        id="edit_client_private_key"
                        type={showPrivateKey ? "text" : "password"}
                        placeholder={t('edit.clientPrivateKeyPlaceholder')}
                        {...registerEdit("client_private_key")}
                        disabled={!isAdmin}
                        className={!isAdmin ? "bg-muted" : ""}
                      />
                      <button
                        type="button"
                        onClick={() => setShowPrivateKey(!showPrivateKey)}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      >
                        {showPrivateKey ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {t('edit.clientPrivateKeyHint')}
                    </p>
                    {editErrors.client_private_key && (
                      <p className="text-sm text-red-500">{editErrors.client_private_key.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_allowed_ips">{t('edit.allowedIPs')}</Label>
                    <Input
                      id="edit_allowed_ips"
                      placeholder={t('edit.allowedIPsPlaceholder')}
                      {...registerEdit("allowed_ips")}
                    />
                    <p className="text-xs text-muted-foreground">
                      {t('edit.allowedIPsHint')}
                    </p>
                    {editErrors.allowed_ips && (
                      <p className="text-sm text-red-500">{editErrors.allowed_ips.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_dns">{t('edit.dns')}</Label>
                    <Input
                      id="edit_dns"
                      placeholder={t('edit.dnsPlaceholder')}
                      {...registerEdit("dns")}
                    />
                    <p className="text-xs text-muted-foreground">
                      {t('edit.dnsHint')}
                    </p>
                    {editErrors.dns && (
                      <p className="text-sm text-red-500">{editErrors.dns.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_endpoint">{t('edit.endpoint')}</Label>
                    <Input
                      id="edit_endpoint"
                      placeholder={t('edit.endpointPlaceholder')}
                      {...registerEdit("endpoint")}
                    />
                    <p className="text-xs text-muted-foreground">
                      {t('edit.endpointHint')}
                    </p>
                    {editErrors.endpoint && (
                      <p className="text-sm text-red-500">{editErrors.endpoint.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_persistent_keepalive">Persistent Keepalive (seconds)</Label>
                    <Input
                      id="edit_persistent_keepalive"
                      type="number"
                      min="0"
                      max="65535"
                      placeholder="25"
                      {...registerEdit("persistent_keepalive", { valueAsNumber: true })}
                    />
                    <p className="text-xs text-muted-foreground">
                      Keepalive interval in seconds (0-65535, optional)
                    </p>
                    {editErrors.persistent_keepalive && (
                      <p className="text-sm text-red-500">{editErrors.persistent_keepalive.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="edit_status">{tCommon('common.status')}</Label>
                    <Select
                      value={watchEdit("status") || editingPeer.status}
                      onValueChange={(value) => setEditValue("status", value as "active" | "disabled")}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={tCommon('common.status')} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="active">{tCommon('status.active')}</SelectItem>
                        <SelectItem value="disabled">{tCommon('status.disabled')}</SelectItem>
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      {t('edit.status')}
                    </p>
                    {editErrors.status && (
                      <p className="text-sm text-red-500">{editErrors.status.message}</p>
                    )}
                  </div>
                </div>
              </div>
              <DialogFooter className="flex-shrink-0">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setIsEditOpen(false);
                    setEditingPeer(null);
                    resetEdit();
                  }}
                >
                  {tCommon('buttons.cancel')}
                </Button>
                <Button type="submit" disabled={isEditSubmitting}>
                  {isEditSubmitting ? tCommon('status.loading') : t('edit.updateButton')}
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>

      {/* View Config Dialog */}
      <Dialog open={isViewConfigOpen} onOpenChange={(open) => {
        setIsViewConfigOpen(open);
        if (!open) {
          setViewingConfig(null);
        }
      }}>
        <DialogContent className="max-w-3xl max-h-[90vh] flex flex-col">
          <DialogHeader className="flex-shrink-0">
            <DialogTitle>{tCommon('buttons.view')} {tCommon('common.name')} - {viewingConfig?.peerName || t('title')}</DialogTitle>
            <DialogDescription>
              {tCommon('buttons.view')} {t('title')} {tCommon('common.description')}
            </DialogDescription>
          </DialogHeader>
          <div className="flex-1 flex flex-col min-h-0">
            <div className="flex-1 overflow-auto bg-muted p-4 rounded">
              <pre className="font-mono text-sm whitespace-pre-wrap break-words">
                {viewingConfig?.config || tCommon('status.loading')}
              </pre>
            </div>
          </div>
          <DialogFooter className="flex-shrink-0">
            <Button
              type="button"
              variant="outline"
              onClick={handleCopyConfig}
              disabled={!viewingConfig}
            >
              <Copy className="mr-2 h-4 w-4" /> {tCommon('buttons.copy')}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                if (viewingConfig) {
                  handleDownloadConfig(viewingConfig.peerId);
                }
              }}
              disabled={!viewingConfig}
            >
              <Download className="mr-2 h-4 w-4" /> {tCommon('buttons.download')}
            </Button>
            <Button
              type="button"
              onClick={() => setIsViewConfigOpen(false)}
            >
              {tCommon('buttons.close')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Batch Create Dialog */}
      <Dialog open={isBatchCreateOpen} onOpenChange={setIsBatchCreateOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{t('batch.createTitle')}</DialogTitle>
            <DialogDescription>
              {t('batch.createDescription', { max: 50 })}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {batchItems.map((item, index) => (
              <div key={index} className="p-4 border rounded-md space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="font-medium">{t('batch.item', { index: index + 1 })}</h4>
                  {batchItems.length > 1 && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeBatchItem(index)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label>{t('create.deviceName')} *</Label>
                    <Input
                      value={item.device_name || ''}
                      onChange={(e) => updateBatchItem(index, 'device_name', e.target.value)}
                      placeholder={t('create.deviceNamePlaceholder')}
                    />
                  </div>
                  {isAdmin && (
                    <div className="space-y-2">
                      <Label>{t('create.user')} *</Label>
                      <Select
                        value={item.username || ''}
                        onValueChange={(value) => updateBatchItem(index, 'username', value)}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder={t('create.selectUser')} />
                        </SelectTrigger>
                        <SelectContent>
                          {users.map((user) => (
                            <SelectItem key={user.username} value={user.username}>
                              {user.username} {user.nickname && `(${user.nickname})`}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label>{t('create.ipPool')}</Label>
                    <Select
                      value={item.ip_pool_id || '__none__'}
                      onValueChange={(value) => updateBatchItem(index, 'ip_pool_id', value === '__none__' ? undefined : value)}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t('create.selectIPPool')} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__none__">{tCommon('common.none')}</SelectItem>
                        {pools.map((pool) => (
                          <SelectItem key={pool.id} value={pool.id}>
                            {pool.name} ({pool.cidr})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>{t('create.clientIP')}</Label>
                    <Input
                      value={item.client_ip || ''}
                      onChange={(e) => updateBatchItem(index, 'client_ip', e.target.value)}
                      placeholder={t('create.clientIPPlaceholder')}
                    />
                  </div>
                </div>
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              onClick={addBatchItem}
              disabled={batchItems.length >= 50}
            >
              <Plus className="mr-2 h-4 w-4" /> {t('batch.addItem')}
            </Button>
            <div className="text-sm text-muted-foreground">
              {t('batch.itemsCount', { current: batchItems.length, max: 50 })}
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setIsBatchCreateOpen(false);
                setBatchItems([]);
              }}
            >
              {tCommon('buttons.cancel')}
            </Button>
            <Button type="button" onClick={handleBatchCreateSubmit}>
              {t('batch.createButton')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Batch Edit Dialog */}
      <Dialog open={isBatchEditOpen} onOpenChange={setIsBatchEditOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('batch.editTitle', { count: selectedPeers.size })}</DialogTitle>
            <DialogDescription>
              {t('batch.editDescription', { count: selectedPeers.size })}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleEditSubmit((data) => handleBatchEditSubmit(data))} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="batch-status">{t('edit.status')}</Label>
              <Select
                value={watchEdit("status") || ""}
                onValueChange={(value) => setEditValue("status", value as "active" | "disabled")}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('batch.selectStatus')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">{tCommon('status.active')}</SelectItem>
                  <SelectItem value="disabled">{tCommon('status.disabled')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsBatchEditOpen(false);
                }}
              >
                {tCommon('buttons.cancel')}
              </Button>
              <Button type="submit" disabled={isEditSubmitting}>
                {isEditSubmitting ? t('batch.updating') : t('batch.saveButton')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Batch Delete Dialog */}
      <Dialog open={isBatchDeleteOpen} onOpenChange={setIsBatchDeleteOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('batch.deleteTitle')}</DialogTitle>
            <DialogDescription>
              {t('batch.deleteDescription', { count: selectedPeers.size })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setIsBatchDeleteOpen(false);
              }}
            >
              {tCommon('buttons.cancel')}
            </Button>
            <Button
              type="button"
              variant="destructive"
              onClick={confirmBatchDelete}
            >
              {t('batch.deleteButton')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
