import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { api, GlobalSettings } from "../services/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { toast } from "sonner";
import { Loader2, Save } from "lucide-react";

const settingsSchema = z.object({
  endpoint_address: z.string().min(1, "Endpoint address is required"),
  dns_servers: z.string().min(1, "At least one DNS server is required"), 
  mtu: z.coerce.number().min(576, "MTU must be at least 576"),
  keepalive: z.coerce.number().min(0, "Keepalive must be non-negative"),
  firewall_mark: z.coerce.number().nonnegative().default(0),
});

type SettingsFormValues = z.infer<typeof settingsSchema>;

export function Settings() {
  const [loading, setLoading] = useState(true);

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsSchema),
  });

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const data = await api.settings.get();
        reset({
            endpoint_address: data.endpoint_address,
            dns_servers: data.dns_servers.join(", "),
            mtu: data.mtu,
            keepalive: data.keepalive,
            firewall_mark: data.firewall_mark
        });
      } catch (error) {
        toast.error("Failed to load settings");
        console.error(error);
      } finally {
        setLoading(false);
      }
    };
    loadSettings();
  }, [reset]);

  const onSubmit = async (data: SettingsFormValues) => {
    try {
      const payload: Partial<GlobalSettings> = {
          endpoint_address: data.endpoint_address,
          dns_servers: data.dns_servers.split(",").map(s => s.trim()).filter(Boolean),
          mtu: data.mtu,
          keepalive: data.keepalive,
          firewall_mark: data.firewall_mark
      };

      await api.settings.update(payload);
      toast.success("Settings updated successfully");
    } catch (error) {
      toast.error("Failed to update settings");
      console.error(error);
    }
  };

  if (loading) {
      return <div className="flex items-center justify-center h-full"><Loader2 className="h-8 w-8 animate-spin" /></div>;
  }

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <h1 className="text-3xl font-bold tracking-tight">Global Settings</h1>
      
      <div className="max-w-2xl">
        <Card>
            <CardHeader>
                <CardTitle>WireGuard Configuration</CardTitle>
                <CardDescription>
                    Configure global settings for the WireGuard interface and generated peer configs.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                    <div className="space-y-2">
                        <Label htmlFor="endpoint">Endpoint Address</Label>
                        <Input 
                            id="endpoint" 
                            placeholder="vpn.example.com" 
                            {...register("endpoint_address")} 
                        />
                        <p className="text-sm text-muted-foreground">Public IP or domain name clients use to connect.</p>
                        {errors.endpoint_address && <p className="text-sm text-red-500">{errors.endpoint_address.message}</p>}
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="dns">DNS Servers</Label>
                        <Input 
                            id="dns" 
                            placeholder="1.1.1.1, 8.8.8.8" 
                            {...register("dns_servers")} 
                        />
                        <p className="text-sm text-muted-foreground">Comma-separated list of DNS servers for clients.</p>
                        {errors.dns_servers && <p className="text-sm text-red-500">{errors.dns_servers.message}</p>}
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="mtu">MTU</Label>
                            <Input 
                                id="mtu" 
                                type="number" 
                                {...register("mtu")} 
                            />
                            {errors.mtu && <p className="text-sm text-red-500">{errors.mtu.message}</p>}
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="keepalive">Persistent Keepalive (seconds)</Label>
                            <Input 
                                id="keepalive" 
                                type="number" 
                                {...register("keepalive")} 
                            />
                            {errors.keepalive && <p className="text-sm text-red-500">{errors.keepalive.message}</p>}
                        </div>
                    </div>

                     <div className="space-y-2">
                        <Label htmlFor="fwmark">Firewall Mark</Label>
                        <Input 
                            id="fwmark" 
                            type="number" 
                            {...register("firewall_mark")} 
                        />
                        <p className="text-sm text-muted-foreground">Optional fwmark for routing (0 to disable).</p>
                        {errors.firewall_mark && <p className="text-sm text-red-500">{errors.firewall_mark.message}</p>}
                    </div>

                    <Button type="submit" disabled={isSubmitting}>
                        {isSubmitting ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                            <Save className="mr-2 h-4 w-4" />
                        )}
                        Save Changes
                    </Button>
                </form>
            </CardContent>
        </Card>
      </div>
    </div>
  );
}
