import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { usePageTitle } from "@/hooks/usePageTitle";
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
import { GitCredentialList } from "@/components/GitCredentialList";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import type {
  GitCredential,
  GitCredentialListParams,
} from "@/types/git-credentials";
import { GitCredentialType } from "@/types/git-credentials";
import { Plus } from "lucide-react";

const GitCredentialListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

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

  const handleCreateNew = () => {
    navigate("/git-credentials/create");
  };

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("gitCredentials.title")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t(
                "gitCredentials.subtitle",
                "Manage your Git repository access credentials"
              )}
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={handleCreateNew}>
              <Plus className="h-4 w-4 mr-2" />
              {t("gitCredentials.create")}
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {error && (
          <Card className="mb-6 border-red-200 bg-red-50">
            <CardContent className="pt-6">
              <p className="text-red-600">{error}</p>
            </CardContent>
          </Card>
        )}

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
          onRefresh={handleRefresh}
        />
      </div>

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
    </div>
  );
};

export default GitCredentialListPage;
