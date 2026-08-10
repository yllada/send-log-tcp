"use client";

import { UseFormReturn } from "react-hook-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FileTextIcon } from "lucide-react";
import { TemplateManager } from "@/components/template-manager";
import {
  FormData,
  MessageFormat,
  MESSAGE_FORMATS,
  MESSAGE_FORMAT_OPTIONS,
} from "./types";
import { FACILITY_OPTIONS, SEVERITY_OPTIONS } from "./constants";

// ================================================================================
// MESSAGE CONFIG CARD COMPONENT
// Single Responsibility: Handles message configuration UI (message format,
// facility, severity, hostname, app-name)
// ================================================================================

interface MessageConfigCardProps {
  form: UseFormReturn<FormData>;
}

// isMessageFormat narrows a persisted string to a supported message format
function isMessageFormat(value: string | undefined): value is MessageFormat {
  return !!value && (MESSAGE_FORMATS as readonly string[]).includes(value);
}

export function MessageConfigCard({ form }: MessageConfigCardProps) {
  const messageFormat = form.watch("MessageFormat");

  // In raw-pri mode the app only prepends <PRI>; the pasted line supplies every
  // header field, so hostname and app-name are never sent.
  const isRawPRI = messageFormat === "raw-pri";

  return (
    <Card>
      <CardHeader className="pb-2 pt-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <FileTextIcon className="w-4 h-4" />
            Message Configuration
          </CardTitle>
          <TemplateManager
            currentValues={{
              Messages: form.watch("Messages"),
              Facility: form.watch("Facility"),
              Severity: form.watch("Severity"),
              Appname: form.watch("Appname"),
              MessageFormat: messageFormat,
              UseRFC5424: form.watch("UseRFC5424"),
            }}
            onLoadTemplate={(template) => {
              form.setValue("Messages", template.message);
              form.setValue("Facility", template.facility);
              form.setValue("Severity", template.severity);
              form.setValue("Appname", template.appname);
              // Templates saved before MessageFormat existed only carry
              // useRfc5424, so fall back to it.
              const loadedFormat: MessageFormat = isMessageFormat(template.messageFormat)
                ? template.messageFormat
                : template.useRfc5424
                  ? "rfc5424"
                  : "rfc3164";
              form.setValue("MessageFormat", loadedFormat);
              form.setValue("UseRFC5424", loadedFormat === "rfc5424");
            }}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-3 pb-3">
        {/* Message Format Selector */}
        <FormField
          control={form.control}
          name="MessageFormat"
          render={({ field }) => (
            <FormItem className="rounded-md border border-border/60 p-3 bg-secondary/20 space-y-1.5">
              <FormLabel className="text-xs font-medium leading-none">
                Message Format
              </FormLabel>
              <Select
                value={field.value}
                onValueChange={(value: string) => {
                  const next = value as MessageFormat;
                  field.onChange(next);
                  // Keep the legacy flag coherent for saved templates/profiles.
                  form.setValue("UseRFC5424", next === "rfc5424");
                }}
              >
                <FormControl>
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select a format" />
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
              <FormDescription className="text-[10px] text-muted-foreground">
                {MESSAGE_FORMAT_OPTIONS.find((option) => option.value === field.value)
                  ?.description}
              </FormDescription>
              <FormMessage className="text-xs" />
            </FormItem>
          )}
        />

        {/* Facility & Severity */}
        <div className="grid grid-cols-2 gap-3">
          <FormField
            control={form.control}
            name="Facility"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">Facility</FormLabel>
                <Select
                  onValueChange={(value: string) => field.onChange(parseInt(value))}
                  defaultValue={field.value?.toString()}
                >
                  <FormControl>
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent className="max-h-[200px]">
                    {FACILITY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value.toString()}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="Severity"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">Severity</FormLabel>
                <Select
                  onValueChange={(value: string) => field.onChange(parseInt(value))}
                  defaultValue={field.value?.toString()}
                >
                  <FormControl>
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent className="max-h-[200px]">
                    {SEVERITY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value.toString()}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />
        </div>

        {/* Hostname & App Name — unused in raw-pri mode */}
        <div className="grid grid-cols-2 gap-3">
          <FormField
            control={form.control}
            name="Hostname"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">Hostname</FormLabel>
                <FormControl>
                  <Input
                    placeholder="Optional"
                    className="h-9"
                    disabled={isRawPRI}
                    {...field}
                  />
                </FormControl>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="Appname"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">App Name</FormLabel>
                <FormControl>
                  <Input
                    placeholder="sendlog"
                    className="h-9"
                    disabled={isRawPRI}
                    {...field}
                  />
                </FormControl>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />
        </div>

        {isRawPRI && (
          <p className="text-[10px] text-muted-foreground">
            Hostname and app-name are supplied by the pasted message. Only facility
            and severity are used, to build the <code>&lt;PRI&gt;</code>.
          </p>
        )}
      </CardContent>
    </Card>
  );
}
