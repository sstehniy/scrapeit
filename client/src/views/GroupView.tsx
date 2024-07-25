import { type FC, useCallback, useEffect, useRef, useState } from "react";
import {
	type Field,
	type NotificationConfig,
	type ScrapeGroup,
	type ScrapeResult,
	SelectorStatus,
} from "../types";
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
import { ConfigureGroupNotificationModal } from "../components/modals/ConfigureGroupNotification";

const pageSize = 500;

export type SearchConfig = {
	offset: number;
	limit: number;
	endpointIds: string[];
	q: string;
	isArchive: boolean;
	filters: SearchFilter[];
	sort: SearchSort;
};

export type SearchSort = {
	fieldId: string;
	order: 1 | -1;
};

export type SearchFilter = {
	fieldId: string;
	value: number | string | null;
	operator: "=" | "!=" | ">" | "<";
};

const defaultSearchConfig: SearchConfig = {
	offset: 0,
	limit: pageSize,
	endpointIds: [] as string[],
	q: "",
	isArchive: false,
	filters: [],
	sort: {
		fieldId: "",
		order: 1,
	},
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

export const GroupView: FC<{
	archived?: boolean;
}> = ({ archived = false }) => {
	const [showGroupSchemaSettings, setShowGroupSchemaSettings] = useState(false);
	const [showGroupNotificationConfig, setShowGroupNotificationConfig] =
		useState(false);
	const { groupId } = useParams<{ groupId: string }>();
	const navigate = useNavigate();
	const [headerRef, size] = useComponentSize<HTMLDivElement>();
	const [windowHeight, setWindowHeight] = useState(window.innerHeight);
	const [showConfirmGroupArchive, setShowConfirmGroupArchive] = useState<{
		isOpen: boolean;
		onConfirm: ((versionTag: string) => void) | null;
	}>({
		isOpen: false,
		onConfirm: () => {
			return;
		},
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

	const { data: group } = useQuery<ScrapeGroup>({
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

	const { refetch: refetchScrapeResultsExist } = useQuery<boolean>({
		queryKey: ["scrapeResultsExist", groupId],
		queryFn: () =>
			axios
				.get(`/api/scrape/results/not-empty/${groupId}`)
				.then((res) => res.data.resultsNotEmpty),
		enabled: !!groupId && !archived,
	});

	const {
		data: scrapingGroupNotificationConfig,
		isLoading: scrapingGroupNotificationConfigLoading,
	} = useQuery<NotificationConfig>({
		queryKey: ["scrapingGroupNotificationSettings", groupId],
		queryFn: () =>
			axios
				.get(`/api/scrape-groups/${groupId}/notification-config`)
				.then((res) => res.data)
				.catch((err) => {
					if (err.response.status === 404) {
						return undefined;
					}
				}),
		enabled: !!groupId && !archived,
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
				.post("/api/scrape/results", {
					...searchConfig,
					offset: pageParam,
					groupId,
					isArchive: archived,
				})
				.then((res) => res.data);
		},
		refetchInterval: archived ? 30000 : 0,
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
			filters: [],
			endpointIds: group.endpoints?.map((e) => e.id) || [],
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
			axios.post(
				"/api/scrape/endpoints",
				{
					groupId: group?.id,
					endpointIds: group?.endpoints.map((e) => e.id),
				},
				// 10 Minutes
				{ timeout: 600000 },
			),
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

	const updateGroupNotificationConfigMutation = useMutation({
		mutationFn: (config: NotificationConfig) =>
			axios.put(`/api/scrape-groups/${groupId}/notification-config`, config),
		onMutate: () => {
			setShowGroupNotificationConfig(false);
			toast.info("Updating group notification config...");
		},
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: ["scrapingGroupNotificationSettings", groupId],
			});
			toast.success("Group notification config updated successfully");
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

			const exist = await refetchScrapeResultsExist();
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
			}
			updateGroupSchemaMutation.mutate({ fields, changes });
			setShowGroupSchemaSettings(false);
		},
		[group, refetchScrapeResultsExist, updateGroupSchemaMutation],
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
					{!archived && (
						<>
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

							<Button
								onClick={() => setShowGroupNotificationConfig(true)}
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
										d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0M3.124 7.5A8.969 8.969 0 0 1 5.292 3m13.416 0a8.969 8.969 0 0 1 2.168 4.5"
									/>
								</svg>
							</Button>
						</>
					)}
				</div>
				{group && (
					<GroupEndpoints
						group={group}
						archived={archived}
						onEndpointChange={() => {
							queryClient.invalidateQueries({ queryKey: ["group", groupId] });
							queryClient.invalidateQueries({
								queryKey: ["scrapeResultsExist", groupId],
							});
						}}
						disabledTooltip={
							scrapeAllEndpointsMutation.isPending
								? "Scraping is in progress"
								: group.endpoints?.some((e) =>
											e.detailFieldSelectors.some(
												(s) => s.selectorStatus === SelectorStatus.NEW,
											),
										)
									? "Some endpoints need update since last schema change"
									: ""
						}
						allowScrapeAllEndpoints={
							!scrapeAllEndpointsMutation.isPending &&
							!group.endpoints?.some((e) =>
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
			{showGroupNotificationConfig &&
				!scrapingGroupNotificationConfigLoading &&
				group && (
					<ConfigureGroupNotificationModal
						isOpen={showGroupNotificationConfig}
						onClose={() => setShowGroupNotificationConfig(false)}
						groupNotificationConfig={scrapingGroupNotificationConfig}
						onConfirm={(config) => {
							updateGroupNotificationConfigMutation.mutate(config);
						}}
						group={group}
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
					onConfirm={
						showConfirmGroupArchive.onConfirm ||
						(() => {
							return;
						})
					}
				/>
			)}

			{!!scrapeResults?.pages.length && group && (
				<ResultsTable
					group={group}
					hasMore={hasNextPage}
					loadMore={() => {
						fetchNextPage();
					}}
					height={windowHeight - size.height - 85}
					loading={isLoading || isFetchingNextPage}
					results={scrapeResults.pages
						.flatMap((page) => page.results)
						.filter(Boolean)}
				/>
			)}
		</div>
	);
};
