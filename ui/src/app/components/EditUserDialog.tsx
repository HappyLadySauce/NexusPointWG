import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import * as z from "zod";
import { Button } from "./ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { useAuth } from "../context/AuthContext";
import { api, UpdateUserRequest, User, UserResponse } from "../services/api";

// Form schema for updating user (only basic fields for self-edit)
const updateUserSchema = z.object({
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

interface EditUserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * EditUserDialog 组件
 * 用于编辑当前登录用户的信息
 */
export function EditUserDialog({ open, onOpenChange }: EditUserDialogProps) {
  const { user: currentUser, login } = useAuth();
  const { t } = useTranslation('users');
  const { t: tCommon } = useTranslation('common');
  const [loading, setLoading] = useState(false);
  const [userDetails, setUserDetails] = useState<UserResponse | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<UpdateUserFormValues>({
    resolver: zodResolver(updateUserSchema),
  });

  // Load user details when dialog opens
  useEffect(() => {
    if (open && currentUser) {
      loadUserDetails();
    }
  }, [open, currentUser]);

  const loadUserDetails = async () => {
    if (!currentUser) return;

    try {
      setLoading(true);
      const details = await api.users.get(currentUser.username);
      setUserDetails(details);
      reset({
        nickname: details.nickname || "",
        email: details.email,
        password: "",
        confirmPassword: "",
      });
    } catch (error: any) {
      toast.error(error?.message || t('messages.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  const onSubmit = async (data: UpdateUserFormValues) => {
    if (!currentUser) return;

    try {
      const request: UpdateUserRequest = {};

      // Only include fields that have changed
      if (data.nickname !== (userDetails?.nickname || "")) {
        request.nickname = data.nickname || undefined;
      }
      if (data.email && data.email !== userDetails?.email) {
        request.email = data.email;
      }
      if (data.password && data.password !== "") {
        request.password = data.password;
      }

      // Only send request if there are changes
      if (Object.keys(request).length === 0) {
        toast.info(tCommon('messages.updateSuccess'));
        onOpenChange(false);
        return;
      }

      const updatedUserResponse = await api.users.update(currentUser.username, request);
      toast.success(t('messages.updateSuccess'));

      // Update localStorage with updated user info
      // Convert UserResponse to User format for AuthContext
      const updatedUser: User = {
        id: currentUser.id,
        username: currentUser.username,
        role: currentUser.role,
        nickname: updatedUserResponse.nickname,
        email: updatedUserResponse.email,
      };
      localStorage.setItem("user", JSON.stringify(updatedUser));
      
      // Reload to update AuthContext
      window.location.reload();
    } catch (error: any) {
      const errorMessage = error?.message || t('messages.updateFailed');
      toast.error(errorMessage);
    }
  };

  if (!currentUser) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t('edit.title')}</DialogTitle>
          <DialogDescription>
            {tCommon('topBar.editProfile')}
          </DialogDescription>
        </DialogHeader>
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-username">{t('edit.username')}</Label>
              <Input
                id="edit-username"
                value={currentUser.username}
                readOnly
                className="bg-muted"
              />
              <p className="text-xs text-muted-foreground">
                {tCommon('common.username')} {tCommon('common.na')}
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-nickname">{t('edit.nickname')}</Label>
              <Input
                id="edit-nickname"
                placeholder={t('edit.nicknamePlaceholder')}
                {...register("nickname")}
              />
              {errors.nickname && (
                <p className="text-sm text-red-500">{errors.nickname.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-email">{t('edit.email')}</Label>
              <Input
                id="edit-email"
                type="email"
                placeholder={t('edit.emailPlaceholder')}
                {...register("email")}
              />
              {errors.email && (
                <p className="text-sm text-red-500">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-password">{t('edit.password')}</Label>
              <Input
                id="edit-password"
                type="password"
                placeholder={t('edit.passwordPlaceholder')}
                {...register("password")}
              />
              {errors.password && (
                <p className="text-sm text-red-500">{errors.password.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-confirm-password">{t('edit.confirmPassword')}</Label>
              <Input
                id="edit-confirm-password"
                type="password"
                placeholder={t('edit.confirmPasswordPlaceholder')}
                {...register("confirmPassword")}
              />
              {errors.confirmPassword && (
                <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>
              )}
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
              >
                {tCommon('buttons.cancel')}
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    {tCommon('status.loading')}
                  </>
                ) : (
                  t('edit.updateButton')
                )}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}

