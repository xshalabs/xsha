import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { ChevronDown, ChevronUp, Info } from "lucide-react";
import { MCPTemplateCard } from "./MCPTemplateCard";
import { Context7FormSheet } from "./Context7FormSheet";
import { DeepwikiFormSheet } from "./DeepwikiFormSheet";
import {
  Section,
  SectionHeader,
  SectionTitle,
  SectionDescription,
} from "@/components/content/section";
import { Button } from "@/components/ui/button";
import {
  getAllTemplates,
  type MCPTemplate,
} from "@/lib/mcp/templateGenerators";

interface MCPTemplatesProps {
  onMCPCreated: () => void;
}

export function MCPTemplates({ onMCPCreated }: MCPTemplatesProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);

  const templates = getAllTemplates();

  const handleTemplateClick = (template: MCPTemplate) => {
    if (!template.enabled) return;

    setSelectedTemplate(template.id);
  };

  const handleTemplateFormClose = () => {
    setSelectedTemplate(null);
  };

  const handleTemplateSuccess = async () => {
    setSelectedTemplate(null);
    await onMCPCreated();
  };

  return (
    <>
      <Section className="mt-8 border-t bg-muted/30 rounded-lg p-4">
        <SectionHeader>
          <div className="flex items-center justify-between w-full">
            <div className="space-y-1">
              <SectionTitle className="flex items-center gap-2 text-base">
                {t("mcp.templates.title")}
                <Info className="h-4 w-4 text-muted-foreground" />
              </SectionTitle>
              <SectionDescription className="text-sm">
                {t("mcp.templates.description")}{" "}
              </SectionDescription>
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="ml-4"
              onClick={() => setIsOpen(!isOpen)}
            >
              {isOpen ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </Button>
          </div>
        </SectionHeader>

        {isOpen && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 mt-4">
            {templates.map((template) => (
              <MCPTemplateCard
                key={template.id}
                template={template}
                onClick={() => handleTemplateClick(template)}
              />
            ))}
          </div>
        )}
      </Section>

      {/* Template-specific forms */}
      <Context7FormSheet
        isOpen={selectedTemplate === "context7"}
        onClose={handleTemplateFormClose}
        onSuccess={handleTemplateSuccess}
      />

      <DeepwikiFormSheet
        isOpen={selectedTemplate === "deepwiki"}
        onClose={handleTemplateFormClose}
        onSuccess={handleTemplateSuccess}
      />

      {/* Future template forms can be added here */}
      {/*
      <SlackFormSheet
        isOpen={selectedTemplate === 'slack'}
        onClose={handleTemplateFormClose}
        onSuccess={handleTemplateSuccess}
      />
      */}
    </>
  );
}
