"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "@/hooks/use-toast";
import {
  GetTemplates,
  SaveTemplate,
  DeleteTemplate,
} from "@/wailsjs/go/main/TemplateService";
import {
  FileTextIcon,
  SaveIcon,
  Trash2Icon,
  PlusIcon,
  BookmarkIcon,
} from "lucide-react";

// Template type matching Go struct
interface LogTemplate {
  id: string;
  name: string;
  description?: string;
  message: string;
  facility: number;
  severity: number;
  appname: string;
  // Optional: templates saved before the message format selector existed only
  // carry useRfc5424.
  messageFormat?: string;
  useRfc5424: boolean;
  createdAt: number;
  updatedAt: number;
}

interface TemplateManagerProps {
  // Current form values to save
  currentValues: {
    Messages: string;
    Facility: number;
    Severity: number;
    Appname: string;
    MessageFormat: string;
    UseRFC5424: boolean;
  };
  // Callback to load template into form
  onLoadTemplate: (template: LogTemplate) => void;
}

// Facility labels
const facilityLabels: Record<number, string> = {
  0: "kernel", 1: "user", 2: "mail", 3: "daemon", 4: "auth",
  5: "syslog", 6: "lpr", 7: "news", 8: "uucp", 9: "clock",
  10: "authpriv", 11: "ftp", 12: "ntp", 13: "audit", 14: "alert",
  15: "clock2", 16: "local0", 17: "local1", 18: "local2", 19: "local3",
  20: "local4", 21: "local5", 22: "local6", 23: "local7",
};

// Severity labels
const severityLabels: Record<number, string> = {
  0: "Emergency", 1: "Alert", 2: "Critical", 3: "Error",
  4: "Warning", 5: "Notice", 6: "Info", 7: "Debug",
};

// Message format labels
const messageFormatLabels: Record<string, string> = {
  "rfc5424": "RFC 5424",
  "rfc3164": "RFC 3164",
  "raw-pri": "Raw + PRI",
};

// formatLabel resolves the display label for a stored template. Templates saved
// before messageFormat existed fall back to the legacy useRfc5424 flag.
function formatLabel(messageFormat: string | undefined, useRfc5424: boolean): string {
  if (messageFormat && messageFormatLabels[messageFormat]) {
    return messageFormatLabels[messageFormat];
  }
  return useRfc5424 ? "RFC 5424" : "RFC 3164";
}

