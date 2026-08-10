"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { useState, useEffect, useCallback } from "react";
import { EventsOn, EventsOff } from "@/wailsjs/runtime/runtime";

import { toast } from "@/hooks/use-toast";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import {
  MessageFormat,
  MESSAGE_FORMATS,
  MESSAGE_FORMAT_OPTIONS,
} from "./form/types";

import {
  StartContinuousSend,
  StopContinuousSend,
  IsContinuousRunning,
  GetContinuousStats,
} from "@/wailsjs/go/main/StressTestService";
import {
  CheckConnection,
  Disconnect,
} from "@/wailsjs/go/main/ConnectionService";
import {
  PlayIcon,
  SquareIcon,
  ZapIcon,
  ClockIcon,
  GaugeIcon,
  SendIcon,
  AlertCircleIcon,
  ActivityIcon,
  CheckCircleIcon,
  XCircleIcon,
} from "lucide-react";

// IP validation regex
const ipRegex = /^(?:(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}|(?:[a-fA-F0-9]{1,4}:){7}[a-fA-F0-9]{1,4}|(?:[a-fA-F0-9]{1,4}:){1,7}:|(?:[a-fA-F0-9]{1,4}:){1,6}:[a-fA-F0-9]{1,4}|::(?:[a-fA-F0-9]{1,4}:){0,5}[a-fA-F0-9]{1,4}|[a-fA-F0-9]{1,4}::(?:[a-fA-F0-9]{1,4}:){0,4}[a-fA-F0-9]{1,4}|localhost)$/;

const StressTestSchema = z.object({
  Address: z.string().regex(ipRegex, { message: "Invalid IP address" }),
  Port: z.string().refine((val) => {
    const num = parseInt(val, 10);
    return !isNaN(num) && num >= 1 && num <= 65535;
  }, { message: "Port must be 1-65535" }),
  Protocol: z.string({ message: "Select protocol" }),
  Message: z.string().min(1, { message: "Message required" }),
  FramingMethod: z.enum(["octet-counting", "non-transparent"]),
  Facility: z.number().min(0).max(23),
  Severity: z.number().min(0).max(7),
  Hostname: z.string().optional(),
  Appname: z.string().min(1, "App name required"),
  MessageFormat: z.enum(MESSAGE_FORMATS),
  // Legacy format switch, kept in sync with MessageFormat.
  UseRFC5424: z.boolean(),
  UseTLS: z.boolean(),
  TLSVerify: z.boolean(),
  CACertPath: z.string().optional(),
  ClientCertPath: z.string().optional(),
  ClientKeyPath: z.string().optional(),
  // Stress test specific
  Duration: z.number().min(0).max(3600), // 0 = indefinite, max 1 hour
  MessagesPerSec: z.number().min(1).max(10000),
  MaxMessages: z.number().min(0), // 0 = no limit
  RandomizeData: z.boolean(),
});

interface ContinuousStats {
  totalSent: number;
  totalErrors: number;
  currentRate: number;
  elapsedSeconds: number;
  isRunning: boolean;
  startTime: number;
  targetRate: number;
  duration: number;
}

