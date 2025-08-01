import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { SUPPORTED_LANGUAGES, STORAGE_KEYS } from "@/lib/constants";

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const handleLanguageChange = (languageCode: string) => {
    i18n.changeLanguage(languageCode);
    localStorage.setItem(STORAGE_KEYS.language, languageCode);
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
