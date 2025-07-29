import React from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { GitCredential, GitCredentialType } from "@/types/git-credentials";
import { GitCredentialType as CredentialTypes } from "@/types/git-credentials";
import {
  Edit,
  Trash2,
  Key,
  Shield,
  User,
  Clock,
  ChevronLeft,
  ChevronRight,
  MoreHorizontal,
  RefreshCw,
} from "lucide-react";

interface GitCredentialListProps {
  credentials: GitCredential[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  typeFilter?: GitCredentialType;
  onPageChange: (page: number) => void;
  onTypeFilterChange: (type: GitCredentialType | undefined) => void;
  onEdit: (credential: GitCredential) => void;
  onDelete: (id: number) => void;
  onRefresh: () => void;
}

export const GitCredentialList: React.FC<GitCredentialListProps> = ({
  credentials,
  loading,
  currentPage,
  totalPages,
  total,
  typeFilter: _typeFilter,
  onPageChange,
  onTypeFilterChange: _onTypeFilterChange,
  onEdit,
  onDelete,
  onRefresh,
}) => {
  const { t } = useTranslation();

  const getTypeIcon = (type: GitCredentialType) => {
    switch (type) {
      case CredentialTypes.PASSWORD:
        return <Key className="w-4 h-4" />;
      case CredentialTypes.TOKEN:
        return <Shield className="w-4 h-4" />;
      case CredentialTypes.SSH_KEY:
        return <User className="w-4 h-4" />;
      default:
        return <Key className="w-4 h-4" />;
    }
  };

  const getTypeName = (type: GitCredentialType) => {
    switch (type) {
      case CredentialTypes.PASSWORD:
        return t("gitCredentials.filter.password");
      case CredentialTypes.TOKEN:
        return t("gitCredentials.filter.token");
      case CredentialTypes.SSH_KEY:
        return t("gitCredentials.filter.sshKey");
      default:
        return t("gitCredentials.unknown", "Unknown");
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">{t("common.loading")}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="text-sm text-foreground">
          {t("common.total")} {total} {t("common.items")}
        </div>
        <div className="flex gap-2">
          <Button
            onClick={onRefresh}
            disabled={loading}
            size="sm"
            variant="ghost"
            className="text-foreground"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            {t("common.refresh")}
          </Button>
        </div>
      </div>

      {credentials.length === 0 ? (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center py-8">
              <Key className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-2">
                {t("gitCredentials.messages.noCredentials")}
              </h3>
              <p className="text-muted-foreground mb-4">
                {t("gitCredentials.messages.noCredentialsDesc")}
              </p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <>
          <Card>
            <CardHeader>
              <CardTitle>{t("gitCredentials.list")}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("gitCredentials.name")}</TableHead>
                      <TableHead>{t("gitCredentials.type")}</TableHead>
                      <TableHead>{t("gitCredentials.username")}</TableHead>
                      <TableHead>{t("gitCredentials.description")}</TableHead>
                      <TableHead>{t("gitCredentials.createdAt")}</TableHead>
                      <TableHead className="text-right">
                        {t("gitCredentials.actions")}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {credentials.map((credential) => (
                      <TableRow key={credential.id}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getTypeIcon(credential.type)}
                            <span className="font-medium">{credential.name}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className="px-2 py-1 bg-muted text-foreground text-xs rounded-full">
                            {getTypeName(credential.type)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className="text-sm">{credential.username}</span>
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-foreground">
                            {credential.description || "-"}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1 text-sm text-foreground">
                            <Clock className="w-3 h-3" />
                            <span>{formatDate(credential.created_at)}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" className="h-8 w-8 p-0">
                                <span className="sr-only">
                                  {t("common.open_menu")}
                                </span>
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuLabel>
                                {t("gitCredentials.actions")}
                              </DropdownMenuLabel>
                              <DropdownMenuItem onClick={() => onEdit(credential)}>
                                <Edit className="mr-2 h-4 w-4" />
                                {t("gitCredentials.edit")}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => onDelete(credential.id)}
                                className="text-destructive"
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                {t("gitCredentials.delete")}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>

          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                {t("common.page")} {currentPage} / {totalPages}
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPageChange(currentPage - 1)}
                  disabled={currentPage <= 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPageChange(currentPage + 1)}
                  disabled={currentPage >= totalPages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default GitCredentialList;
