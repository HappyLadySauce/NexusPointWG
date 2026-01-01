import { zodResolver } from "@hookform/resolvers/zod";
import { Download, Edit, MoreHorizontal, Plus, Search, Trash2 } from "lucide-react";
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
  RadioGroup,
  RadioGroupItem,
} from "../components/ui/radio-group";
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
import { api, CreateWGPeerRequest, IPPoolResponse, UserResponse, WGPeerResponse } from "../services/api";

// Enhanced form schema
const peerSchema = z.object({
  username: z.string().optional(), // Admin can specify
  device_name: z.string().min(1, "Device name is required").max(64, "Device name must be at most 64 characters"),
  config_mode: z.enum(["pool", "manual"]),
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
  }, "Invalid Endpoint format (e.g., 118.24.41.142:51820)"),
  persistent_keepalive: z.number().min(0).max(65535).optional(),
}).refine((data) => {
  // Pool mode requires ip_pool_id
  if (data.config_mode === "pool" && !data.ip_pool_id) {
    return false;
  }
  return true;
}, {
  message: "IP Pool is required in Pool mode",
  path: ["ip_pool_id"],
}).refine((data) => {
  // Manual mode requires allowed_ips
  if (data.config_mode === "manual" && !data.allowed_ips) {
    return false;
  }
  return true;
}, {
  message: "Allowed IPs is required in Manual mode",
  path: ["allowed_ips"],
});

type PeerFormValues = z.infer<typeof peerSchema>;

