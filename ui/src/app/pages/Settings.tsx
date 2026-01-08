import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeOff, Loader2 } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import * as z from "zod";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Textarea } from "../components/ui/textarea";
import { useAuth } from "../context/AuthContext";
import { api, GetServerConfigResponse, UpdateServerConfigRequest } from "../services/api";

// Form schema for server configuration
const serverConfigSchema = z.object({
  address: z.string().min(1, "Address is required").regex(
    /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/,
    "Invalid CIDR format (e.g., 100.100.100.1/24)"
  ),
  listen_port: z.number().min(1, "Port must be at least 1").max(65535, "Port must be at most 65535"),
  private_key: z.string().min(1, "Private key is required"),
  mtu: z.number().min(68, "MTU must be at least 68").max(65535, "MTU must be at most 65535"),
  post_up: z.string().max(1000, "PostUp command must be at most 1000 characters").optional().or(z.literal("")),
  post_down: z.string().max(1000, "PostDown command must be at most 1000 characters").optional().or(z.literal("")),
  server_ip: z.string().regex(
    /^(\d{1,3}\.){3}\d{1,3}$/,
    "Invalid IPv4 address format"
  ).optional().or(z.literal("")),
  dns: z.string().optional().or(z.literal("")).refine(
    (val) => {
      if (!val || val.trim() === "") return true;
      // Validate comma-separated IP addresses
      const parts = val.split(",");
      const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/;
      return parts.every((part) => {
        const trimmed = part.trim();
        return trimmed === "" || ipRegex.test(trimmed);
      });
    },
    { message: "Invalid DNS format (e.g., 1.1.1.1, 8.8.8.8)" }
  ),
});

type ServerConfigFormValues = z.infer<typeof serverConfigSchema>;

