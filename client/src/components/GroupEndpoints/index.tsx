import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { type FC, useCallback, useState } from "react";
import { toast } from "react-toastify";
import { v4 } from "uuid";
import { type Endpoint, type ScrapeGroup, SelectorStatus } from "../../types";
import { getBaseUrl } from "../../util/url";
import { getColoredEndpointPill } from "../ColoredEndpointPill";
import { ConfigureGroupEndpoint } from "../modals/ConfigureGroupEndpoint";
import { ConfirmRemoveEndpoint } from "../modals/ConfirmRemoveEndpoint";
import { ConfirmRemoveEndpointResults } from "../modals/ConfirmRemoveEndpointResults";
import { Button } from "../ui/Button";
import { WithTooltip } from "../ui/WithTooltip";

type GroupEndpointsProps = {
	group: ScrapeGroup;
	onEndpointChange: () => void | Promise<void>;
	allowScrapeAllEndpoints: boolean;
	disabledTooltip: string;
	onScrapeAllEndpoints: () => void | Promise<void>;
	archived: boolean;
};

export const GroupEndpoints: FC<GroupEndpointsProps> = ({
	group,
	onEndpointChange,
	allowScrapeAllEndpoints,
	disabledTooltip,
	onScrapeAllEndpoints,
	archived,
}) => {
	const { endpoints } = group;
	const [showCreateEndpointModal, setShowCreateEndpointModal] = useState(false);
	const [showEditEndpointModal, setShowEditEndpointModal] = useState<{
		isOpen: boolean;
		endpoint: Endpoint | null;
	}>({
		isOpen: false,
		endpoint: null,
	});
	const [showConfirmRemoveEndpointModal, setShowConfirmRemoveEndpointModal] =
		useState<{
			isOpen: boolean;
			onConfirm: () => void;
		}>({
			isOpen: false,
			onConfirm: () => {
				return;
			},
		});
	const [
		showConfirmRemoveEndpointResultsModal,
		setShowConfirmRemoveEndpointResultsModal,
	] = useState<{
		isOpen: boolean;
		onConfirm: () => void;
	}>({
		isOpen: false,
		onConfirm: () => {
			return;
		},
	});

	const deleteEndpointResultsMutation = useMutation({
		mutationFn: ({
			endpointId,
			groupId,
		}: { groupId: string; endpointId: string }) =>
			axios.delete(
				`/api/scrape-groups/${groupId}/endpoints/results/${endpointId}`,
			),
		onMutate: () => {
			setShowConfirmRemoveEndpointResultsModal({
				isOpen: false,
				onConfirm: () => {
					return;
				},
			});
		},
		onSuccess: onEndpointChange,
		onError: (e) => {
			toast.error("Failed to delete endpoint results");
			console.error(e);
		},
	});

	const handleDeleteEndpointResults = useCallback(
		async (endpointId: string) => {
			setShowConfirmRemoveEndpointResultsModal({
				isOpen: true,
				onConfirm: () => {
					deleteEndpointResultsMutation.mutate({
						groupId: group.id,
						endpointId,
					});
				},
			});
		},
		[group.id, deleteEndpointResultsMutation],
	);

	const deleteEndpointMutation = useMutation({
		mutationFn: ({
			endpointId,
			groupId,
		}: { groupId: string; endpointId: string }) =>
			axios.delete(`/api/scrape-groups/${groupId}/endpoints/${endpointId}`),
		onMutate: () => {
			setShowConfirmRemoveEndpointModal({
				isOpen: false,
				onConfirm: () => {
					return;
				},
			});
		},
		onSuccess: onEndpointChange,
		onError: (e) => {
			toast.error("Failed to delete endpoint");
			console.error(e);
		},
	});
	const handleDeleteEndpoint = useCallback(
		async (endpointId: string) => {
			setShowConfirmRemoveEndpointModal({
				isOpen: true,
				onConfirm: () => {
					deleteEndpointMutation.mutate({ endpointId, groupId: group.id });
				},
			});
		},
		[group.id, deleteEndpointMutation],
	);

	const createEndpointMutation = useMutation({
		mutationFn: ({
			endpoint,
			groupId,
		}: { groupId: string; endpoint: Endpoint }) =>
			axios.post(`/api/scrape-groups/${groupId}/endpoints`, {
				endpoint,
			}),
		onSuccess: onEndpointChange,
		onError: (e) => {
			toast.error("Failed to create endpoint");
			console.error(e);
		},
	});

	const editEndpointMutation = useMutation({
		mutationFn: ({
			endpoint,
			groupId,
		}: { groupId: string; endpoint: Endpoint }) =>
			axios.put(`/api/scrape-groups/${groupId}/endpoints/${endpoint.id}`, {
				endpoint,
			}),
		onSuccess: onEndpointChange,
		onError: (e) => {
			toast.error("Failed to save endpoint");
			console.error(e);
		},
	});

	const handleCreateEndpoint = useCallback(
		async (endpoint: Endpoint, type: "create" | "save") => {
			if (type === "create") {
				createEndpointMutation.mutate({ endpoint, groupId: group.id });
			} else {
				editEndpointMutation.mutate({ endpoint, groupId: group.id });
			}
		},
		[group.id, editEndpointMutation, createEndpointMutation],
	);

	const handleExportGroupEndpointConfig = (ep: Endpoint) => {
		const json = JSON.stringify(ep, null, 2);
		const blob = new Blob([json], { type: "application/json" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `endpoint_${group.name}${group.versionTag ? "_" + group.versionTag : ""}_${ep.name}_config.json`;
		a.click();
		URL.revokeObjectURL(url);
	};

	return (
		<div>
			{!archived && (
				<div className="flex items-end  mb-5 gap-5">
					<Button
						className="btn btn-primary btn-sm"
						onClick={() => {
							setShowCreateEndpointModal(true);
						}}
						disabled={group.fields.length === 0}
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
								d="M12 4.5v15m7.5-7.5h-15"
							/>
						</svg>
						New Endpoint
					</Button>
					{!!group.endpoints?.length && (
						<WithTooltip
							tooltip={!allowScrapeAllEndpoints ? disabledTooltip : ""}
						>
							<Button
								className="btn btn-primary btn-sm"
								onClick={onScrapeAllEndpoints}
								disabled={!allowScrapeAllEndpoints}
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
										d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
									/>
								</svg>
								Scrape All Endpoints
							</Button>
						</WithTooltip>
					)}
				</div>
			)}

			{endpoints?.length === 0 && (
				<p className="text-base-content">No endpoints configured yet</p>
			)}
			<div className="flex flex-wrap gap-1">
				{endpoints?.map((endpoint) => (
					<div
						key={endpoint.id}
						className="mr-2 mb-2 p-2 border-2 rounded-lg border-base-content border-opacity-25 bg-base-100 text-base-content flex items-center"
					>
						<div className="flex">
							<div className="mr-3">
								{getColoredEndpointPill(endpoint.id, group)}
							</div>
							<div className="mr-4">{getBaseUrl(endpoint.url, false)}</div>
						</div>
						<div className="flex gap-3">
							{endpoint.detailFieldSelectors.some(
								(ds) => ds.selectorStatus !== SelectorStatus.OK,
							) && (
								<div
									className="tooltip"
									data-tip="This endpoint needs update since last schema change"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="currentColor"
										className="size-5 text-warning"
									>
										<path
											fillRule="evenodd"
											d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003ZM12 8.25a.75.75 0 0 1 .75.75v3.75a.75.75 0 0 1-1.5 0V9a.75.75 0 0 1 .75-.75Zm0 8.25a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z"
											clipRule="evenodd"
										/>
									</svg>
								</div>
							)}
							<Button
								onClick={() => {
									handleExportGroupEndpointConfig(endpoint);
								}}
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									strokeWidth={1.5}
									stroke="currentColor"
									className="size-5 hover:text-success"
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M9 8.25H7.5a2.25 2.25 0 0 0-2.25 2.25v9a2.25 2.25 0 0 0 2.25 2.25h9a2.25 2.25 0 0 0 2.25-2.25v-9a2.25 2.25 0 0 0-2.25-2.25H15M9 12l3 3m0 0 3-3m-3 3V2.25"
									/>
								</svg>
							</Button>
							{!archived && (
								<>
									<Button
										onClick={() => {
											const endpointCopy = JSON.parse(
												JSON.stringify(endpoint),
											) as Endpoint;
											endpointCopy.id = v4();
											endpointCopy.active = false;
											endpointCopy.detailFieldSelectors.forEach((dfs) => {
												dfs.lockedForEdit = false;
												dfs.selectorStatus = SelectorStatus.NEW;
											});

											endpointCopy.name = `${endpoint.name} - Copy`;
											endpointCopy.detailFieldSelectors =
												endpointCopy.detailFieldSelectors.map((dfs) => {
													dfs.id = v4();
													return dfs;
												});
											handleCreateEndpoint(endpointCopy, "create");
										}}
									>
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											strokeWidth={1.5}
											stroke="currentColor"
											className="size-5 hover:text-warning"
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"
											/>
										</svg>
									</Button>
									<Button
										onClick={() => {
											handleDeleteEndpointResults(endpoint.id);
										}}
									>
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											strokeWidth={1.5}
											stroke="currentColor"
											className="size-5 hover:text-warning"
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="m20.25 7.5-.625 10.632a2.25 2.25 0 0 1-2.247 2.118H6.622a2.25 2.25 0 0 1-2.247-2.118L3.75 7.5m6 4.125 2.25 2.25m0 0 2.25 2.25M12 13.875l2.25-2.25M12 13.875l-2.25 2.25M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125Z"
											/>
										</svg>
									</Button>
									<Button
										onClick={() => {
											setShowEditEndpointModal({
												isOpen: true,
												endpoint,
											});
										}}
									>
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											strokeWidth={1.5}
											stroke="currentColor"
											className="size-5 hover:text-primary"
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
									<Button>
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											strokeWidth={1.5}
											stroke="currentColor"
											className="size-5 hover:text-error"
											onClick={() => handleDeleteEndpoint(endpoint.id)}
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
											/>
										</svg>
									</Button>
								</>
							)}
						</div>
					</div>
				))}
			</div>

			{showCreateEndpointModal && (
				<ConfigureGroupEndpoint
					fields={group.fields}
					isOpen={showCreateEndpointModal}
					onClose={() => setShowCreateEndpointModal(false)}
					onConfirm={(endpoint) => handleCreateEndpoint(endpoint, "create")}
				/>
			)}
			{showEditEndpointModal.isOpen && (
				<ConfigureGroupEndpoint
					fields={group.fields}
					isOpen={showEditEndpointModal.isOpen}
					onClose={() =>
						setShowEditEndpointModal({
							isOpen: false,
							endpoint: null,
						})
					}
					onConfirm={(endpoint) => handleCreateEndpoint(endpoint, "save")}
					editEndpoint={showEditEndpointModal.endpoint ?? undefined}
				/>
			)}
			{showConfirmRemoveEndpointResultsModal.isOpen && (
				<ConfirmRemoveEndpointResults
					isOpen={showConfirmRemoveEndpointResultsModal.isOpen}
					onClose={() =>
						setShowConfirmRemoveEndpointResultsModal({
							isOpen: false,
							onConfirm: () => {
								return;
							},
						})
					}
					onConfirm={showConfirmRemoveEndpointResultsModal.onConfirm}
				/>
			)}
			{showConfirmRemoveEndpointModal.isOpen && (
				<ConfirmRemoveEndpoint
					isOpen={showConfirmRemoveEndpointModal.isOpen}
					onClose={() =>
						setShowConfirmRemoveEndpointModal({
							isOpen: false,
							onConfirm: () => {
								return;
							},
						})
					}
					onConfirm={showConfirmRemoveEndpointModal.onConfirm}
				/>
			)}
		</div>
	);
};
