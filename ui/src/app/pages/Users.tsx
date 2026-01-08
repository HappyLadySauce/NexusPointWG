import { zodResolver } from "@hookform/resolvers/zod";
import { Edit, Plus, Trash2, User as UserIcon } from "lucide-react";
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
  const { t } = useTranslation('users');
  const { t: tCommon } = useTranslation('common');
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserResponse | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingUser, setDeletingUser] = useState<UserResponse | null>(null);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

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
    setLoading(true);
    try {
      const offset = (currentPage - 1) * pageSize;
      const response = await api.users.list({
        offset,
        limit: pageSize,
      });
      setUsers(response.items || []);
      setTotal(response.total || 0);
    } catch (e) {
      toast.error(t('messages.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [currentPage]);

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
      toast.success(t('messages.createSuccess'));
      setIsCreateOpen(false);
      reset();
      // Refresh current page
      fetchUsers();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.createFailed');
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
      toast.error(t('messages.loadFailed'));
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
      toast.success(t('messages.updateSuccess'));
      setIsEditOpen(false);
      setEditingUser(null);
      resetEdit();
      // Refresh current page
      fetchUsers();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.updateFailed');
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
      toast.success(t('messages.deleteSuccess'));
      setIsDeleteOpen(false);
      setDeletingUser(null);

      // If we deleted the last item on the current page and it's not the first page, go to previous page
      if (users.length === 1 && currentPage > 1) {
        setCurrentPage(currentPage - 1);
      } else {
        fetchUsers();
      }
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.deleteFailed');
      toast.error(errorMessage);
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

  const canEditUser = (user: UserResponse) => {
    // Admin can edit anyone, regular users can only edit themselves
    return isAdmin || currentUser?.username === user.username;
  };

  const canDeleteUser = (user: UserResponse) => {
    // Admin can delete anyone, regular users can only delete themselves
    return isAdmin || currentUser?.username === user.username;
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
                  <Label htmlFor="username">{t('create.username')} *</Label>
                  <Input
                    id="username"
                    placeholder={t('create.usernamePlaceholder')}
                    {...register("username")}
                  />
                  {errors.username && (
                    <p className="text-sm text-red-500">{errors.username.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="nickname">{t('create.nickname')}</Label>
                  <Input
                    id="nickname"
                    placeholder={t('create.nicknamePlaceholder')}
                    {...register("nickname")}
                  />
                  {errors.nickname && (
                    <p className="text-sm text-red-500">{errors.nickname.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="email">{t('create.email')} *</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder={t('create.emailPlaceholder')}
                    {...register("email")}
                  />
                  {errors.email && (
                    <p className="text-sm text-red-500">{errors.email.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="password">{t('create.password')} *</Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder={t('create.passwordPlaceholder')}
                    {...register("password")}
                  />
                  {errors.password && (
                    <p className="text-sm text-red-500">{errors.password.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirmPassword">{t('create.confirmPassword')} *</Label>
                  <Input
                    id="confirmPassword"
                    type="password"
                    placeholder={t('create.confirmPasswordPlaceholder')}
                    {...register("confirmPassword")}
                  />
                  {errors.confirmPassword && (
                    <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>
                  )}
                </div>

                {isAdmin && (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="role">{t('create.role')}</Label>
                      <Select
                        value={watch("role") || "user"}
                        onValueChange={(value) => setValue("role", value as "user" | "admin")}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder={t('create.rolePlaceholder')} />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="user">{tCommon('common.user')}</SelectItem>
                          <SelectItem value="admin">{tCommon('common.admin')}</SelectItem>
                        </SelectContent>
                      </Select>
                      {errors.role && (
                        <p className="text-sm text-red-500">{errors.role.message}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="status">{t('create.status')}</Label>
                      <Select
                        value={watch("status") || "active"}
                        onValueChange={(value) => setValue("status", value as "active" | "inactive")}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder={t('create.statusPlaceholder')} />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="active">{tCommon('status.active')}</SelectItem>
                          <SelectItem value="inactive">{tCommon('status.inactive')}</SelectItem>
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
      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>Role</TableHead>
              {isAdmin && <TableHead className="text-right">Peers</TableHead>}
              <TableHead className="text-right">Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={isAdmin ? 5 : 4} className="text-center py-8">Loading...</TableCell>
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
                    <Badge variant={user.role === "admin" ? "default" : "secondary"}>
                      {user.role === "admin" ? "Admin" : "User"}
                    </Badge>
                  </TableCell>
                  {isAdmin && (
                    <TableCell className="text-right">
                      <span className="text-sm font-medium">{user.peer_count ?? 0}</span>
                    </TableCell>
                  )}
                  <TableCell className="text-right">
                    <span className={`text-sm font-medium ${
                      user.status === "active" ? "text-emerald-600" : 
                      user.status === "inactive" ? "text-yellow-600" : 
                      "text-red-600"
                    }`}>
                      {user.status === "active" ? "Active" : 
                       user.status === "inactive" ? "Inactive" : 
                       "Deleted"}
                    </span>
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
