import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { SUPPORTED_LANGUAGES } from "@/lib/constants";
import { apiService } from "@/lib/api/index";
import { useAuth } from "@/contexts/AuthContext";

export function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const { isAuthenticated } = useAuth();

  const handleLanguageChange = async (languageCode: string) => {
    i18n.changeLanguage(languageCode);

    if (isAuthenticated) {
      try {
        await apiService.setLanguagePreference(languageCode);
      } catch (error) {
        console.warn("Failed to sync language preference with backend:", error);
      }
    }
  };

  return (
    <div className="flex items-center space-x-2">
      {SUPPORTED_LANGUAGES.map((language) => (
        <Button
          key={language.code}
          variant={i18n.language === language.code ? "default" : "outline"}
          size="sm"
          onClick={() => handleLanguageChange(language.code)}
          className="text-xs"
        >
          <span className="mr-1">{language.flag}</span>
          {language.name}
        </Button>
      ))}
    </div>
  );
}
