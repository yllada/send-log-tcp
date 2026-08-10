"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Form } from "@/components/ui/form";
import { toast } from "@/hooks/use-toast";
import { SendSyslogMessages } from "@/wailsjs/go/main/SyslogService";

import { FormSchema, FormData, defaultFormValues, SyslogPayload } from "./types";
import { useConnection } from "./hooks/useConnection";
import { useCertificates } from "./hooks/useCertificates";
import { ConnectionCard } from "./ConnectionCard";
import { MessageConfigCard } from "./MessageConfigCard";
import { SendMessagesCard } from "./SendMessagesCard";

// ================================================================================
// MAIN FORM COMPONENT
// Single Responsibility: Orchestrates all form sub-components
// Open/Closed: Easy to extend with new cards
// ================================================================================

export function InputForm() {
  // Initialize form with schema validation
  const form = useForm<FormData>({
    resolver: zodResolver(FormSchema),
    defaultValues: defaultFormValues,
  });

  // Custom hooks for separated concerns
  const { isConnected, handleConnectionToggle } = useConnection(form);
  const certificateHandlers = useCertificates(form);

  // Form submission handler
  async function onSubmit(data: FormData) {
    const payload: SyslogPayload = {
      Address: data.Address,
      Port: data.Port,
      Protocol: data.Protocol,
      Messages: data.Messages.split("\n").filter((msg: string) => msg.trim() !== ""),
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
    };

    try {
      const response = await SendSyslogMessages(payload);

      const successCount = response.sentMessages?.length || 0;
      const errorCount = response.errors?.length || 0;

      if (errorCount === 0 && successCount > 0) {
        toast({
          title: "✓ Messages Sent",
          description: `Successfully sent ${successCount} message${successCount > 1 ? "s" : ""}`,
          variant: "default",
        });
      } else if (errorCount > 0 && successCount > 0) {
        toast({
          title: "⚠ Partial Success",
          description: `Sent: ${successCount} | Failed: ${errorCount}`,
          variant: "default",
        });
      } else if (errorCount > 0) {
        toast({
          title: "✗ Send Failed",
          description: response.errors?.[0] || "Failed to send messages",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to send messages",
        variant: "destructive",
      });
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="w-full max-w-3xl mx-auto space-y-3"
      >
        <ConnectionCard
          form={form}
          isConnected={isConnected}
          onConnectionToggle={handleConnectionToggle}
          certificates={certificateHandlers}
        />

        <MessageConfigCard form={form} />

        <SendMessagesCard form={form} isConnected={isConnected} />
      </form>
    </Form>
  );
}

// Re-export for backwards compatibility
export default InputForm;
