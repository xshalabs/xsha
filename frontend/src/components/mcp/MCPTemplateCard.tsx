import { useTranslation } from "react-i18next";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { MCPTemplate } from "@/lib/mcp/templateGenerators";

interface MCPTemplateCardProps {
  template: MCPTemplate;
  onClick: () => void;
}

export function MCPTemplateCard({ template, onClick }: MCPTemplateCardProps) {
  const { t } = useTranslation();

  const handleClick = () => {
    if (template.enabled) {
      onClick();
    }
  };

  return (
    <Card
      className={`
        relative transition-all duration-200 cursor-pointer group border-2 py-3
        ${
          template.enabled
            ? "hover:shadow-sm hover:border-primary/30 hover:bg-primary/5"
            : "opacity-60 cursor-not-allowed bg-muted/30 border-muted"
        }
      `}
      onClick={handleClick}
    >
      <CardContent className="px-3 py-1">
        <div className="flex items-center gap-2">
          {/* Title and Badge */}
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <h3 className="font-medium text-sm truncate">{template.name}</h3>
            {template.comingSoon && (
              <Badge variant="secondary" className="text-xs flex-shrink-0">
                {t("mcp.templates.comingSoon")}
              </Badge>
            )}
          </div>
        </div>

        {/* Coming Soon Overlay */}
        {template.comingSoon && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/90 rounded-lg">
            <div className="text-center">
              <p className="text-xs font-medium text-muted-foreground">
                {t("mcp.templates.comingSoon")}
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
