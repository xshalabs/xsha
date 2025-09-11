import { useState, useCallback } from "react";
import { useTranslation } from "react-i18next";
import type { Project } from "@/types/project";
import { projectsApi } from "@/lib/api/projects";

interface UseProjectBranchesOptions {
  currentProject?: Project;
}

export function useProjectBranches({ currentProject }: UseProjectBranchesOptions) {
  const { t } = useTranslation();
  
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  const [fetching, setFetching] = useState(false);
  const [error, setError] = useState<string>("");

  const fetchProjectBranches = useCallback(async () => {
    if (!currentProject) return;

    try {
      setFetching(true);
      setError("");
      setAvailableBranches([]);

      const response = await projectsApi.fetchBranches(currentProject.id);

      if (response.result.can_access && response.result.branches && response.result.branches.length > 0) {
        setAvailableBranches(response.result.branches);
      } else {
        const errorMsg = response.result.error_message || 
          (response.result.can_access ? t("tasks.errors.noBranchesFound") : t("tasks.errors.fetchBranchesFailed"));
        setError(errorMsg);
      }
    } catch (error) {
      console.error("Failed to fetch branches:", error);
      setError(t("tasks.errors.fetchBranchesFailed"));
    } finally {
      setFetching(false);
    }
  }, [currentProject, t]);

  return {
    availableBranches,
    fetching,
    error,
    fetchProjectBranches,
  };
}