import React, { useEffect, useState } from "react";
import { api, IPPool } from "../services/api";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { Progress } from "../components/ui/progress";
import { toast } from "sonner";
import { Database } from "lucide-react";

export function IPPools() {
  const [pools, setPools] = useState<IPPool[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPools = async () => {
        try {
            const data = await api.wg.listIPPools();
            setPools(data);
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
              <TableHead>CIDR</TableHead>
              <TableHead>Usage</TableHead>
              <TableHead>Available</TableHead>
              <TableHead>Total</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
                 <TableRow>
                    <TableCell colSpan={4} className="text-center py-8">Loading...</TableCell>
                </TableRow>
            ) : (
                pools.map((pool) => {
                const percentage = (pool.used_ips / pool.total_ips) * 100;
                return (
                    <TableRow key={pool.id}>
                    <TableCell className="font-medium flex items-center gap-2">
                         <Database className="h-4 w-4 text-blue-500" />
                        {pool.cidr}
                    </TableCell>
                    <TableCell className="w-[300px]">
                        <div className="space-y-1">
                            <div className="flex justify-between text-xs text-muted-foreground">
                                <span>{percentage.toFixed(1)}% Used</span>
                                <span>{pool.used_ips} IPs</span>
                            </div>
                            <Progress value={percentage} className="h-2" />
                        </div>
                    </TableCell>
                    <TableCell>{pool.available_ips}</TableCell>
                    <TableCell>{pool.total_ips}</TableCell>
                    </TableRow>
                );
                })
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
