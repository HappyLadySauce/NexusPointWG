import React, { useEffect, useState } from "react";
import { api, User } from "../services/api";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { Badge } from "../components/ui/badge";
import { toast } from "sonner";
import { Shield, User as UserIcon } from "lucide-react";

export function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUsers = async () => {
        try {
            const data = await api.users.list();
            setUsers(data);
        } catch (e) {
            toast.error("Failed to load users");
        } finally {
            setLoading(false);
        }
    };
    fetchUsers();
  }, []);

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <h1 className="text-3xl font-bold tracking-tight">Users</h1>
      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>Role</TableHead>
              <TableHead className="text-right">Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
                <TableRow>
                    <TableCell colSpan={3} className="text-center py-8">Loading...</TableCell>
                </TableRow>
            ) : (
                users.map((user) => (
                <TableRow key={user.id}>
                    <TableCell className="font-medium flex items-center gap-2">
                        <div className="bg-slate-100 p-2 rounded-full">
                            <UserIcon className="h-4 w-4 text-slate-500" />
                        </div>
                        {user.username}
                    </TableCell>
                    <TableCell>
                        {user.role === 'admin' ? (
                            <Badge className="bg-purple-500 hover:bg-purple-600">
                                <Shield className="w-3 h-3 mr-1" /> Admin
                            </Badge>
                        ) : (
                            <Badge variant="secondary">User</Badge>
                        )}
                    </TableCell>
                    <TableCell className="text-right">
                        <span className="text-emerald-600 text-sm font-medium">Active</span>
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
