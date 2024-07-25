import { FC, useEffect, useState } from "react";
import { ArchivedScrapeGroup, ScrapeGroup } from "../types";
import axios from "axios";
import { toast } from "react-toastify";
import { GroupCard } from "../components/GroupCard";
import { CreateGroupModal } from "../components/modals/CreateGroup";
import { useNavigate } from "react-router-dom";

export const GroupsOverview: FC = () => {
	const [showArchivedGroups, setShowArchivedGroups] = useState(false);

	return (
		<div className="container mx-auto px-4 py-8">
			<div role="tablist" className="tabs tabs-bordered  tabs-md w-96 mb-8">
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
										<GroupCard key={group.id} group={group} />
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

	useEffect(() => {
		axios.get(`/api/scrape-groups`).then((data) => {
			setScrapeGroups(data.data);
		});
	}, []);

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
			<div className="flex justify-between items-center mb-6">
				<button
					className="btn btn-primary btn-sm"
					onClick={() => setShowCreateGroupModal(true)}
				>
					+ Create new Group
				</button>
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
							onDelete={() => {
								axios
									.delete(`/api/scrape-groups/${group.id}`)
									.then(() => {
										setScrapeGroups(
											(prev) => prev?.filter((g) => g.id !== group.id) || null,
										);
										toast.success("Group deleted");
									})
									.catch((error) => {
										console.error(error);
										toast.error("Failed to delete group");
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
		</>
	);
};
