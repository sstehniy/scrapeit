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
import { ResultsFilters } from "../components/ResultsFilters";

const pageSize = 500;

export type SearchConfig = {
  offset: number;
  limit: number;
  endpointIds: string[];
  q: string;
  isArchive: boolean;
};

const defaultSearchConfig: SearchConfig = {
  offset: 0,
  limit: pageSize,
  endpointIds: [] as string[],
  q: "",
  isArchive: false,
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
  const [showConfirmGroupArchive, setShowConfirmGroupArchive] = useState<{
    isOpen: boolean;
    onConfirm: ((versionTag: string) => void) | null;
  }>({
    isOpen: false,
    onConfirm: () => {},
  });
  const queryClient = useQueryClient();
  const noSchemaMessageShown = useRef(false);

  const [searchConfig, setSearchConfig] =
    useState<SearchConfig>(defaultSearchConfig);

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

  const { data: scrapeResultsExist, refetch: refetchScrapeResultsAxist } =
    useQuery<boolean>({
      queryKey: ["scrapeResultsExist", groupId],
      queryFn: () =>
        axios
          .get(`/api/scrape/results/not-empty/${groupId}`)
          .then((res) => res.data.resultsNotEmpty),
      enabled: !!groupId,
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
    queryFn: ({ pageParam }) => {
      return axios
        .get(`/api/scrape/results/${groupId}`, {
          params: {
            offset: pageParam,
            limit: searchConfig.limit,
            endpointIds: searchConfig.endpointIds.join(","),
            q: searchConfig.q,
            isArchive: searchConfig.isArchive,
          },
        })
        .then((res) => res.data);
    },
    refetchInterval: 30000,
    getNextPageParam: (lastPage, pages) => {
      return lastPage.hasMore ? pages.length * searchConfig.limit : undefined;
    },

    initialPageParam: defaultSearchConfig.offset,
    enabled: !!groupId && searchConfig.endpointIds.length > 0,
  });

  useEffect(() => {
    if (!group) return;
    setSearchConfig((prev) => ({
      ...prev,
      endpointIds: group.endpoints.map((e) => e.id),
    }));
  }, [group]);

  const updateGroupSchemaMutation = useMutation({
    mutationFn: (data: {
      fields: Field[];
      changes: FieldChange[];
      versionTag?: string;
      shouldArchive?: boolean;
    }) => {
      return axios.put(`/api/scrape-groups/${groupId}/schema`, data);
    },
    onMutate: (data) => {
      setShowGroupSchemaSettings(false);
      setShowConfirmGroupArchive({
        isOpen: false,
        onConfirm: null,
      });
      if (data.shouldArchive)
        toast.info(
          "Updating group schema... Once completed, all current results will disappear",
        );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
      queryClient.invalidateQueries({
        queryKey: ["scrapeResultsExist", groupId],
      });
      toast.success(
        "Group schema updated successfully. Please modify the endpoints to start automatic scraping",
      );
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
    onMutate: () => {
      toast.info("Scraping all endpoints...");
    },
    onSuccess: () => {
      setSearchConfig((prev) => ({ ...prev, offset: 0 }));
      queryClient.invalidateQueries({ queryKey: ["groupResults", groupId] });
      toast.success("All endpoints scraped successfully");
    },
    onError: (err) => {
      console.log(err);
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      if (err.response.data.error) {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        toast.warn(err.response.data.error);
      } else {
        toast.error("Failed to scrape all endpoints");
      }
    },
  });

  useEffect(() => {
    if (!group || !!group.fields.length || noSchemaMessageShown.current) return;
    noSchemaMessageShown.current = true;
    setShowGroupSchemaSettings(true);
    toast.info("Please configure the group schema");
  }, [group]);

  const handleUpdateGroupSchema = useCallback(
    async (fields: Field[], changes: FieldChange[]) => {
      if (!group) return;

      const exist = await refetchScrapeResultsAxist();
      if (
        group.fields.length &&
        changes.some((c) => c.type === "add_field") &&
        exist.data
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
          },
        });
        return;
      } else {
        updateGroupSchemaMutation.mutate({ fields, changes });
        setShowGroupSchemaSettings(false);
      }
    },
    [group, refetchScrapeResultsAxist, updateGroupSchemaMutation],
  );

  const handleScrapeAllEndpoints = useCallback(() => {
    scrapeAllEndpointsMutation.mutate();
  }, [scrapeAllEndpointsMutation]);

  if (!groupId) {
    return <Navigate to="/not-found" />;
  }

  return (
    <div>
      <div ref={headerRef}>
        <div className="flex items-center gap-5 mb-5">
          <div className="flex">
            <Button
              onClick={() => navigate("/")}
              className="btn btn-accent btn-outline btn-sm mr-4 btn-square border-0"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={2.5}
                stroke="currentColor"
                className="size-5"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
                />
              </svg>
            </Button>
            <h1 className="text-3xl font-bold">{group?.name}</h1>
          </div>
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
        {group && (
          <ResultsFilters
            params={searchConfig}
            group={group}
            setParams={setSearchConfig}
          />
        )}
      </div>
      {showGroupSchemaSettings && (
        <ConfigureGroupSchema
          isOpen={showGroupSchemaSettings}
          onClose={() => setShowGroupSchemaSettings(false)}
          onConfirm={handleUpdateGroupSchema}
          fieldsToEdit={group?.fields || []}
        />
      )}
      {showConfirmGroupArchive.isOpen && (
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
      )}

      {!!scrapeResults?.pages.length && group && (
        <ResultsTable
          group={group}
          hasMore={hasNextPage}
          loadMore={() => {
            fetchNextPage();
          }}
          height={windowHeight - size.height - 95}
          loading={isLoading || isFetchingNextPage}
          results={scrapeResults.pages
            .flatMap((page) => page.results)
            .filter(Boolean)}
        />
      )}
    </div>
  );
};
