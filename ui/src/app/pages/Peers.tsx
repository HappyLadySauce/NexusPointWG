import React, { useEffect, useState } from "react";
import { api, Peer } from "../services/api";
import { Button } from "../components/ui/button";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../components/ui/dialog";
import { Badge } from "../components/ui/badge";
import { Download, MoreHorizontal, Plus, Search, Trash2, Edit } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "../components/ui/dropdown-menu";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

const peerSchema = z.object({
  name: z.string().min(1, "Name is required"),
});

export function Peers() {
  const [peers, setPeers] = useState<Peer[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<{ name: string }>({
    resolver: zodResolver(peerSchema),
  });

  const fetchPeers = async () => {
    try {
      const data = await api.wg.listPeers();
      setPeers(data);
    } catch (error) {
      toast.error("Failed to fetch peers");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPeers();
  }, []);

  const onSubmit = async (data: { name: string }) => {
    try {
      await api.wg.createPeer(data);
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
            ? `${peer.name.replace(/[^a-zA-Z0-9_-]/g, '_')}.conf` 
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
    p.name.toLowerCase().includes(search.toLowerCase()) || 
    p.public_key.includes(search) ||
    p.allowed_ips.some(ip => ip.includes(search))
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
                <Label htmlFor="name">Device Name</Label>
                <Input id="name" placeholder="e.g. My iPhone" {...register("name")} />
                {errors.name && <p className="text-sm text-red-500">{errors.name.message}</p>}
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
                        <span>{peer.name}</span>
                        <span className="text-xs text-muted-foreground font-mono">{peer.public_key.substring(0, 10)}...</span>
                    </div>
                  </TableCell>
                  <TableCell>
                      {peer.allowed_ips.map(ip => (
                          <div key={ip} className="text-sm">{ip}</div>
                      ))}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{peer.endpoint || "N/A"}</TableCell>
                  <TableCell>
                      {peer.latest_handshake ? (
                          <span title={peer.latest_handshake}>{new Date(peer.latest_handshake).toLocaleString()}</span>
                      ) : (
                          <span className="text-muted-foreground">Never</span>
                      )}
                  </TableCell>
                  <TableCell>
                      {peer.status === 'active' ? (
                           <Badge variant="default" className="bg-emerald-500 hover:bg-emerald-600">Active</Badge>
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
