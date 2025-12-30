import { zodResolver } from "@hookform/resolvers/zod";
import { Edit, Plus, Trash2, User as UserIcon } from "lucide-react";
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
import { api, CreateUserRequest, UpdateUserRequest, UserResponse } from "../services/api";

// Form schema for creating user
const createUserSchema = z.object({
  username: z.string()
    .min(3, "Username must be at least 3 characters")
    .max(32, "Username must be at most 32 characters")
    .refine((val) => /^[a-zA-Z0-9_-]+$/.test(val), {
      message: "Username can only contain letters, numbers, underscores, and hyphens",
    })
    .refine((val) => !/[\u4e00-\u9fa5]/.test(val), {
      message: "Username cannot contain Chinese characters",
    }),
  nickname: z.string().max(32, "Nickname must be at most 32 characters").optional().or(z.literal("")),
  email: z.string()
    .email("Invalid email address")
    .max(255, "Email must be at most 255 characters")
    .refine((val) => {
      const allowedDomains = ["qq.com", "163.com", "gmail.com", "outlook.com"];
      const parts = val.split("@");
      if (parts.length !== 2) return false;
      const domain = parts[1]?.toLowerCase();
      return allowedDomains.includes(domain);
    }, {
      message: "Email must use one of the allowed domains: qq.com, 163.com, gmail.com, outlook.com",
    }),
  password: z.string().min(8, "Password must be at least 8 characters").max(32, "Password must be at most 32 characters"),
  confirmPassword: z.string(),
  role: z.enum(["user", "admin"]).optional(),
  status: z.enum(["active", "inactive"]).optional(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords do not match",
  path: ["confirmPassword"],
}).refine((data) => !data.nickname || data.nickname.length === 0 || data.nickname.length >= 3, {
  message: "Nickname must be at least 3 characters if provided",
  path: ["nickname"],
});

type CreateUserFormValues = z.infer<typeof createUserSchema>;

// Form schema for updating user
const updateUserSchema = z.object({
  username: z.string()
    .min(3, "Username must be at least 3 characters")
    .max(32, "Username must be at most 32 characters")
    .refine((val) => /^[a-zA-Z0-9_-]+$/.test(val), {
      message: "Username can only contain letters, numbers, underscores, and hyphens",
    })
    .refine((val) => !/[\u4e00-\u9fa5]/.test(val), {
      message: "Username cannot contain Chinese characters",
    })
    .optional(),
  nickname: z.string().max(32, "Nickname must be at most 32 characters").optional().or(z.literal("")),
  email: z.string()
    .email("Invalid email address")
    .max(255, "Email must be at most 255 characters")
    .refine((val) => {
      const allowedDomains = ["qq.com", "163.com", "gmail.com", "outlook.com"];
      const parts = val.split("@");
      if (parts.length !== 2) return false;
      const domain = parts[1]?.toLowerCase();
      return allowedDomains.includes(domain);
    }, {
      message: "Email must use one of the allowed domains: qq.com, 163.com, gmail.com, outlook.com",
    })
    .optional(),
  password: z.string().min(8, "Password must be at least 8 characters").max(32, "Password must be at most 32 characters").optional().or(z.literal("")),
  confirmPassword: z.string().optional(),
  role: z.enum(["user", "admin"]).optional(),
  status: z.enum(["active", "inactive", "deleted"]).optional(),
}).refine((data) => {
  // If password is provided, confirmPassword must match
  if (data.password && data.password !== "") {
    return data.password === data.confirmPassword;
  }
  return true;
}, {
  message: "Passwords do not match",
  path: ["confirmPassword"],
}).refine((data) => !data.nickname || data.nickname.length === 0 || data.nickname.length >= 3, {
  message: "Nickname must be at least 3 characters if provided",
  path: ["nickname"],
});

type UpdateUserFormValues = z.infer<typeof updateUserSchema>;

export function UsersPage() {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserResponse | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingUser, setDeletingUser] = useState<UserResponse | null>(null);

  const { register, handleSubmit, reset, formState: { errors, isSubmitting }, watch, setValue } = useForm<CreateUserFormValues>({
    resolver: zodResolver(createUserSchema),
    defaultValues: {
      role: "user",
      status: "active",
    },
  });

  const { register: registerEdit, handleSubmit: handleEditSubmit, reset: resetEdit, formState: { errors: editErrors, isSubmitting: isEditSubmitting }, watch: watchEdit, setValue: setEditValue } = useForm<UpdateUserFormValues>({
    resolver: zodResolver(updateUserSchema),
  });

  const isAdmin = currentUser?.role === "admin";

  const fetchUsers = async () => {
    try {
      const response = await api.users.list();
      setUsers(response.items || []);
    } catch (e) {
      toast.error("Failed to load users");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const onSubmit = async (data: CreateUserFormValues) => {
    try {
      const request: CreateUserRequest = {
        username: data.username,
        nickname: data.nickname || undefined,
        email: data.email,
        password: data.password,
      };

      // Only include role and status if user is admin
      // Status is already limited to "active" or "inactive" in schema
      if (isAdmin) {
        if (data.role) {
          request.role = data.role;
        }
        if (data.status) {
          request.status = data.status;
        }
      }

      await api.users.create(request);
      toast.success("User created successfully");
      setIsCreateOpen(false);
      reset();
      fetchUsers();
    } catch (error: any) {
      const errorMessage = error?.message || "Failed to create user";
      toast.error(errorMessage);
    }
  };

  const handleEdit = async (user: UserResponse) => {
    try {
      // Get full user details
      const userDetails = await api.users.get(user.username);
      setEditingUser(userDetails);

      // Populate form with existing data
      setEditValue("username", userDetails.username);
      setEditValue("nickname", userDetails.nickname || "");
      setEditValue("email", userDetails.email);
      setEditValue("password", "");
      setEditValue("confirmPassword", "");

      // Note: role and status are not in UserResponse, so we can't pre-fill them
      // Admin can still set them in the form
      if (isAdmin) {
        setEditValue("role", "user"); // Default, admin can change
        setEditValue("status", "active"); // Default, admin can change
      }

      setIsEditOpen(true);
    } catch (error: any) {
      toast.error("Failed to load user details");
    }
  };

  const onEditSubmit = async (data: UpdateUserFormValues) => {
    if (!editingUser) return;

    try {
      const request: UpdateUserRequest = {};

      // Only include fields that have changed
      if (data.username && data.username !== editingUser.username) {
        request.username = data.username;
      }
      if (data.nickname !== (editingUser.nickname || "")) {
        request.nickname = data.nickname || undefined;
      }
      if (data.email && data.email !== editingUser.email) {
        request.email = data.email;
      }
      if (data.password && data.password !== "") {
        request.password = data.password;
      }

      // Only include role and status if user is admin
      if (isAdmin) {
        if (data.role) {
          request.role = data.role;
        }
        if (data.status) {
          request.status = data.status;
        }
      }

      await api.users.update(editingUser.username, request);
      toast.success("User updated successfully");
      setIsEditOpen(false);
      setEditingUser(null);
      resetEdit();
      fetchUsers();
    } catch (error: any) {
      const errorMessage = error?.message || "Failed to update user";
      toast.error(errorMessage);
    }
  };

  const handleDelete = (user: UserResponse) => {
    setDeletingUser(user);
    setIsDeleteOpen(true);
  };

  const confirmDelete = async () => {
    if (!deletingUser) return;

    try {
      await api.users.delete(deletingUser.username);
      toast.success("User deleted successfully");
      setIsDeleteOpen(false);
      setDeletingUser(null);
      fetchUsers();
    } catch (error: any) {
      const errorMessage = error?.message || "Failed to delete user";
      toast.error(errorMessage);
    }
  };

  const canEditUser = (user: UserResponse) => {
    // Admin can edit anyone, regular users can only edit themselves
    return isAdmin || currentUser?.username === user.username;
  };

  const canDeleteUser = (user: UserResponse) => {
    // Admin can delete anyone, regular users can only delete themselves
    return isAdmin || currentUser?.username === user.username;
  };

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Users</h1>
        {isAdmin && (
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" /> Create User
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Create New User</DialogTitle>
                <DialogDescription>
                  Add a new user to the system. Fill in the required information below.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="username">Username *</Label>
                  <Input
                    id="username"
                    placeholder="e.g. johndoe"
                    {...register("username")}
                  />
                  {errors.username && (
                    <p className="text-sm text-red-500">{errors.username.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="nickname">Nickname</Label>
                  <Input
                    id="nickname"
                    placeholder="e.g. John Doe"
                    {...register("nickname")}
                  />
                  {errors.nickname && (
                    <p className="text-sm text-red-500">{errors.nickname.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="email">Email *</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="e.g. john@example.com"
                    {...register("email")}
                  />
                  {errors.email && (
                    <p className="text-sm text-red-500">{errors.email.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="password">Password *</Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder="At least 8 characters"
                    {...register("password")}
                  />
                  {errors.password && (
                    <p className="text-sm text-red-500">{errors.password.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirmPassword">Confirm Password *</Label>
                  <Input
                    id="confirmPassword"
                    type="password"
                    placeholder="Re-enter password"
                    {...register("confirmPassword")}
                  />
                  {errors.confirmPassword && (
                    <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>
                  )}
                </div>

                {isAdmin && (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="role">Role</Label>
                      <Select
                        value={watch("role") || "user"}
                        onValueChange={(value) => setValue("role", value as "user" | "admin")}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select role" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="user">User</SelectItem>
                          <SelectItem value="admin">Admin</SelectItem>
                        </SelectContent>
                      </Select>
                      {errors.role && (
                        <p className="text-sm text-red-500">{errors.role.message}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="status">Status</Label>
                      <Select
                        value={watch("status") || "active"}
                        onValueChange={(value) => setValue("status", value as "active" | "inactive")}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select status" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="active">Active</SelectItem>
                          <SelectItem value="inactive">Inactive</SelectItem>
                        </SelectContent>
                      </Select>
                      {errors.status && (
                        <p className="text-sm text-red-500">{errors.status.message}</p>
                      )}
                    </div>
                  </>
                )}

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
                    {isSubmitting ? "Creating..." : "Create User"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>
      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>Role</TableHead>
              <TableHead className="text-right">Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center py-8">Loading...</TableCell>
              </TableRow>
            ) : (
              users.map((user, index) => (
                <TableRow key={user.username || index}>
                  <TableCell className="font-medium flex items-center gap-2">
                    <div className="bg-slate-100 p-2 rounded-full">
                      <UserIcon className="h-4 w-4 text-slate-500" />
                    </div>
                    <div>
                      <div>{user.username}</div>
                      {user.nickname && user.nickname !== user.username && (
                        <div className="text-xs text-muted-foreground">{user.nickname}</div>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">User</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-emerald-600 text-sm font-medium">Active</span>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-2">
                      {canEditUser(user) && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(user)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                      )}
                      {canDeleteUser(user) && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(user)}
                        >
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Edit User Dialog */}
      <Dialog open={isEditOpen} onOpenChange={setIsEditOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
            <DialogDescription>
              Update user information. {isAdmin ? "You can modify all fields." : "You can only modify basic information."}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleEditSubmit(onEditSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-username">Username</Label>
              <Input
                id="edit-username"
                placeholder="e.g. johndoe"
                {...registerEdit("username")}
              />
              {editErrors.username && (
                <p className="text-sm text-red-500">{editErrors.username.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-nickname">Nickname</Label>
              <Input
                id="edit-nickname"
                placeholder="e.g. John Doe"
                {...registerEdit("nickname")}
              />
              {editErrors.nickname && (
                <p className="text-sm text-red-500">{editErrors.nickname.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-email">Email</Label>
              <Input
                id="edit-email"
                type="email"
                placeholder="e.g. john@example.com"
                {...registerEdit("email")}
              />
              {editErrors.email && (
                <p className="text-sm text-red-500">{editErrors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-password">Password (leave empty to keep current)</Label>
              <Input
                id="edit-password"
                type="password"
                placeholder="Leave empty to keep current password"
                {...registerEdit("password")}
              />
              {editErrors.password && (
                <p className="text-sm text-red-500">{editErrors.password.message}</p>
              )}
            </div>

            {watchEdit("password") && watchEdit("password") !== "" && (
              <div className="space-y-2">
                <Label htmlFor="edit-confirmPassword">Confirm Password</Label>
                <Input
                  id="edit-confirmPassword"
                  type="password"
                  placeholder="Re-enter password"
                  {...registerEdit("confirmPassword")}
                />
                {editErrors.confirmPassword && (
                  <p className="text-sm text-red-500">{editErrors.confirmPassword.message}</p>
                )}
              </div>
            )}

            {isAdmin && (
              <>
                <div className="space-y-2">
                  <Label htmlFor="edit-role">Role</Label>
                  <Select
                    value={watchEdit("role") || "user"}
                    onValueChange={(value) => setEditValue("role", value as "user" | "admin")}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="user">User</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                  {editErrors.role && (
                    <p className="text-sm text-red-500">{editErrors.role.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="edit-status">Status</Label>
                  <Select
                    value={watchEdit("status") || "active"}
                    onValueChange={(value) => setEditValue("status", value as "active" | "inactive" | "deleted")}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="active">Active</SelectItem>
                      <SelectItem value="inactive">Inactive</SelectItem>
                      <SelectItem value="deleted">Deleted</SelectItem>
                    </SelectContent>
                  </Select>
                  {editErrors.status && (
                    <p className="text-sm text-red-500">{editErrors.status.message}</p>
                  )}
                </div>
              </>
            )}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsEditOpen(false);
                  setEditingUser(null);
                  resetEdit();
                }}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isEditSubmitting}>
                {isEditSubmitting ? "Updating..." : "Update User"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete User Dialog */}
      <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Delete User</DialogTitle>
            <DialogDescription>
              {isAdmin ? (
                <>
                  Are you sure you want to permanently delete the user "{deletingUser?.username}"?
                  This action cannot be undone and will permanently remove the user from the system.
                </>
              ) : (
                <>
                  Are you sure you want to delete your account "{deletingUser?.username}"?
                  This will deactivate your account (soft delete). You can contact an administrator to restore it.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setIsDeleteOpen(false);
                setDeletingUser(null);
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
    </div>
  );
}