export function StressTestForm() {
  const [isRunning, setIsRunning] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [stats, setStats] = useState<ContinuousStats | null>(null);

  const form = useForm<z.infer<typeof StressTestSchema>>({
    resolver: zodResolver(StressTestSchema),
    defaultValues: {
      Address: "",
      Port: "514",
      Protocol: "tcp",
      Message: "Test syslog message from SendLog Stress Test",
      FramingMethod: "octet-counting",
      Facility: 16,
      Severity: 6,
      Hostname: "",
      Appname: "sendlog-stress",
      MessageFormat: "rfc5424",
      UseRFC5424: true,
      UseTLS: false,
      TLSVerify: false,
      CACertPath: "",
      ClientCertPath: "",
      ClientKeyPath: "",
      Duration: 30,
      MessagesPerSec: 100,
      MaxMessages: 0,
      RandomizeData: true,
    },
  });

  // In raw-pri mode the app only prepends <PRI>; the message supplies the header.
  const isRawPRI = form.watch("MessageFormat") === "raw-pri";

  // Event listeners
  useEffect(() => {
    const handleStats = (data: ContinuousStats) => {
      setStats(data);
      setIsRunning(data.isRunning);
    };

    const handleStarted = (data: ContinuousStats) => {
      setIsRunning(true);
      setStats(data);
      toast({
        title: "▶ Stress Test Started",
        description: `Sending ${data.targetRate} msg/sec`,
        variant: "default",
      });
    };

    const handleStopped = (data: ContinuousStats) => {
      setIsRunning(false);
      setStats(data);
      toast({
        title: "■ Stress Test Stopped",
        description: `Sent ${data.totalSent.toLocaleString()} messages in ${data.elapsedSeconds.toFixed(1)}s`,
        variant: "default",
      });
    };

    const handleError = (error: string) => {
      setIsRunning(false);
      toast({
        title: "✗ Error",
        description: error,
        variant: "destructive",
      });
    };

    // Subscribe to events
    EventsOn("continuous:stats", handleStats);
    EventsOn("continuous:started", handleStarted);
    EventsOn("continuous:stopped", handleStopped);
    EventsOn("continuous:error", handleError);

    // Check initial state
    IsContinuousRunning().then(setIsRunning);
    GetContinuousStats().then((s) => {
      if (s.isRunning) setStats(s);
    });

    // Cleanup
    return () => {
      EventsOff("continuous:stats");
      EventsOff("continuous:started");
      EventsOff("continuous:stopped");
      EventsOff("continuous:error");
    };
  }, []);

  const handleStart = useCallback(async () => {
    const data = form.getValues();
    const valid = await form.trigger();
    if (!valid) return;

    try {
      await StartContinuousSend({
        Address: data.Address,
        Port: data.Port,
        Protocol: data.Protocol,
        Message: data.Message,
        FramingMethod: data.FramingMethod,
        Facility: data.Facility,
        Severity: data.Severity,
        Hostname: data.Hostname || "",
        Appname: data.Appname,
        MessageFormat: data.MessageFormat,
        UseRFC5424: data.UseRFC5424,
        UseTLS: data.UseTLS,
        TLSVerify: data.TLSVerify,
        CACertPath: data.CACertPath || "",
        ClientCertPath: data.ClientCertPath || "",
        ClientKeyPath: data.ClientKeyPath || "",
        Duration: data.Duration,
        MessagesPerSec: data.MessagesPerSec,
        MaxMessages: data.MaxMessages,
        RandomizeData: data.RandomizeData,
      });
    } catch (error) {
      toast({
        title: "✗ Failed to start",
        description: error instanceof Error ? error.message : "Unknown error",
        variant: "destructive",
      });
    }
  }, [form]);

  const handleStop = useCallback(async () => {
    await StopContinuousSend();
  }, []);

  const handleConnectionToggle = useCallback(async () => {
    const { Address, Port, Protocol, UseTLS, TLSVerify, CACertPath, ClientCertPath, ClientKeyPath } = form.getValues();

    if (isConnected) {
      await Disconnect();
      setIsConnected(false);
      toast({
        title: "✓ Disconnected",
        description: `Closed connection to ${Address}:${Port}`,
        variant: "default",
      });
    } else {
      try {
        const response = await CheckConnection(
          Address, Port, Protocol, UseTLS, TLSVerify,
          CACertPath || "", ClientCertPath || "", ClientKeyPath || ""
        );
        if (response) {
          setIsConnected(true);
          toast({
            title: "✓ Connected",
            description: `Ready to send to ${Address}:${Port}`,
            variant: "default",
          });
        }
      } catch (error) {
        toast({
          title: "✗ Connection Failed",
          description: error instanceof Error ? error.message : "Unable to connect",
          variant: "destructive",
        });
      }
    }
  }, [form, isConnected]);

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
  };

  return (
    <Form {...form}>
      <div className="w-full max-w-3xl mx-auto space-y-3">
        {/* Stats Card - Always visible when running */}
        {(isRunning || stats) && (
          <Card className={isRunning ? "border-blue-500/50 bg-blue-500/5" : ""}>
            <CardHeader className="pb-2 pt-3">
              <CardTitle className="text-sm font-semibold flex items-center gap-2">
                <ActivityIcon className={`w-4 h-4 ${isRunning ? "animate-pulse text-blue-500" : ""}`} />
                Live Statistics
                {isRunning && (
                  <span className="ml-auto flex items-center gap-1 text-xs font-normal text-blue-500">
                    <span className="relative flex h-2 w-2">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
                    </span>
                    Running
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="pb-3">
              <div className="grid grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-500">
                    {(stats?.totalSent ?? 0).toLocaleString()}
                  </div>
                  <div className="text-xs text-muted-foreground">Messages Sent</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-red-500">
                    {(stats?.totalErrors ?? 0).toLocaleString()}
                  </div>
                  <div className="text-xs text-muted-foreground">Errors</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-500">
                    {(stats?.currentRate ?? 0).toFixed(1)}
                  </div>
                  <div className="text-xs text-muted-foreground">Msgs/sec</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {formatDuration(stats?.elapsedSeconds ?? 0)}
                  </div>
                  <div className="text-xs text-muted-foreground">Elapsed</div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Connection Settings */}
        <Card>
          <CardHeader className="pb-2 pt-3">
            <CardTitle className="text-sm font-semibold flex items-center gap-2">
              <ZapIcon className="w-4 h-4" />
              Target Server
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 pb-3">
            <div className="grid grid-cols-12 gap-3">
              <FormField
                control={form.control}
                name="Address"
                render={({ field }) => (
                  <FormItem className="col-span-5 space-y-1">
                    <FormLabel className="text-xs font-medium">IP Address</FormLabel>
                    <FormControl>
                      <Input placeholder="192.168.1.100" className="h-9" disabled={isRunning} {...field} />
                    </FormControl>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="Port"
                render={({ field }) => (
                  <FormItem className="col-span-2 space-y-1">
                    <FormLabel className="text-xs font-medium">Port</FormLabel>
                    <FormControl>
                      <Input placeholder="514" type="number" className="h-9" disabled={isRunning} {...field} />
                    </FormControl>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="Protocol"
                render={({ field }) => (
                  <FormItem className="col-span-4 space-y-1">
                    <FormLabel className="text-xs font-medium">Protocol</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value} disabled={isRunning}>
                      <FormControl>
                        <SelectTrigger className="h-9">
                          <SelectValue placeholder="Select" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="tcp">TCP</SelectItem>
                        <SelectItem value="udp">UDP</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
              <div className="col-span-1 flex items-end pb-[3px]">
                <Button
                  type="button"
                  onClick={handleConnectionToggle}
                  size="icon"
                  disabled={isRunning}
                  className={`h-9 w-9 rounded-full transition-all ${
                    isConnected
                      ? "bg-red-500 hover:bg-red-600 shadow-red-500/25"
                      : "bg-green-500 hover:bg-green-600 shadow-green-500/25"
                  } shadow-lg`}
                >
                  {isConnected ? (
                    <XCircleIcon className="w-4 h-4" />
                  ) : (
                    <CheckCircleIcon className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Rate Configuration */}
        <Card>
          <CardHeader className="pb-2 pt-3">
            <CardTitle className="text-sm font-semibold flex items-center gap-2">
              <GaugeIcon className="w-4 h-4" />
              Rate Configuration
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 pb-3">
            <div className="grid grid-cols-3 gap-3">
              <FormField
                control={form.control}
                name="MessagesPerSec"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">Messages/sec</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={1}
                        max={10000}
                        className="h-9"
                        disabled={isRunning || !isConnected}
                        {...field}
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 1)}
                      />
                    </FormControl>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="Duration"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">Duration (sec)</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={0}
                        max={3600}
                        placeholder="0 = indefinite"
                        className="h-9"
                        disabled={isRunning || !isConnected}
                        {...field}
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      />
                    </FormControl>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="MaxMessages"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">Max Messages</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={0}
                        placeholder="0 = no limit"
                        className="h-9"
                        disabled={isRunning || !isConnected}
                        {...field}
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      />
                    </FormControl>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
            </div>

            {/* Options Row */}
            <div className="flex items-center gap-4 pt-1">
              <FormField
                control={form.control}
                name="RandomizeData"
                render={({ field }) => (
                  <FormItem className="flex items-center gap-2 space-y-0">
                    <FormControl>
                      <input
                        type="checkbox"
                        checked={field.value}
                        onChange={field.onChange}
                        disabled={isRunning || !isConnected}
                        className="h-4 w-4 rounded border-gray-300"
                      />
                    </FormControl>
                    <FormLabel className="text-xs font-medium cursor-pointer">
                      Add sequence number & timestamp
                    </FormLabel>
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Message Template */}
        <Card>
          <CardHeader className="pb-2 pt-3">
            <CardTitle className="text-sm font-semibold flex items-center gap-2">
              <SendIcon className="w-4 h-4" />
              Message Template
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 pb-3">
            <FormField
              control={form.control}
              name="Message"
              render={({ field }) => (
                <FormItem className="space-y-1">
                  <FormControl>
                    <Input
                      placeholder="Enter message template..."
                      className="h-9"
                      disabled={isRunning || !isConnected}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage className="text-xs" />
                </FormItem>
              )}
            />

            {/* Syslog Options */}
            <div className="grid grid-cols-4 gap-3">
              <FormField
                control={form.control}
                name="Facility"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">Facility</FormLabel>
                    <Select
                      onValueChange={(v) => field.onChange(parseInt(v))}
                      defaultValue={String(field.value)}
                      disabled={isRunning || !isConnected}
                    >
                      <FormControl>
                        <SelectTrigger className="h-9">
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="16">local0</SelectItem>
                        <SelectItem value="17">local1</SelectItem>
                        <SelectItem value="18">local2</SelectItem>
                        <SelectItem value="19">local3</SelectItem>
                        <SelectItem value="20">local4</SelectItem>
                        <SelectItem value="21">local5</SelectItem>
                        <SelectItem value="22">local6</SelectItem>
                        <SelectItem value="23">local7</SelectItem>
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="Severity"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">Severity</FormLabel>
                    <Select
                      onValueChange={(v) => field.onChange(parseInt(v))}
                      defaultValue={String(field.value)}
                      disabled={isRunning || !isConnected}
                    >
                      <FormControl>
                        <SelectTrigger className="h-9">
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="6">Info</SelectItem>
                        <SelectItem value="5">Notice</SelectItem>
                        <SelectItem value="4">Warning</SelectItem>
                        <SelectItem value="3">Error</SelectItem>
                        <SelectItem value="2">Critical</SelectItem>
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="MessageFormat"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">Format</FormLabel>
                    <Select
                      value={field.value}
                      onValueChange={(v) => {
                        const next = v as MessageFormat;
                        field.onChange(next);
                        // Keep the legacy flag coherent.
                        form.setValue("UseRFC5424", next === "rfc5424");
                      }}
                      disabled={isRunning || !isConnected}
                    >
                      <FormControl>
                        <SelectTrigger className="h-9">
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {MESSAGE_FORMAT_OPTIONS.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="Appname"
                render={({ field }) => (
                  <FormItem className="space-y-1">
                    <FormLabel className="text-xs font-medium">App Name</FormLabel>
                    <FormControl>
                      <Input
                        className="h-9"
                        disabled={isRunning || !isConnected || isRawPRI}
                        {...field}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            {isRawPRI && (
              <p className="text-[10px] text-muted-foreground">
                Raw + PRI sends only <code>&lt;PRI&gt;</code> plus your message. The
                app-name is supplied by the message itself.
              </p>
            )}
          </CardContent>
        </Card>

        {/* Start/Stop Button */}
        <div className="flex justify-center pt-2">
          <Button
            type="button"
            onClick={isRunning ? handleStop : handleStart}
            disabled={!isConnected && !isRunning}
            size="lg"
            className={`min-w-[200px] h-11 font-semibold transition-all ${
              isRunning
                ? "bg-red-500 hover:bg-red-600 shadow-red-500/25"
                : "bg-blue-500 hover:bg-blue-600 shadow-blue-500/25"
            } shadow-lg disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {isRunning ? (
              <>
                <SquareIcon className="w-5 h-5 mr-2" />
                Stop Test
              </>
            ) : (
              <>
                <PlayIcon className="w-5 h-5 mr-2" />
                Start Stress Test
              </>
            )}
          </Button>
        </div>
      </div>
    </Form>
  );
}
