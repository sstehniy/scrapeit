import type { FC } from "react";
import { NavLink } from "react-router-dom";
import type { ScrapeGroup } from "../../types";
import { getBaseUrl } from "../../util/url";
import { WithTooltip } from "../ui/WithTooltip";

export const GroupCard: FC<{
	group: ScrapeGroup;
	archived?: boolean;
	onArchive?: () => void;
	onDelete?: () => void;
	onExport?: () => void;
}> = ({ group, onArchive, onDelete, onExport, archived = false }) => {
	return (
		<div className="card bg-base-300 shadow-xl hover:shadow-2xl transition-shadow duration-300">
			<NavLink
				to={`/group/${archived ? "archived/" : ""}${group.id}`}
				className="card-body p-5 px-7 flex flex-col"
			>
				<h2 className="card-title text-2xl">
					{group.name}{" "}
					{group.versionTag && (
						<span className="text-sm text-gray-500">({group.versionTag})</span>
					)}
				</h2>
				<div className="flex-1 ps-5">
					{group.endpoints.map((endpoint) => (
						<div key={endpoint.id} className="mt-1">
							<h3 className="font-bold inline-block mr-2">{endpoint.name}</h3>
							<span className="text-sm text-gray-500">
								({getBaseUrl(endpoint.url)})
							</span>
						</div>
					))}
				</div>
				<div className="card-actions justify-end">
					{onArchive && (
						<button
							className="btn btn-sm btn-outline"
							onClick={(e) => {
								e.preventDefault();
								onArchive();
							}}
						>
							Archive
						</button>
					)}
					{onExport && (
						<WithTooltip tooltip="Export Group">
							<button
								className="btn btn-sm btn-outline btn-square"
								onClick={(e) => {
									e.preventDefault();
									onExport();
								}}
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									strokeWidth={1.5}
									stroke="currentColor"
									className="size-5"
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M9 8.25H7.5a2.25 2.25 0 0 0-2.25 2.25v9a2.25 2.25 0 0 0 2.25 2.25h9a2.25 2.25 0 0 0 2.25-2.25v-9a2.25 2.25 0 0 0-2.25-2.25H15M9 12l3 3m0 0 3-3m-3 3V2.25"
									/>
								</svg>
							</button>
						</WithTooltip>
					)}
					{onDelete && (
						<button
							className="btn btn-sm btn-outline btn-error btn-square"
							onClick={(e) => {
								e.preventDefault();
								onDelete();
							}}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								strokeWidth={1.5}
								stroke="currentColor"
								className="size-4"
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
								/>
							</svg>
						</button>
					)}
				</div>
			</NavLink>
		</div>
	);
};
