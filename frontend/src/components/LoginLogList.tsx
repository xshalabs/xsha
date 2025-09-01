import React, { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "react-router-dom";
import { apiService } from "@/lib/api/index";
import { logError, handleApiError } from "@/lib/errors";
import { toast } from "sonner";
import { formatDateToLocal } from "@/lib/utils";
import { formatDateRangeForAPI } from "@/lib/timezone";
import { useTranslation } from "react-i18next";
import {
  Section,
  SectionHeader,
  SectionTitle,
  SectionDescription,
} from "@/components/content/section";
import { DataTable } from "@/components/ui/data-table";
import { useLoginLogColumns } from "@/components/data-table/login-logs/columns";
import { LoginLogDataTableToolbar } from "@/components/data-table/login-logs/data-table-toolbar";
import { CustomPagination } from "@/components/data-table/login-logs/custom-pagination";
import type { LoginLog } from "@/types/admin-logs";
import type { ColumnFiltersState } from "@tanstack/react-table";

export const LoginLogList: React.FC = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const [logs, setLogs] = useState<LoginLog[]>([]);
  const [loading, setLoading] = useState(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;

  // DataTable state
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  // Prevent duplicate requests
  const lastRequestRef = useRef<string>("");

  // Initialize component (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  // Track initial column filters to prevent duplicate requests
  const initialColumnFiltersRef = useRef<ColumnFiltersState>([]);

  // Unified data loading function with debouncing and duplicate request prevention
  const loadLoginLogsData = useCallback(
    async (
      page: number,
      filters: ColumnFiltersState,
      shouldDebounce = true,
      updateUrl = true
    ) => {
      const requestKey = `${page}-${JSON.stringify(filters)}`;

      // Prevent duplicate requests
      if (lastRequestRef.current === requestKey) {
        return;
      }

      if (shouldDebounce) {
        // Debounce to prevent rapid duplicate requests
        const debounceTimer = setTimeout(async () => {
          if (lastRequestRef.current === requestKey) {
            return; // Request was cancelled
          }

          lastRequestRef.current = requestKey;
          await executeRequest();
        }, 300);

        // Store timer for potential cleanup
        return () => clearTimeout(debounceTimer);
      } else {
        lastRequestRef.current = requestKey;
        await executeRequest();
      }

      async function executeRequest() {
        try {
          setLoading(true);

          // Convert DataTable filters to API parameters
          const apiParams: any = {
            page,
            page_size: pageSize,
          };

          // Handle column filters
          filters.forEach((filter) => {
            if (filter.id === "username" && filter.value) {
              apiParams.username = filter.value;
            } else if (filter.id === "ip" && filter.value) {
              apiParams.ip = filter.value;
            } else if (
              filter.id === "success" &&
              Array.isArray(filter.value) &&
              filter.value.length > 0
            ) {
              apiParams.success = filter.value[0] === "true";
            } else if (filter.id === "login_time" && filter.value) {
              const dateRange = filter.value as {
                startDate?: Date;
                endDate?: Date;
              };
              const apiTimeParams = formatDateRangeForAPI(dateRange.startDate, dateRange.endDate);
              if (apiTimeParams.start_time) {
                apiParams.start_time = apiTimeParams.start_time;
              }
              if (apiTimeParams.end_time) {
                apiParams.end_time = apiTimeParams.end_time;
              }
            }
          });

          const response = await apiService.adminLogs.getLoginLogs(apiParams);

          setLogs(response.logs);
          setTotal(response.total);
          setTotalPages(response.total_pages);
          setCurrentPage(page);

          // Update URL parameters
          if (updateUrl) {
            const params = new URLSearchParams();

            // Add filter parameters
            filters.forEach((filter) => {
              if (filter.value) {
                if (filter.id === "login_time") {
                  // Handle date range filter
                  const { startDate, endDate } = filter.value as {
                    startDate?: Date;
                    endDate?: Date;
                  };
                  if (startDate) {
                    params.set("start_date", formatDateToLocal(startDate));
                  }
                  if (endDate) {
                    params.set("end_date", formatDateToLocal(endDate));
                  }
                } else if (
                  Array.isArray(filter.value) &&
                  filter.value.length > 0
                ) {
                  // Handle array filters (success)
                  params.set(filter.id, filter.value[0]);
                } else if (
                  typeof filter.value === "string" &&
                  filter.value.trim()
                ) {
                  // Handle string filters (username, ip)
                  params.set(filter.id, filter.value);
                }
              }
            });

            // Add page parameter (only if not page 1)
            if (page > 1) {
              params.set("page", String(page));
            }

            // Update URL without causing navigation
            setSearchParams(params, { replace: true });
          }
        } catch (err: any) {
          logError(err, "Failed to load login logs");
          console.error("Failed to load login logs:", err);
          toast.error(handleApiError(err));
        } finally {
          setLoading(false);
          // Clear the request tracking after a short delay
          setTimeout(() => {
            if (lastRequestRef.current === requestKey) {
              lastRequestRef.current = "";
            }
          }, 300);
        }
      }
    },
    [pageSize, setSearchParams]
  );

  const handlePageChange = useCallback(
    (page: number) => {
      loadLoginLogsData(page, columnFilters);
    },
    [columnFilters, loadLoginLogsData]
  );

  // Create columns
  const columns = useLoginLogColumns();

  // Initialize from URL on component mount (only once)
  useEffect(() => {
    // Get URL params directly to avoid dependency issues
    const usernameParam = searchParams.get("username");
    const ipParam = searchParams.get("ip");
    const successParam = searchParams.get("success");
    const startDateParam = searchParams.get("start_date");
    const endDateParam = searchParams.get("end_date");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (usernameParam) {
      initialFilters.push({ id: "username", value: usernameParam });
    }

    if (ipParam) {
      initialFilters.push({ id: "ip", value: ipParam });
    }

    if (successParam) {
      initialFilters.push({ id: "success", value: [successParam] });
    }

    if (startDateParam || endDateParam) {
      const dateFilter: { startDate?: Date; endDate?: Date } = {};
      if (startDateParam) {
        dateFilter.startDate = new Date(startDateParam);
      }
      if (endDateParam) {
        dateFilter.endDate = new Date(endDateParam);
      }
      initialFilters.push({ id: "login_time", value: dateFilter });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);

    // Store initial filters to prevent duplicate requests
    initialColumnFiltersRef.current = initialFilters;

    // Load initial data using the unified function
    loadLoginLogsData(initialPage, initialFilters, false, false).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Handle column filter changes (skip initial load)
  useEffect(() => {
    if (isInitialized) {
      // Check if columnFilters has actually changed from initial values
      const hasFilterChanged =
        JSON.stringify(columnFilters) !==
        JSON.stringify(initialColumnFiltersRef.current);

      if (hasFilterChanged) {
        loadLoginLogsData(1, columnFilters); // Reset to page 1 when filtering
        // Update the reference to current filters
        initialColumnFiltersRef.current = columnFilters;
      }
    }
  }, [columnFilters, isInitialized, loadLoginLogsData]);

  return (
    <>
      <Section>
        <SectionHeader>
          <SectionTitle>{t("adminLogs.loginLogs.title")}</SectionTitle>
          <SectionDescription>
            {t("adminLogs.loginLogs.description")}
          </SectionDescription>
        </SectionHeader>
      </Section>
      <Section>
        <div className="space-y-4">
          <DataTable
            columns={columns}
            data={logs}
            loading={loading}
            toolbarComponent={LoginLogDataTableToolbar}
            columnFilters={columnFilters}
            setColumnFilters={setColumnFilters}
          />
          <CustomPagination
            currentPage={currentPage}
            totalPages={totalPages}
            total={total}
            onPageChange={handlePageChange}
          />
        </div>
      </Section>
    </>
  );
};
