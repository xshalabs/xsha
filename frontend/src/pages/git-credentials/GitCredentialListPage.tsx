import React, { useState, useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { usePageTitle } from "@/hooks/usePageTitle";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { usePageActions } from "@/contexts/PageActionsContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Section,
  SectionGroup,
  SectionHeader,
  SectionTitle,
  SectionDescription,
} from "@/components/content/section";
import {
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
  MetricCardButton,
} from "@/components/metric/metric-card";
import { GitCredentialList } from "@/components/GitCredentialList";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import type {
  GitCredential,
  GitCredentialListParams,
} from "@/types/git-credentials";
import { GitCredentialType } from "@/types/git-credentials";
import { Plus, Key, Shield, User, ListFilter } from "lucide-react";

const GitCredentialListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();

  const [credentials, setCredentials] = useState<GitCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [typeFilter, setTypeFilter] = useState<GitCredentialType | undefined>();
  const [error, setError] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [credentialToDelete, setCredentialToDelete] = useState<number | null>(null);

  const pageSize = 10;

  usePageTitle(t("common.pageTitle.gitCredentials"));

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      navigate("/git-credentials/create");
    };

    setActions(
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t("gitCredentials.create")}
      </Button>
    );

    // Clear breadcrumb items (we're at root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [navigate, setActions, setItems, t]);

  const loadCredentials = async (params?: GitCredentialListParams) => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiService.gitCredentials.list({
        page: currentPage,
        page_size: pageSize,
        type: typeFilter,
        ...params,
      });

      setCredentials(response.credentials);
      setTotal(response.total);
      setTotalPages(response.total_pages);
    } catch (err: any) {
      setError(err.message || t("gitCredentials.messages.loadFailed"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCredentials();
  }, [currentPage, typeFilter]);

  const handleRefresh = () => {
    setCurrentPage(1);
    loadCredentials({ page: 1 });
  };

  const handleEdit = (credential: GitCredential) => {
    navigate(`/git-credentials/${credential.id}/edit`);
  };

  const handleDelete = (id: number) => {
    setCredentialToDelete(id);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (!credentialToDelete) return;

    try {
      await apiService.gitCredentials.delete(credentialToDelete);
      toast.success(t("gitCredentials.messages.deleteSuccess"));
      await loadCredentials();
    } catch (err: any) {
      const errorMessage = err.message || t("gitCredentials.messages.deleteFailed");
      toast.error(errorMessage);
    } finally {
      setDeleteDialogOpen(false);
      setCredentialToDelete(null);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setCredentialToDelete(null);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleTypeFilterChange = (type: GitCredentialType | undefined) => {
    setTypeFilter(type);
    setCurrentPage(1);
  };



  const handleBatchDelete = async (ids: number[]) => {
    try {
      await Promise.all(ids.map(id => apiService.gitCredentials.delete(id)));
      toast.success(t("gitCredentials.messages.batchDeleteSuccess", `Successfully deleted ${ids.length} credentials`));
      await loadCredentials();
    } catch (err: any) {
      const errorMessage = err.message || t("gitCredentials.messages.batchDeleteFailed", "Failed to delete credentials");
      toast.error(errorMessage);
    }
  };

  // Calculate statistics
  const statistics = useMemo(() => {
    const passwordCount = credentials.filter(
      (cred) => cred.type === GitCredentialType.PASSWORD
    ).length;
    const tokenCount = credentials.filter(
      (cred) => cred.type === GitCredentialType.TOKEN
    ).length;
    const sshKeyCount = credentials.filter(
      (cred) => cred.type === GitCredentialType.SSH_KEY
    ).length;

    return [
      {
        title: t("gitCredentials.filter.password"),
        value: passwordCount,
        variant: "success" as const,
        type: GitCredentialType.PASSWORD,
        icon: Key,
      },
      {
        title: t("gitCredentials.filter.token"),
        value: tokenCount,
        variant: "warning" as const,
        type: GitCredentialType.TOKEN,
        icon: Shield,
      },
      {
        title: t("gitCredentials.filter.sshKey"),
        value: sshKeyCount,
        variant: "default" as const,
        type: GitCredentialType.SSH_KEY,
        icon: User,
      },
      {
        title: t("common.total"),
        value: total,
        variant: "ghost" as const,
        type: undefined,
        icon: ListFilter,
      },
    ];
  }, [credentials, total, t]);

  const handleStatisticClick = (statisticType: GitCredentialType | undefined) => {
    if (statisticType === undefined) {
      // Clear all filters
      setTypeFilter(undefined);
    } else {
      // Toggle filter
      if (typeFilter === statisticType) {
        setTypeFilter(undefined);
      } else {
        setTypeFilter(statisticType);
      }
    }
    setCurrentPage(1);
  };

  return (
    <>
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>{t("gitCredentials.title")}</SectionTitle>
            <SectionDescription>
              {t(
                "gitCredentials.subtitle",
                "Manage your Git repository access credentials"
              )}
            </SectionDescription>
          </SectionHeader>

          {error && (
            <Card className="border-destructive/20 bg-destructive/10">
              <CardContent className="pt-6">
                <p className="text-destructive">{error}</p>
              </CardContent>
            </Card>
          )}

          <MetricCardGroup>
            {statistics.map((stat) => {
              const isActive = typeFilter === stat.type || (stat.type === undefined && typeFilter === undefined);
              const Icon = stat.icon;

              return (
                <MetricCardButton
                  key={stat.title}
                  variant={stat.variant}
                  onClick={() => handleStatisticClick(stat.type)}
                  className={`transition-all duration-200 hover:scale-[1.02] ${
                    isActive 
                      ? "ring-2 ring-ring ring-offset-2 shadow-md" 
                      : "hover:shadow-sm"
                  }`}
                >
                  <MetricCardHeader className="flex justify-between items-center gap-2 w-full">
                    <MetricCardTitle className="truncate font-medium">
                      {stat.title}
                    </MetricCardTitle>
                    <Icon className={`size-4 transition-colors ${
                      isActive ? "text-current" : "text-muted-foreground"
                    }`} />
                  </MetricCardHeader>
                  <MetricCardValue className="text-xl font-semibold">
                    {stat.value}
                  </MetricCardValue>
                </MetricCardButton>
              );
            })}
          </MetricCardGroup>
        </Section>

        <Section>
          <GitCredentialList
            credentials={credentials}
            loading={loading}
            currentPage={currentPage}
            totalPages={totalPages}
            total={total}
            typeFilter={typeFilter}
            onPageChange={handlePageChange}
            onTypeFilterChange={handleTypeFilterChange}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onBatchDelete={handleBatchDelete}
            onRefresh={handleRefresh}
          />
        </Section>
      </SectionGroup>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-foreground">
              {t("gitCredentials.messages.delete_confirm_title")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("gitCredentials.messages.deleteConfirm")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleCancelDelete}
            >
              {t("common.cancel")}
            </Button>
            <Button
              variant="destructive"
              className="text-foreground"
              onClick={handleConfirmDelete}
            >
              {t("common.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default GitCredentialListPage;
