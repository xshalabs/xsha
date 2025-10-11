import { Card, CardContent } from "@/components/ui/card";
import type { MCPTemplate } from "@/lib/mcp/templateGenerators";

interface MCPTemplateCardProps {
  template: MCPTemplate;
  onClick: () => void;
}

export function MCPTemplateCard({ template, onClick }: MCPTemplateCardProps) {
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
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <h3 className="font-medium text-sm truncate">{template.name}</h3>
          </div>
        </div>

      </CardContent>
    </Card>
  );
}