export function Peers() {
  const { user: currentUser } = useAuth();
  const [peers, setPeers] = useState<WGPeerResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  // Enhanced state management
  const [pools, setPools] = useState<IPPoolResponse[]>([]);
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [selectedPool, setSelectedPool] = useState<string>("");

  const { register, handleSubmit, reset, formState: { errors, isSubmitting }, watch, setValue } = useForm<PeerFormValues>({
    resolver: zodResolver(peerSchema),
    defaultValues: {
      config_mode: "pool",
      persistent_keepalive: 25,
    },
  });

  const isAdmin = currentUser?.role === "admin";
  const configMode = watch("config_mode");

  const fetchPeers = async () => {
    try {
      const response = await api.wg.listPeers();
      setPeers(response.items || []);
    } catch (error) {
      toast.error("Failed to fetch peers");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPeers();
    fetchPools();
    if (isAdmin) {
      fetchUsers();
    }
  }, [isAdmin]);

  const fetchPools = async () => {
    try {
      const response = await api.wg.listIPPools({ status: "active" });
      setPools(response.items || []);
    } catch (error) {
      toast.error("Failed to load IP pools");
    }
  };

  const fetchUsers = async () => {
    try {
      const response = await api.users.list();
      setUsers(response.items || []);
    } catch (error) {
      toast.error("Failed to load users");
    }
  };

  const handlePoolChange = (poolID: string) => {
    setSelectedPool(poolID);
    setValue("ip_pool_id", poolID);

    // Auto-fill Pool configuration
    const pool = pools.find(p => p.id === poolID);
    if (pool && configMode === "pool") {
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

      // Pool mode
      if (data.config_mode === "pool" && data.ip_pool_id) {
        request.ip_pool_id = data.ip_pool_id;
        // Use Pool config, but allow manual override
        if (data.allowed_ips) request.allowed_ips = data.allowed_ips;
        if (data.dns) request.dns = data.dns;
        if (data.endpoint) request.endpoint = data.endpoint;
      } else {
        // Manual mode: all fields need to be manually filled
        if (data.ip_pool_id) request.ip_pool_id = data.ip_pool_id;
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
      toast.success("Peer created successfully");
      setIsCreateOpen(false);
      reset();
      setSelectedPool("");
      fetchPeers();
    } catch (error: any) {
      toast.error(error?.message || "Failed to create peer");
    }
  };

  const handleDelete = async (id: string) => {
    if (confirm("Are you sure you want to delete this peer?")) {
      try {
        await api.wg.deletePeer(id);
        toast.success("Peer deleted");
        fetchPeers();
      } catch (error) {
        toast.error("Failed to delete peer");
      }
    }
  };

  const handleDownloadConfig = async (id: string) => {
    try {
      toast.info("Downloading configuration...");
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

      toast.success("Config downloaded successfully");
    } catch (error) {
      toast.error("Failed to download config");
      console.error(error);
    }
  };

  const filteredPeers = peers.filter(p =>
    p.device_name.toLowerCase().includes(search.toLowerCase()) ||
    p.client_public_key.includes(search) ||
    p.client_ip.includes(search) ||
    p.allowed_ips.includes(search)
  );

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Peers</h1>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" /> Create Peer
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Create New Peer</DialogTitle>
              <DialogDescription>
                Add a new device to the WireGuard network. Choose between Pool auto-fill or manual configuration.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {/* Basic Information */}
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="device_name">Device Name *</Label>
                  <Input id="device_name" placeholder="e.g. My iPhone" {...register("device_name")} />
                  {errors.device_name && <p className="text-sm text-red-500">{errors.device_name.message}</p>}
                </div>

                {isAdmin && (
                  <div className="space-y-2">
                    <Label htmlFor="username">User *</Label>
                    <Select
                      value={watch("username") || ""}
                      onValueChange={(value) => setValue("username", value)}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select user" />
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

                {/* Configuration Mode */}
                <div className="space-y-2">
                  <Label>Configuration Mode</Label>
                  <RadioGroup
                    value={configMode}
                    onValueChange={(value) => {
                      setValue("config_mode", value as "pool" | "manual");
                      // Reset fields when switching modes
                      if (value === "manual") {
                        setValue("allowed_ips", "");
                        setValue("dns", "");
                        setValue("endpoint", "");
                      }
                    }}
                    className="flex gap-6"
                  >
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="pool" id="pool" />
                      <Label htmlFor="pool" className="font-normal cursor-pointer">Pool Auto-fill</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="manual" id="manual" />
                      <Label htmlFor="manual" className="font-normal cursor-pointer">Manual Configuration</Label>
                    </div>
                  </RadioGroup>
                </div>
              </div>

              {/* Pool Mode Fields */}
              {configMode === "pool" && (
                <div className="space-y-4 border-t pt-4">
                  <div className="space-y-2">
                    <Label htmlFor="ip_pool_id">IP Pool *</Label>
                    <Select
                      value={watch("ip_pool_id") || ""}
                      onValueChange={handlePoolChange}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select IP Pool" />
                      </SelectTrigger>
                      <SelectContent>
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
                    <Label htmlFor="client_ip">Client IP (optional)</Label>
                    <Input
                      id="client_ip"
                      placeholder="Enter IP manually or leave empty for auto-allocation (e.g., 100.100.100.2)"
                      {...register("client_ip")}
                    />
                    {errors.client_ip && <p className="text-sm text-red-500">{errors.client_ip.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="allowed_ips">Allowed IPs (auto-filled, can override)</Label>
                    <Input
                      id="allowed_ips"
                      placeholder="e.g., 0.0.0.0/0, 192.168.1.0/24"
                      {...register("allowed_ips")}
                    />
                    {errors.allowed_ips && <p className="text-sm text-red-500">{errors.allowed_ips.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="dns">DNS (auto-filled, can override)</Label>
                    <Input
                      id="dns"
                      placeholder="e.g., 1.1.1.1, 8.8.8.8"
                      {...register("dns")}
                    />
                    {errors.dns && <p className="text-sm text-red-500">{errors.dns.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="endpoint">Endpoint (auto-filled, can override)</Label>
                    <Input
                      id="endpoint"
                      placeholder="e.g., 118.24.41.142:51820"
                      {...register("endpoint")}
                    />
                    {errors.endpoint && <p className="text-sm text-red-500">{errors.endpoint.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="persistent_keepalive">Persistent Keepalive (seconds)</Label>
                    <Input
                      id="persistent_keepalive"
                      type="number"
                      min="0"
                      max="65535"
                      placeholder="25"
                      {...register("persistent_keepalive", { valueAsNumber: true })}
                    />
                    {errors.persistent_keepalive && <p className="text-sm text-red-500">{errors.persistent_keepalive.message}</p>}
                  </div>
                </div>
              )}

              {/* Manual Mode Fields */}
              {configMode === "manual" && (
                <div className="space-y-4 border-t pt-4">
                  <div className="space-y-2">
                    <Label htmlFor="ip_pool_id_manual">IP Pool (optional, for IP allocation)</Label>
                    <Select
                      value={watch("ip_pool_id") || ""}
                      onValueChange={(value) => {
                        setValue("ip_pool_id", value);
                        handlePoolChange(value);
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select IP Pool (optional)" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">None</SelectItem>
                        {pools.map((pool) => (
                          <SelectItem key={pool.id} value={pool.id}>
                            {pool.name} ({pool.cidr})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="client_ip_manual">Client IP (optional)</Label>
                    <Input
                      id="client_ip_manual"
                      placeholder="Enter IP manually or leave empty for auto-allocation (e.g., 100.100.100.2)"
                      {...register("client_ip")}
                    />
                    {errors.client_ip && <p className="text-sm text-red-500">{errors.client_ip.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="allowed_ips_manual">Allowed IPs *</Label>
                    <Input
                      id="allowed_ips_manual"
                      placeholder="e.g., 0.0.0.0/0, 192.168.1.0/24"
                      {...register("allowed_ips")}
                    />
                    {errors.allowed_ips && <p className="text-sm text-red-500">{errors.allowed_ips.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="dns_manual">DNS (optional)</Label>
                    <Input
                      id="dns_manual"
                      placeholder="e.g., 1.1.1.1, 8.8.8.8"
                      {...register("dns")}
                    />
                    {errors.dns && <p className="text-sm text-red-500">{errors.dns.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="endpoint_manual">Endpoint (optional)</Label>
                    <Input
                      id="endpoint_manual"
                      placeholder="e.g., 118.24.41.142:51820"
                      {...register("endpoint")}
                    />
                    {errors.endpoint && <p className="text-sm text-red-500">{errors.endpoint.message}</p>}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="persistent_keepalive_manual">Persistent Keepalive (seconds)</Label>
                    <Input
                      id="persistent_keepalive_manual"
                      type="number"
                      min="0"
                      max="65535"
                      placeholder="25"
                      {...register("persistent_keepalive", { valueAsNumber: true })}
                    />
                    {errors.persistent_keepalive && <p className="text-sm text-red-500">{errors.persistent_keepalive.message}</p>}
                  </div>
                </div>
              )}

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setIsCreateOpen(false);
                    reset();
                    setSelectedPool("");
                  }}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "Creating..." : "Create Peer"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="flex items-center space-x-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search peers..."
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
              <TableHead>Name</TableHead>
              <TableHead>IP Address</TableHead>
              <TableHead>Endpoint</TableHead>
              <TableHead>Latest Handshake</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8">Loading peers...</TableCell>
              </TableRow>
            ) : filteredPeers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8">No peers found.</TableCell>
              </TableRow>
            ) : (
              filteredPeers.map((peer) => (
                <TableRow key={peer.id}>
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
                  <TableCell className="text-muted-foreground text-sm">{peer.endpoint || "N/A"}</TableCell>
                  <TableCell>
                    <span className="text-muted-foreground">N/A</span>
                  </TableCell>
                  <TableCell>
                    {peer.status === 'active' ? (
                      <Badge variant="default">Active</Badge>
                    ) : (
                      <Badge variant="secondary">Inactive</Badge>
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
                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                        <DropdownMenuItem onClick={() => handleDownloadConfig(peer.id)}>
                          <Download className="mr-2 h-4 w-4" /> Download Config
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => toast.info("Edit not implemented in demo")}>
                          <Edit className="mr-2 h-4 w-4" /> Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-red-600" onClick={() => handleDelete(peer.id)}>
                          <Trash2 className="mr-2 h-4 w-4" /> Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
