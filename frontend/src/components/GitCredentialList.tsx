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
import type { GitCredential, GitCredentialType } from "@/types/git-credentials";
import { GitCredentialType as CredentialTypes } from "@/types/git-credentials";
import {
  Edit,
  Trash2,
  Eye,
  EyeOff,
  Key,
  Shield,
  User,
  Clock,
  ChevronLeft,
  ChevronRight,
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
  onRefresh: _onRefresh,
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

  if (credentials.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-8">
            <Key className="w-12 h-12 mx-auto text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {t("gitCredentials.messages.noCredentials")}
            </h3>
            <p className="text-gray-600 mb-4">
              {t("gitCredentials.messages.noCredentialsDesc")}
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="text-sm text-foreground">
        {t("gitCredentials.pagination.total")} {total}{" "}
        {t("gitCredentials.pagination.items")}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">
            {t("gitCredentials.filter.title")}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("gitCredentials.name")}</TableHead>
                  <TableHead>{t("gitCredentials.type")}</TableHead>
                  <TableHead>{t("gitCredentials.status")}</TableHead>
                  <TableHead>{t("gitCredentials.username")}</TableHead>
                  <TableHead>{t("gitCredentials.description")}</TableHead>
                  <TableHead>{t("gitCredentials.createdAt")}</TableHead>
                  <TableHead>{t("gitCredentials.lastUsed")}</TableHead>
                  <TableHead className="text-right">
                    {t("gitCredentials.actions")}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {credentials.map((credential) => (
                  <TableRow
                    key={credential.id}
                    className={credential.is_active ? "" : "opacity-60"}
                  >
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getTypeIcon(credential.type)}
                        <span className="font-medium">{credential.name}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                        {getTypeName(credential.type)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {credential.is_active ? (
                          <Eye className="w-4 h-4 text-green-500" />
                        ) : (
                          <EyeOff className="w-4 h-4 text-gray-400" />
                        )}
                        <span
                          className={`text-xs ${
                            credential.is_active
                              ? "text-green-600"
                              : "text-gray-500"
                          }`}
                        >
                          {credential.is_active
                            ? t("gitCredentials.active")
                            : t("gitCredentials.inactive")}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm">{credential.username}</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-gray-600">
                        {credential.description || "-"}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1 text-sm text-gray-500">
                        <Clock className="w-3 h-3" />
                        <span>{formatDate(credential.created_at)}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {credential.last_used ? (
                        <div className="flex items-center gap-1 text-sm text-gray-500">
                          <Clock className="w-3 h-3" />
                          <span>{formatDate(credential.last_used)}</span>
                        </div>
                      ) : (
                        <span className="text-sm text-gray-400">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onEdit(credential)}
                          title={t("gitCredentials.edit")}
                        >
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onDelete(credential.id)}
                          className="text-red-600 hover:text-red-700 hover:border-red-300"
                          title={t("gitCredentials.delete")}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-600">
            {t("gitCredentials.pagination.page")} {currentPage} / {totalPages}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              className="text-foreground"
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-foreground"
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default GitCredentialList;
