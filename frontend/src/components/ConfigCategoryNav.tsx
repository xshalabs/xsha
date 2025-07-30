import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { 
  Settings, 
  Code, 
  Database, 
  Users, 
  Shield,
  Globe,
  Cpu
} from "lucide-react";

interface ConfigCategory {
  key: string;
  name: string;
  description: string;
  icon: React.ElementType;
  count: number;
}

interface ConfigCategoryNavProps {
  categories: ConfigCategory[];
  activeCategory: string;
  onCategoryChange: (category: string) => void;
}

export function ConfigCategoryNav({
  categories,
  activeCategory,
  onCategoryChange,
}: ConfigCategoryNavProps) {
  const { t } = useTranslation();

  const getCategoryIcon = (key: string) => {
    switch (key) {
      case 'general': return Settings;
      case 'development': return Code;
      case 'database': return Database;
      case 'auth': return Shield;
      case 'system': return Cpu;
      case 'ui': return Globe;
      case 'user': return Users;
      default: return Settings;
    }
  };

  return (
    <Card className="w-64 h-fit p-4">
      <div className="mb-4">
        <h3 className="font-semibold text-lg">{t("system-config.categories")}</h3>
        <p className="text-sm text-muted-foreground">
          {t("system-config.categories_description")}
        </p>
      </div>
      
      <div className="space-y-2">
        {categories.map((category) => {
          const Icon = category.icon || getCategoryIcon(category.key);
          const isActive = activeCategory === category.key;
          
          return (
            <Button
              key={category.key}
              variant={isActive ? "default" : "ghost"}
              className={`w-full justify-start h-auto p-3 ${
                isActive ? "bg-primary text-primary-foreground" : ""
              }`}
              onClick={() => onCategoryChange(category.key)}
            >
              <div className="flex items-center w-full">
                <Icon className="w-4 h-4 mr-3 flex-shrink-0" />
                <div className="flex-1 text-left">
                  <div className="font-medium">{category.name}</div>
                  <div className="text-xs opacity-75 mt-1">
                    {category.description}
                  </div>
                </div>
                <Badge variant="secondary" className="ml-2">
                  {category.count}
                </Badge>
              </div>
            </Button>
          );
        })}
      </div>
    </Card>
  );
} 