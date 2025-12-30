import { Database } from "lucide-react";
import React, { useEffect, useState } from "react";
import { toast } from "sonner";
import { Badge } from "../components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { api, IPPoolResponse } from "../services/api";

export function IPPools() {
  const [pools, setPools] = useState<IPPoolResponse[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPools = async () => {
      try {
        const response = await api.wg.listIPPools();
        setPools(response.items || []);
      } catch (e) {
        toast.error("Failed to load IP pools");
      } finally {
        setLoading(false);
      }
    };
    fetchPools();
  }, []);

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <h1 className="text-3xl font-bold tracking-tight">IP Address Pools</h1>
      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>CIDR</TableHead>
              <TableHead>Server IP</TableHead>
              <TableHead>Gateway</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8">
                  Loading...
                </TableCell>
              </TableRow>
            ) : pools.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8">
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
                  <TableCell className="font-mono text-sm">{pool.server_ip}</TableCell>
                  <TableCell className="font-mono text-sm">
                    {pool.gateway || "N/A"}
                  </TableCell>
                  <TableCell>
                    {pool.status === "active" ? (
                      <Badge variant="default">Active</Badge>
                    ) : (
                      <Badge variant="secondary">Disabled</Badge>
                    )}
                  </TableCell>
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
    </div>
  );
}
