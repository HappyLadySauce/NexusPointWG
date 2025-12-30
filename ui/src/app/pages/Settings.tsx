import { AlertCircle } from "lucide-react";
import React from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";

export function Settings() {
  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <h1 className="text-3xl font-bold tracking-tight">Settings</h1>

      <div className="max-w-2xl">
        <Card>
          <CardHeader>
            <CardTitle>Global Settings</CardTitle>
            <CardDescription>
              Configure global settings for the WireGuard interface and generated peer configs.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-3 p-4 bg-muted rounded-lg">
              <AlertCircle className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="font-medium">Feature Not Available</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Global settings management is not yet implemented in the backend API.
                  WireGuard configuration is managed through the server configuration file.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
