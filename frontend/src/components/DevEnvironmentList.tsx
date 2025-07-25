import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import {
  MoreHorizontal,
  Edit,
  Trash2,
  Monitor,
  Filter,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import type {
  DevEnvironmentDisplay,
  DevEnvironmentListParams,
  DevEnvironmentType,
} from '@/types/dev-environment';

interface DevEnvironmentListProps {
  environments: DevEnvironmentDisplay[];
  loading: boolean;
  params: DevEnvironmentListParams;
  totalPages: number;
  onPageChange: (page: number) => void;
  onFiltersChange: (filters: Partial<DevEnvironmentListParams>) => void;
  onEdit: (environment: DevEnvironmentDisplay) => void;
  onDelete: (id: number) => void;
}

// 环境类型配置
const typeConfig: Record<DevEnvironmentType, { label: string; color: string }> = {
  'claude_code': { label: 'Claude Code', color: 'text-blue-600' },
  'gemini_cli': { label: 'Gemini CLI', color: 'text-purple-600' },
  'opencode': { label: 'OpenCode', color: 'text-green-600' },
};

const DevEnvironmentList: React.FC<DevEnvironmentListProps> = ({
  environments,
  loading,
  params,
  totalPages,
  onPageChange,
  onFiltersChange,
  onEdit,
  onDelete,
}) => {
  const { t } = useTranslation();
  const [showFilters, setShowFilters] = useState(false);

  // 格式化内存大小
  const formatMemory = (mb: number) => {
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(1)} GB`;
    }
    return `${mb} MB`;
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>{t('dev_environments.list')}</span>
            <Button variant="outline" onClick={() => setShowFilters(!showFilters)}>
              <Filter className="h-4 w-4 mr-2" />
              {t('common.filter')}
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="flex items-center space-x-4">
                <Skeleton className="h-12 w-12 rounded-full" />
                <div className="space-y-2">
                  <Skeleton className="h-4 w-[250px]" />
                  <Skeleton className="h-4 w-[200px]" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>{t('dev_environments.list')}</span>
          <Button variant="outline" onClick={() => setShowFilters(!showFilters)}>
            <Filter className="h-4 w-4 mr-2" />
            {t('common.filter')}
          </Button>
        </CardTitle>

        {/* 筛选器 */}
        {showFilters && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t">
            <div className="space-y-2">
              <Label>{t('dev_environments.filter.type')}</Label>
              <Select
                value={params.type || 'all'}
                onValueChange={(value) => onFiltersChange({ type: value === 'all' ? undefined : value as DevEnvironmentType })}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('common.all')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('common.all')}</SelectItem>
                  <SelectItem value="claude_code">Claude Code</SelectItem>
                  <SelectItem value="gemini_cli">Gemini CLI</SelectItem>
                  <SelectItem value="opencode">OpenCode</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>{t('common.page_size')}</Label>
              <Select
                value={params.page_size?.toString() || '20'}
                onValueChange={(value) => onFiltersChange({ page_size: parseInt(value) })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="20">20</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        )}
      </CardHeader>

      <CardContent>
        {environments.length === 0 ? (
          <div className="text-center py-8">
            <Monitor className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">{t('dev_environments.empty.title')}</h3>
            <p className="text-muted-foreground">{t('dev_environments.empty.description')}</p>
          </div>
        ) : (
          <div className="space-y-4">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('dev_environments.table.name')}</TableHead>
                  <TableHead>{t('dev_environments.table.type')}</TableHead>
                  <TableHead>{t('dev_environments.table.resources')}</TableHead>
                  <TableHead className="text-right">{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {environments.map((environment) => (
                  <TableRow key={environment.id}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{environment.name}</div>
                        {environment.description && (
                          <div className="text-sm text-muted-foreground">
                            {environment.description}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className={typeConfig[environment.type].color}>
                        {typeConfig[environment.type].label}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        <div>CPU: {environment.cpu_limit} {t('dev_environments.stats.cores')}</div>
                        <div>{t('dev_environments.stats.memory')}: {formatMemory(environment.memory_limit)}</div>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" className="h-8 w-8 p-0">
                            <span className="sr-only">{t('common.open_menu')}</span>
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>{t('common.actions')}</DropdownMenuLabel>
                          
                          <DropdownMenuItem onClick={() => onEdit(environment)}>
                            <Edit className="h-4 w-4 mr-2" />
                            {t('common.edit')}
                          </DropdownMenuItem>
                          
                          <DropdownMenuItem
                            onClick={() => onDelete(environment.id)}
                            className="text-destructive"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            {t('common.delete')}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            {/* 分页 */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  {t('common.page')} {params.page || 1} / {totalPages}
                </div>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange((params.page || 1) - 1)}
                    disabled={!params.page || params.page <= 1}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange((params.page || 1) + 1)}
                    disabled={!params.page || params.page >= totalPages}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default DevEnvironmentList; 