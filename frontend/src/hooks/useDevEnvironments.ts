import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import type { DevEnvironment } from "@/types/dev-environment";
import { devEnvironmentsApi } from "@/lib/api/environments";

export function useDevEnvironments() {
  const { t } = useTranslation();
  
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");

  const loadDevEnvironments = useCallback(async () => {
    try {
      setLoading(true);
      setError("");
      
      const allEnvironments: DevEnvironment[] = [];
      let currentPage = 1;
      let hasMorePages = true;
      
      while (hasMorePages) {
        const response = await devEnvironmentsApi.list({ 
          page: currentPage, 
          page_size: 100 
        });
        
        if (response.environments && response.environments.length > 0) {
          allEnvironments.push(...response.environments);
          hasMorePages = currentPage < response.total_pages;
          currentPage++;
        } else {
          hasMorePages = false;
        }
      }
      
      setDevEnvironments(allEnvironments);
    } catch (error) {
      console.error("Failed to load dev environments:", error);
      setError(t("tasks.errors.loadDevEnvironmentsFailed"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadDevEnvironments();
  }, [loadDevEnvironments]);

  return {
    devEnvironments,
    loading,
    error,
    loadDevEnvironments,
  };
}