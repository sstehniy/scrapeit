import { FC, useEffect, useState } from "react";
import { ArchivedScrapeGroup, ScrapeGroup } from "../types";
import axios from "axios";
import { toast } from "react-toastify";
import { GroupCard } from "../components/GroupCard";
import { CreateGroupModal } from "../components/modals/CreateGroup";
import { useNavigate } from "react-router-dom";
import { ConfirmDeleteGroup } from "../components/modals/ConfirmDeleteGroup";

export const GroupsOverview: FC = () => {
	const [showArchivedGroups, setShowArchivedGroups] = useState(false);

	return (
		<div className="container mx-auto px-4 py-8">
			<div role="tablist" className="tabs tabs-boxed  w-72 mb-4 bg-transparent">
				<div
					role="tab"
					className={`tab ${!showArchivedGroups ? "tab-active" : ""}`}
					onClick={() => setShowArchivedGroups(false)}
				>
					Active Groups
				</div>
				<div
					role="tab"
					className={`tab ${showArchivedGroups ? "tab-active" : ""}`}
					onClick={() => setShowArchivedGroups(true)}
				>
					Archived Groups
				</div>
			</div>
			{showArchivedGroups ? <ArchivedGroups /> : <ActiveGroups />}
		</div>
	);
};

const ArchivedGroups = () => {
	const [scrapeGroups, setScrapeGroups] = useState<
		ArchivedScrapeGroup[] | null
	>(null);

	useEffect(() => {
		axios.get(`/api/scrape-groups/archived`).then((data) => {
			setScrapeGroups(data.data);
		});
	}, []);

	const groupedByOriginalId = scrapeGroups?.reduce(
		(acc, group) => {
			if (!acc[group.originalId]) {
				acc[group.originalId] = [];
			}
			acc[group.originalId].push(group);
			return acc;
		},
		{} as { [key: string]: ScrapeGroup[] },
	);

	return (
		<>
			{scrapeGroups === null ? (
				<div className="flex justify-center items-center h-64">
					<span className="loading loading-spinner loading-lg text-primary"></span>
				</div>
			) : scrapeGroups.length === 0 ? (
				<div className="text-center py-12">
					<p className="text-xl text-gray-600">No archived groups found.</p>
				</div>
			) : (
				<div>
					{Object.entries(groupedByOriginalId || {}).map(
						([originalId, groups]) => (
							<div key={originalId} className="mb-8">
								<h2 className="text-2xl font-bold mb-4">
									{groups[0].name} ({groups.length} versions)
								</h2>
								<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
									{groups.map((group) => (
										<GroupCard key={group.id} group={group} archived />
									))}
								</div>
							</div>
						),
					)}
				</div>
			)}
		</>
	);
};

const ActiveGroups = () => {
	const [scrapeGroups, setScrapeGroups] = useState<ScrapeGroup[] | null>(null);
	const [showCreateGroupModal, setShowCreateGroupModal] = useState(false);
	const navigate = useNavigate();
	const [showConfirmDeleteGroupModal, setShowConfirmDeleteGroupModal] =
		useState<{
			isOpen: boolean;
			onConfirm: () => void;
		} | null>(null);
	useEffect(() => {
		axios.get(`/api/scrape-groups`).then((data) => {
			setScrapeGroups(data.data);
		});
	}, []);

	const handleExportGroupConfig = (group: ScrapeGroup) => {
		const json = JSON.stringify(group, null, 2);
		const blob = new Blob([json], { type: "application/json" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `${group.name}_config.json`;
		a.click();
		URL.revokeObjectURL(url);
	};

	const handleCreateGroup = (name: string) => {
		axios
			.post(`/api/scrape-groups`, { name })
			.then((response) => {
				toast.success("Group created");
				setShowCreateGroupModal(false);
				setScrapeGroups((prev) =>
					prev ? [...prev, response.data] : [response.data],
				);
				navigate(`/group/${response.data.id}`);
			})
			.catch((error) => {
				console.error(error);
				toast.error("Failed to create group");
			});
	};

	return (
		<>
			<div className="flex items-end mb-6 gap-4">
				<button
					className="btn btn-primary btn-sm"
					onClick={() => setShowCreateGroupModal(true)}
				>
					+ Create new Group
				</button>
				<label className="form-control w-full max-w-xs">
					<div className="label">
						<span className="label-text">Import Group</span>
					</div>
					<input
						type="file"
						className="file-input file-input-sm file-input-bordered w-64"
						accept=".json"
						placeholder="Import group config"
						onChange={(e) => {
							const file = e.target.files?.[0];
							if (!file) return;
							const reader = new FileReader();
							reader.onload = (e) => {
								const content = e.target?.result;
								if (typeof content !== "string") return;
								try {
									const group = JSON.parse(content);
									axios
										.post(`/api/scrape-groups`, group)
										.then((response) => {
											toast.success("Group imported");
											setScrapeGroups((prev) =>
												prev ? [...prev, response.data] : [response.data],
											);
										})
										.catch((error) => {
											console.error(error);
											toast.error("Failed to import group");
										});
								} catch (error) {
									console.error(error);
									toast.error("Failed to parse group config");
								}
							};
							reader.readAsText(file);
						}}
					/>
				</label>
			</div>
			{scrapeGroups === null ? (
				<div className="flex justify-center items-center h-64">
					<span className="loading loading-spinner loading-lg text-primary"></span>
				</div>
			) : scrapeGroups.length === 0 ? (
				<div className="text-center py-12">
					<p className="text-xl text-gray-600">
						No groups found. Create your first group!
					</p>
				</div>
			) : (
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{scrapeGroups.map((group) => (
						<GroupCard
							key={group.id}
							group={group}
							onExport={() => handleExportGroupConfig(group)}
							onDelete={() => {
								setShowConfirmDeleteGroupModal({
									isOpen: true,
									onConfirm: () => {
										axios
											.delete(`/api/scrape-groups/${group.id}`)
											.then(() => {
												setScrapeGroups(
													(prev) =>
														prev?.filter((g) => g.id !== group.id) || null,
												);
												toast.success("Group deleted");
											})
											.catch((error) => {
												console.error(error);
												toast.error("Failed to delete group");
											})
											.finally(() => {
												setShowConfirmDeleteGroupModal(null);
											});
									},
								});
							}}
						/>
					))}
				</div>
			)}
			{showCreateGroupModal && (
				<CreateGroupModal
					isOpen={showCreateGroupModal}
					onClose={() => setShowCreateGroupModal(false)}
					onConfirm={handleCreateGroup}
				/>
			)}
			{showConfirmDeleteGroupModal && (
				<ConfirmDeleteGroup
					isOpen={showConfirmDeleteGroupModal.isOpen}
					onClose={() => setShowConfirmDeleteGroupModal(null)}
					onConfirm={showConfirmDeleteGroupModal.onConfirm}
				/>
			)}
		</>
	);
};