export function TemplateManager({ currentValues, onLoadTemplate }: TemplateManagerProps) {
  const [templates, setTemplates] = useState<LogTemplate[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isSaveOpen, setIsSaveOpen] = useState(false);
  const [templateName, setTemplateName] = useState("");
  const [templateDescription, setTemplateDescription] = useState("");
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>("");

  // Load templates on mount
  const loadTemplates = useCallback(async () => {
    try {
      const data = await GetTemplates();
      setTemplates(data || []);
    } catch (error) {
      console.error("Failed to load templates:", error);
    }
  }, []);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  // Save current message config as template
  const handleSave = async () => {
    if (!templateName.trim()) {
      toast({
        title: "✗ Error",
        description: "Template name is required",
        variant: "destructive",
      });
      return;
    }

    try {
      const template: Partial<LogTemplate> = {
        name: templateName.trim(),
        description: templateDescription.trim(),
        message: currentValues.Messages,
        facility: currentValues.Facility,
        severity: currentValues.Severity,
        appname: currentValues.Appname,
        messageFormat: currentValues.MessageFormat,
        useRfc5424: currentValues.UseRFC5424,
      };

      await SaveTemplate(template as LogTemplate);
      await loadTemplates();

      toast({
        title: "✓ Template Saved",
        description: `"${templateName}" saved successfully`,
        variant: "default",
      });

      setTemplateName("");
      setTemplateDescription("");
      setIsSaveOpen(false);
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to save template",
        variant: "destructive",
      });
    }
  };

  // Load selected template
  const handleLoad = () => {
    const template = templates.find((t) => t.id === selectedTemplateId);
    if (template) {
      onLoadTemplate(template);
      toast({
        title: "✓ Template Loaded",
        description: `"${template.name}" loaded`,
        variant: "default",
      });
      setIsOpen(false);
    }
  };

  // Delete template
  const handleDelete = async () => {
    if (!selectedTemplateId) return;

    const template = templates.find((t) => t.id === selectedTemplateId);
    if (!template) return;

    try {
      await DeleteTemplate(selectedTemplateId);
      await loadTemplates();
      setSelectedTemplateId("");

      toast({
        title: "✓ Template Deleted",
        description: `"${template.name}" deleted`,
        variant: "default",
      });
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to delete template",
        variant: "destructive",
      });
    }
  };

  return (
    <div className="flex items-center gap-1">
      {/* Load Template Button */}
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-9 w-9"
            title="Load Template"
          >
            <BookmarkIcon className="w-4 h-4" />
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <FileTextIcon className="w-5 h-5" />
              Log Templates
            </DialogTitle>
            <DialogDescription>
              Load a saved message template
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {templates.length === 0 ? (
              <div className="text-center text-muted-foreground py-8">
                No saved templates
              </div>
            ) : (
              <div className="space-y-2">
                <Label>Select Template</Label>
                <Select
                  value={selectedTemplateId}
                  onValueChange={setSelectedTemplateId}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Choose a template..." />
                  </SelectTrigger>
                  <SelectContent>
                    {templates.map((template) => (
                      <SelectItem key={template.id} value={template.id}>
                        <div className="flex flex-col">
                          <span>{template.name}</span>
                          <span className="text-xs text-muted-foreground">
                            {facilityLabels[template.facility]} / {severityLabels[template.severity]}
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {selectedTemplateId && (
                  <div className="mt-4 p-3 bg-muted rounded-lg text-sm space-y-2">
                    {(() => {
                      const t = templates.find((t) => t.id === selectedTemplateId);
                      if (!t) return null;
                      return (
                        <>
                          <div><strong>App:</strong> {t.appname}</div>
                          <div><strong>Facility:</strong> {facilityLabels[t.facility]} ({t.facility})</div>
                          <div><strong>Severity:</strong> {severityLabels[t.severity]} ({t.severity})</div>
                          <div><strong>Format:</strong> {formatLabel(t.messageFormat, t.useRfc5424)}</div>
                          {t.description && (
                            <div><strong>Description:</strong> {t.description}</div>
                          )}
                          <div className="mt-2">
                            <strong>Message Preview:</strong>
                            <pre className="mt-1 p-2 bg-background rounded text-xs whitespace-pre-wrap break-all max-h-32 overflow-auto">
                              {t.message.slice(0, 200)}{t.message.length > 200 ? "..." : ""}
                            </pre>
                          </div>
                        </>
                      );
                    })()}
                  </div>
                )}
              </div>
            )}
          </div>

          <DialogFooter className="gap-2 sm:gap-0">
            {selectedTemplateId && (
              <Button
                type="button"
                variant="destructive"
                size="sm"
                onClick={handleDelete}
              >
                <Trash2Icon className="w-4 h-4 mr-1" />
                Delete
              </Button>
            )}
            <Button
              type="button"
              onClick={handleLoad}
              disabled={!selectedTemplateId}
            >
              <FileTextIcon className="w-4 h-4 mr-1" />
              Load
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Save Template Button */}
      <Dialog open={isSaveOpen} onOpenChange={setIsSaveOpen}>
        <DialogTrigger asChild>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-9 w-9"
            title="Save Template"
          >
            <PlusIcon className="w-4 h-4" />
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <SaveIcon className="w-5 h-5" />
              Save Message Template
            </DialogTitle>
            <DialogDescription>
              Save current message settings for later use
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="templateName">Template Name *</Label>
              <Input
                id="templateName"
                value={templateName}
                onChange={(e) => setTemplateName(e.target.value)}
                placeholder="e.g., Auth Error Log"
                className="h-9"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="templateDesc">Description (optional)</Label>
              <Input
                id="templateDesc"
                value={templateDescription}
                onChange={(e) => setTemplateDescription(e.target.value)}
                placeholder="e.g., Template for authentication failures"
                className="h-9"
              />
            </div>

            <div className="p-3 bg-muted rounded-lg text-sm space-y-1">
              <div className="font-medium mb-2">Message Details:</div>
              <div><strong>App:</strong> {currentValues.Appname || "(not set)"}</div>
              <div><strong>Facility:</strong> {facilityLabels[currentValues.Facility]} ({currentValues.Facility})</div>
              <div><strong>Severity:</strong> {severityLabels[currentValues.Severity]} ({currentValues.Severity})</div>
              <div><strong>Format:</strong> {formatLabel(currentValues.MessageFormat, currentValues.UseRFC5424)}</div>
              <div className="mt-2">
                <strong>Message Preview:</strong>
                <pre className="mt-1 p-2 bg-background rounded text-xs whitespace-pre-wrap break-all max-h-24 overflow-auto">
                  {currentValues.Messages?.slice(0, 150) || "(empty)"}
                  {(currentValues.Messages?.length || 0) > 150 ? "..." : ""}
                </pre>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" onClick={handleSave}>
              <SaveIcon className="w-4 h-4 mr-1" />
              Save Template
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
