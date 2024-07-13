import { FC, useCallback, useEffect, useRef, useState } from "react";
import { Field, ScrapeGroup, ScrapeResult, SelectorStatus } from "../types";
import { Navigate, useNavigate, useParams } from "react-router-dom";
import axios from "axios";
import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { ResultsTable } from "../components/ResultsTable";
import { useComponentSize } from "../hooks/useComponentSize";
import { throttle } from "lodash";
import { GroupEndpoints } from "../components/GroupEndpoints";
import { toast } from "react-toastify";
import { ConfigureGroupSchema } from "../components/modals/ConfigureGroupSchema";
import { Button } from "../components/ui/Button";
import { ConfirmArchiveCurrentGroup } from "../components/modals/ConfirmArchiveCurrentGroup";

const pageSize = 20;

const defaultSearchConfig = {
  offset: 0,
  limit: pageSize,
  endpointIds: [] as string[],
};

export type FieldChange = {
  fieldId: string;
  fieldIsNewSinceLastSave: boolean;
  type:
    | "change_field_key"
    | "change_field_type"
    | "change_field_name"
    | "delete_field"
    | "add_field";
};

export const GroupView: FC = () => {
  const [showGroupSchemaSettings, setShowGroupSchemaSettings] = useState(false);
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();
  const [headerRef, size] = useComponentSize<HTMLDivElement>();
  const [windowHeight, setWindowHeight] = useState(window.innerHeight);
  const initialLoaded = useRef(false);
  const [showConfirmGroupArchive, setShowConfirmGroupArchive] = useState<{
    isOpen: boolean;
    onConfirm: ((versionTag: string) => void) | null;
  }>({
    isOpen: false,
    onConfirm: () => {},
  });
  const queryClient = useQueryClient();

  const [searchConfig, setSearchConfig] = useState({
    offset: defaultSearchConfig.offset,
    limit: defaultSearchConfig.limit,
    endpointIds: defaultSearchConfig.endpointIds,
  });

  const [searchConfigChanged, setSearchConfigChanged] = useState(false);

  useEffect(() => {
    const handleResize = throttle(() => {
      setWindowHeight(window.innerHeight);
    }, 100);
    handleResize();
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const { data: group, isLoading: groupLoading } = useQuery<ScrapeGroup>({
    queryKey: ["group", groupId],
    queryFn: () =>
      axios.get(`/api/scrape-groups/${groupId}`).then((res) => res.data),
    enabled: !!groupId,
    behavior: {
      onFetch: () => {
        queryClient.invalidateQueries({ queryKey: ["groupResults", groupId] });
      },
    },
  });

  const { data: scrapeResultsExist } = useQuery<boolean>({
    queryKey: ["scrapeResultsExist", groupId],
    queryFn: () =>
      axios
        .get(`/api/scrape/results/not-empty/${groupId}`)
        .then((res) => res.data.resultsNotEmpty),
  });

  const {
    data: scrapeResults,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery<{
    results: ScrapeResult[];
    hasMore: boolean;
  }>({
    queryKey: ["groupResults", groupId, searchConfig],
    queryFn: ({ pageParam = 0 }) => {
      return axios
        .get(`/api/scrape/results/${groupId}`, {
          params: {
            offset: pageParam,
            limit: searchConfig.limit,
            endpointIds: searchConfig.endpointIds.join(","),
          },
        })
        .then((res) => res.data);
    },
    getNextPageParam: (lastPage, pages) => {
      return lastPage.hasMore ? pages.length * searchConfig.limit : undefined;
    },

    initialPageParam: defaultSearchConfig.offset,
    enabled: !!groupId && searchConfig.endpointIds.length > 0,
  });

  const updateGroupSchemaMutation = useMutation({
    mutationFn: (data: {
      fields: Field[];
      changes: FieldChange[];
      versionTag?: string;
      shouldArchive?: boolean;
    }) => {
      console.log(data);
      return axios.put(`/api/scrape-groups/${groupId}/schema`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
      queryClient.invalidateQueries({
        queryKey: ["scrapeResultsExist", groupId],
      });
      toast.success("Group schema updated successfully");
    },
    onError: () => {
      toast.error("Failed to update group schema");
    },
  });

  const scrapeAllEndpointsMutation = useMutation({
    mutationFn: () =>
      axios.post(`/api/scrape/endpoints`, {
        groupId: group?.id,
        endpointIds: group?.endpoints.map((e) => e.id),
      }),
    onSuccess: () => {
      setSearchConfig({
        ...defaultSearchConfig,
        offset: 0,
        endpointIds: group?.endpoints.map((e) => e.id) || [],
      });
      queryClient.invalidateQueries({ queryKey: ["groupResults", groupId] });
      toast.success("All endpoints scraped successfully");
    },
    onError: () => {
      toast.error("Failed to scrape all endpoints");
    },
  });

  useEffect(() => {
    if (!group) return;
    if (!group.fields.length) {
      setShowGroupSchemaSettings(true);
      toast.info("Please configure the group schema");
    }
  }, [group]);

  useEffect(() => {
    if (!group || initialLoaded.current || !group.endpoints.length) return;
    initialLoaded.current = true;
    setSearchConfig((prev) => ({
      ...prev,
      endpointIds: group.endpoints.map((e) => e.id),
    }));
  }, [group]);

  const handleUpdateGroupSchema = useCallback(
    (fields: Field[], changes: FieldChange[]) => {
      if (!group) return;
      if (
        group.fields.length &&
        changes.some((c) => c.type === "add_field") &&
        scrapeResultsExist
      ) {
        setShowConfirmGroupArchive({
          isOpen: true,
          onConfirm: (versionTag: string) => {
            updateGroupSchemaMutation.mutate({
              fields,
              changes,
              versionTag,
              shouldArchive: true,
            });
            setShowGroupSchemaSettings(false);
            setShowConfirmGroupArchive({
              isOpen: false,
              onConfirm: null,
            });
          },
        });
        return;
      } else {
        updateGroupSchemaMutation.mutate({ fields, changes });
        setShowGroupSchemaSettings(false);
      }
    },
    [group, scrapeResultsExist, updateGroupSchemaMutation],
  );

  const handleScrapeAllEndpoints = useCallback(() => {
    scrapeAllEndpointsMutation.mutate();
  }, [scrapeAllEndpointsMutation]);

  const resetSearch = useCallback(() => {
    setSearchConfig({
      offset: defaultSearchConfig.offset,
      limit: defaultSearchConfig.limit,
      endpointIds: group?.endpoints.map((e) => e.id) || [],
    });
    setSearchConfigChanged(false);
    queryClient.invalidateQueries({ queryKey: ["groupResults", groupId] });
  }, [group, groupId, queryClient]);

  if (!groupId) {
    return <Navigate to="/not-found" />;
  }

  return (
    <div>
      <div ref={headerRef}>
        <div className="flex justify-between">
          <Button
            onClick={() => navigate("/")}
            className="btn btn-accent btn-sm mb-3"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={2.5}
              stroke="currentColor"
              className="size-4"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
              />
            </svg>
          </Button>
          {searchConfigChanged && (
            <Button
              onClick={resetSearch}
              className="btn btn-primary btn-sm mb-3"
            >
              Reset Search
            </Button>
          )}
        </div>
        <div className="flex items-center gap-5 mb-5">
          <h1 className="text-3xl font-bold">{group?.name}</h1>
          <Button
            onClick={() => setShowGroupSchemaSettings(true)}
            className="btn btn-secondary btn-outline btn-sm btn-square border-0"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="size-6"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
              />
            </svg>
          </Button>
        </div>
        {group && (
          <GroupEndpoints
            group={group}
            onEndpointChange={() => {
              queryClient.invalidateQueries({ queryKey: ["group", groupId] });
              queryClient.invalidateQueries({
                queryKey: ["scrapeResultsExist", groupId],
              });
            }}
            disabledTooltip={
              scrapeAllEndpointsMutation.isPending
                ? "Scraping is in progress"
                : group.endpoints.some((e) =>
                      e.detailFieldSelectors.some(
                        (s) => s.selectorStatus === SelectorStatus.NEW,
                      ),
                    )
                  ? "Some endpoints need update since last schema change"
                  : ""
            }
            allowScrapeAllEndpoints={
              !scrapeAllEndpointsMutation.isPending &&
              !group.endpoints.some((e) =>
                e.detailFieldSelectors.some(
                  (s) => s.selectorStatus === SelectorStatus.NEW,
                ),
              )
            }
            onScrapeAllEndpoints={handleScrapeAllEndpoints}
          />
        )}
      </div>
      <ConfigureGroupSchema
        isOpen={showGroupSchemaSettings}
        onClose={() => setShowGroupSchemaSettings(false)}
        onConfirm={handleUpdateGroupSchema}
        fieldsToEdit={group?.fields || []}
      />
      <ConfirmArchiveCurrentGroup
        isOpen={showConfirmGroupArchive.isOpen}
        onClose={() =>
          setShowConfirmGroupArchive({
            isOpen: false,
            onConfirm: null,
          })
        }
        onConfirm={showConfirmGroupArchive.onConfirm || (() => {})}
      />

      {!!scrapeResults?.pages.length && group && (
        <ResultsTable
          group={group}
          hasMore={hasNextPage}
          loadMore={() => {
            fetchNextPage();
          }}
          height={windowHeight - size.height - 150}
          loading={isLoading || isFetchingNextPage}
          results={scrapeResults.pages
            .flatMap((page) => page.results)
            .filter(Boolean)}
        />
      )}
    </div>
  );
};