export function Settings() {
  const { user: currentUser } = useAuth();
  const { t } = useTranslation('settings');
  const { t: tCommon } = useTranslation('common');
  const [loading, setLoading] = useState(true);
  const [config, setConfig] = useState<GetServerConfigResponse | null>(null);
  const [showPrivateKey, setShowPrivateKey] = useState(false);
  const [initialValues, setInitialValues] = useState<ServerConfigFormValues | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
    watch,
  } = useForm<ServerConfigFormValues>({
    resolver: zodResolver(serverConfigSchema),
  });

  const isAdmin = currentUser?.role === "admin";
  const formValues = watch();

  // Fetch server configuration
  const fetchConfig = async () => {
    if (!isAdmin) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      const response = await api.wg.getServerConfig();
      setConfig(response);
      const formData: ServerConfigFormValues = {
        address: response.address,
        listen_port: response.listen_port,
        private_key: response.private_key,
        mtu: response.mtu,
        post_up: response.post_up || "",
        post_down: response.post_down || "",
        server_ip: response.server_ip || "",
        dns: response.dns || "",
      };
      setInitialValues(formData);
      reset(formData);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('messages.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConfig();
  }, [isAdmin]);

  // Check if form has changes
  const hasChanges = () => {
    if (!initialValues) return false;
    return (
      formValues.address !== initialValues.address ||
      formValues.listen_port !== initialValues.listen_port ||
      formValues.private_key !== initialValues.private_key ||
      formValues.mtu !== initialValues.mtu ||
      formValues.post_up !== initialValues.post_up ||
      formValues.post_down !== initialValues.post_down ||
      formValues.server_ip !== initialValues.server_ip ||
      formValues.dns !== initialValues.dns
    );
  };

  // Handle form submission
  const onSubmit = async (data: ServerConfigFormValues) => {
    if (!isAdmin) {
      toast.error(t('messages.permissionDenied'));
      return;
    }

    try {
      // Build update request with only changed fields
      const request: UpdateServerConfigRequest = {};
      if (initialValues) {
        if (data.address !== initialValues.address) {
          request.address = data.address;
        }
        if (data.listen_port !== initialValues.listen_port) {
          request.listen_port = data.listen_port;
        }
        if (data.private_key !== initialValues.private_key) {
          request.private_key = data.private_key;
        }
        if (data.mtu !== initialValues.mtu) {
          request.mtu = data.mtu;
        }
        if (data.post_up !== initialValues.post_up) {
          request.post_up = data.post_up || undefined;
        }
        if (data.post_down !== initialValues.post_down) {
          request.post_down = data.post_down || undefined;
        }
        if (data.server_ip !== initialValues.server_ip) {
          request.server_ip = data.server_ip || undefined;
        }
        if (data.dns !== initialValues.dns) {
          request.dns = data.dns || undefined;
        }
      } else {
        // If no initial values, send all fields
        request.address = data.address;
        request.listen_port = data.listen_port;
        request.private_key = data.private_key;
        request.mtu = data.mtu;
        request.post_up = data.post_up || undefined;
        request.post_down = data.post_down || undefined;
        request.server_ip = data.server_ip || undefined;
        request.dns = data.dns || undefined;
      }

      await api.wg.updateServerConfig(request);
      toast.success(t('messages.updateSuccess'));

      // Reload configuration
      await fetchConfig();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('messages.updateFailed'));
    }
  };

  if (!isAdmin) {
    return (
      <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
        <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
        <div className="max-w-2xl">
          <Card>
            <CardHeader>
              <CardTitle>{t('globalSettings.title')}</CardTitle>
              <CardDescription>
                {t('globalSettings.description')}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-3 p-4 bg-muted rounded-lg">
                <p className="text-sm text-muted-foreground">
                  {t('globalSettings.noPermission')}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
        <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
        <div className="max-w-2xl">
          <Card>
            <CardContent className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>

      <div className="max-w-2xl space-y-6">
        {/* Server Configuration */}
        <Card>
          <CardHeader>
            <CardTitle>{t('globalSettings.title')}</CardTitle>
            <CardDescription>
              {t('globalSettings.description')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* Address */}
              <div className="space-y-2">
                <Label htmlFor="address">{t('serverConfig.address')} *</Label>
                <Input
                  id="address"
                  placeholder={t('serverConfig.addressPlaceholder')}
                  {...register("address")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.addressHint')}
                </p>
                {errors.address && (
                  <p className="text-sm text-red-500">{errors.address.message}</p>
                )}
              </div>

              {/* Listen Port */}
              <div className="space-y-2">
                <Label htmlFor="listen_port">{t('serverConfig.listenPort')} *</Label>
                <Input
                  id="listen_port"
                  type="number"
                  placeholder={t('serverConfig.listenPortPlaceholder')}
                  {...register("listen_port", { valueAsNumber: true })}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.listenPortHint')}
                </p>
                {errors.listen_port && (
                  <p className="text-sm text-red-500">{errors.listen_port.message}</p>
                )}
              </div>

              {/* Private Key */}
              <div className="space-y-2">
                <Label htmlFor="private_key">{t('serverConfig.privateKey')} *</Label>
                <div className="relative">
                  <Input
                    id="private_key"
                    type={showPrivateKey ? "text" : "password"}
                    placeholder={t('serverConfig.privateKeyPlaceholder')}
                    {...register("private_key")}
                    className="pr-10"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowPrivateKey(!showPrivateKey)}
                  >
                    {showPrivateKey ? (
                      <EyeOff className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Eye className="h-4 w-4 text-muted-foreground" />
                    )}
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.privateKeyHint')}
                </p>
                {errors.private_key && (
                  <p className="text-sm text-red-500">{errors.private_key.message}</p>
                )}
              </div>

              {/* Public Key (Read-only) */}
              {config && (
                <div className="space-y-2">
                  <Label htmlFor="public_key">{t('serverConfig.publicKey')}</Label>
                  <Input
                    id="public_key"
                    value={config.public_key}
                    readOnly
                    className="bg-muted"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('serverConfig.publicKeyHint')}
                  </p>
                </div>
              )}

              {/* MTU */}
              <div className="space-y-2">
                <Label htmlFor="mtu">{t('serverConfig.mtu')} *</Label>
                <Input
                  id="mtu"
                  type="number"
                  placeholder={t('serverConfig.mtuPlaceholder')}
                  {...register("mtu", { valueAsNumber: true })}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.mtuHint')}
                </p>
                {errors.mtu && (
                  <p className="text-sm text-red-500">{errors.mtu.message}</p>
                )}
              </div>

              {/* PostUp */}
              <div className="space-y-2">
                <Label htmlFor="post_up">{t('serverConfig.postUp')}</Label>
                <Textarea
                  id="post_up"
                  placeholder={t('serverConfig.postUpPlaceholder')}
                  rows={3}
                  {...register("post_up")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.postUpHint')}
                </p>
                {errors.post_up && (
                  <p className="text-sm text-red-500">{errors.post_up.message}</p>
                )}
              </div>

              {/* PostDown */}
              <div className="space-y-2">
                <Label htmlFor="post_down">{t('serverConfig.postDown')}</Label>
                <Textarea
                  id="post_down"
                  placeholder={t('serverConfig.postDownPlaceholder')}
                  rows={3}
                  {...register("post_down")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.postDownHint')}
                </p>
                {errors.post_down && (
                  <p className="text-sm text-red-500">{errors.post_down.message}</p>
                )}
              </div>

              {/* Server IP */}
              <div className="space-y-2">
                <Label htmlFor="server_ip">{t('serverConfig.serverIP')}</Label>
                <Input
                  id="server_ip"
                  placeholder={t('serverConfig.serverIPPlaceholder')}
                  {...register("server_ip")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.serverIPHint')}
                </p>
                {errors.server_ip && (
                  <p className="text-sm text-red-500">{errors.server_ip.message}</p>
                )}
              </div>

              {/* DNS */}
              <div className="space-y-2">
                <Label htmlFor="dns">{t('serverConfig.dns')}</Label>
                <Input
                  id="dns"
                  placeholder={t('serverConfig.dnsPlaceholder')}
                  {...register("dns")}
                />
                <p className="text-xs text-muted-foreground">
                  {t('serverConfig.dnsHint')}
                </p>
                {errors.dns && (
                  <p className="text-sm text-red-500">{errors.dns.message}</p>
                )}
              </div>

              {/* Submit Button */}
              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    if (initialValues) {
                      reset(initialValues);
                    }
                  }}
                  disabled={!hasChanges() || isSubmitting}
                >
                  {t('serverConfig.resetButton')}
                </Button>
                <Button type="submit" disabled={!hasChanges() || isSubmitting}>
                  {isSubmitting ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {t('serverConfig.updating')}
                    </>
                  ) : (
                    t('serverConfig.updateButton')
                  )}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
