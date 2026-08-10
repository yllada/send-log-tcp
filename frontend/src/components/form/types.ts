import { z } from "zod";

// ================================================================================
// FORM TYPES & VALIDATION SCHEMA
// Single Responsibility: Defines all form-related types and validation
// ================================================================================

// IP validation regex (IPv4, IPv6, and localhost)
export const ipRegex = /^(?:(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}|(?:[a-fA-F0-9]{1,4}:){7}[a-fA-F0-9]{1,4}|(?:[a-fA-F0-9]{1,4}:){1,7}:|(?:[a-fA-F0-9]{1,4}:){1,6}:[a-fA-F0-9]{1,4}|::(?:[a-fA-F0-9]{1,4}:){0,5}[a-fA-F0-9]{1,4}|[a-fA-F0-9]{1,4}::(?:[a-fA-F0-9]{1,4}:){0,4}[a-fA-F0-9]{1,4}|localhost)$/;

// Supported syslog message formats.
// "raw-pri" emits only "<PRI>" followed by the message bytes verbatim, so the
// pasted line must already carry its own header.
export const MESSAGE_FORMATS = ["rfc5424", "rfc3164", "raw-pri"] as const;

export type MessageFormat = (typeof MESSAGE_FORMATS)[number];

// Message format options for the format selector
export const MESSAGE_FORMAT_OPTIONS: ReadonlyArray<{
  value: MessageFormat;
  label: string;
  description: string;
}> = [
  {
    value: "rfc5424",
    label: "RFC 5424",
    description: "Modern syslog header (version, timestamp, hostname, app-name)",
  },
  {
    value: "rfc3164",
    label: "RFC 3164",
    description: "Legacy BSD syslog header",
  },
  {
    value: "raw-pri",
    label: "Raw + PRI",
    description: "Only <PRI> plus your message, byte for byte. No header injected.",
  },
];

// Form validation schema using Zod
export const FormSchema = z.object({
  // Connection settings
  Address: z.string().regex(ipRegex, { message: "Invalid IP address (IPv4 or IPv6)" }),
  Port: z.string({ message: "Port is required" })
    .refine((val: string) => {
      const num = parseInt(val, 10);
      return !isNaN(num) && num >= 1 && num <= 65535;
    }, { message: "Port must be between 1-65535" }),
  Protocol: z.string({ message: "Please select a protocol" }),
  FramingMethod: z.enum(["octet-counting", "non-transparent"], {
    message: "Please select a framing method",
  }),
  
  // TLS settings
  UseTLS: z.boolean(),
  TLSVerify: z.boolean(),
  CACertPath: z.string().optional(),
  ClientCertPath: z.string().optional(),
  ClientKeyPath: z.string().optional(),
  
  // Message settings
  Messages: z.string().min(1, { message: "At least one message is required" }),
  Facility: z.number().min(0).max(23, "Facility must be between 0-23"),
  Severity: z.number().min(0).max(7, "Severity must be between 0-7"),
  Hostname: z.string().optional(),
  Appname: z.string().min(1, "Application name is required"),
  MessageFormat: z.enum(MESSAGE_FORMATS, {
    message: "Please select a message format",
  }),
  // Legacy format switch, kept in sync with MessageFormat so saved templates
  // and profiles stay coherent.
  UseRFC5424: z.boolean(),
});

// Infer the form data type from the schema
export type FormData = z.infer<typeof FormSchema>;

// Default form values
export const defaultFormValues: FormData = {
  Address: "",
  Port: "",
  Protocol: "",
  Messages: "",
  FramingMethod: "octet-counting",
  Facility: 16, // local0
  Severity: 6, // info
  Hostname: "",
  Appname: "sendlog",
  MessageFormat: "rfc5424",
  UseRFC5424: true,
  UseTLS: false,
  TLSVerify: false,
  CACertPath: "",
  ClientCertPath: "",
  ClientKeyPath: "",
};

// Connection profile values type (subset for ProfileManager)
export interface ConnectionValues {
  Address: string;
  Port: string;
  Protocol: string;
  FramingMethod: string;
  UseTLS: boolean;
  TLSVerify: boolean;
  CACertPath: string;
  ClientCertPath: string;
  ClientKeyPath: string;
}

// Message config values type (subset for TemplateManager)
export interface MessageConfigValues {
  Messages: string;
  Facility: number;
  Severity: number;
  Appname: string;
  MessageFormat: MessageFormat;
  UseRFC5424: boolean;
}

// Syslog request payload type
export interface SyslogPayload {
  Address: string;
  Port: string;
  Protocol: string;
  Messages: string[];
  FramingMethod: string;
  Facility: number;
  Severity: number;
  Hostname: string;
  Appname: string;
  MessageFormat: MessageFormat;
  UseRFC5424: boolean;
  UseTLS: boolean;
  TLSVerify: boolean;
  CACertPath: string;
  ClientCertPath: string;
  ClientKeyPath: string;
}
