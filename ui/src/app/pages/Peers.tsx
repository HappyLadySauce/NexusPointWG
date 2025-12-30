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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { api, CreateWGPeerRequest, WGPeerResponse } from "../services/api";

const peerSchema = z.object({
  device_name: z.string().min(1, "Device name is required"),
});

export function Peers() {
  const [peers, setPeers] = useState<WGPeerResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<{ device_name: string }>({
    resolver: zodResolver(peerSchema),
  });

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
  }, []);

  const onSubmit = async (data: { device_name: string }) => {
    try {
      const request: CreateWGPeerRequest = {
        device_name: data.device_name,
      };
      await api.wg.createPeer(request);
      toast.success("Peer created successfully");
      setIsCreateOpen(false);
      reset();
      fetchPeers();
    } catch (error) {
      toast.error("Failed to create peer");
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
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Peer</DialogTitle>
              <DialogDescription>
                Add a new device to the WireGuard network. Keys and IPs will be generated automatically.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="device_name">Device Name</Label>
                <Input id="device_name" placeholder="e.g. My iPhone" {...register("device_name")} />
                {errors.device_name && <p className="text-sm text-red-500">{errors.device_name.message}</p>}
              </div>
              <DialogFooter>
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
