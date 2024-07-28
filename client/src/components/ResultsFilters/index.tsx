import { type FC, useEffect, useState } from "react";
import {
	type Endpoint,
	type ExportType,
	type Field,
	FieldType,
	type ScrapeGroup,
} from "../../types";
import type {
	SearchConfig,
	SearchFilter,
	SearchSort,
} from "../../views/GroupView";
import { TextInput } from "../ui/TextInput";
import "./index.css";
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "react-toastify";
import { ConfigureExportModal } from "../modals/ConfigureExport";

type ResultsFiltersProps = {
	params: SearchConfig;
	group: ScrapeGroup;
	setParams: (params: SearchConfig) => void;
};

export const ResultsFilters: FC<ResultsFiltersProps> = ({
	params,
	group,
	setParams,
}) => {
	const [q, setQ] = useState("");
	const [limit, setLimit] = useState(30);
	const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
	const [filters, setFilters] = useState<SearchFilter[]>([]);
	const [showExportOptionsModal, setShowExportOptionsModal] = useState<{
		isOpen: boolean;
		onConfirm: ((name: string, type: ExportType) => void) | null;
	}>({
		isOpen: false,
		onConfirm: null,
	});
	const [hoveredFilter, setHoveredFilter] = useState<number | null>(null);

	useEffect(() => {
		const endpoints = group.endpoints?.filter((e) =>
			params.endpointIds.includes(e.id),
		);
		const q = params.q;
		const limit = params.limit;
		setEndpoints(endpoints);
		setQ(q);
		setLimit(limit);
	}, [group.endpoints, params.endpointIds, params.limit, params.q]);

	const exportMutation = useMutation({
		mutationFn: (data: { name: string; type: ExportType }) =>
			axios
				.post(
					`/api/scrape/results/export/${group.id}`,
					{
						type: data.type,
						fileName: data.name,
						deleteAfterExport: false,
					},
					{
						responseType: "blob",
					},
				)
				.then((response) => {
					const url = window.URL.createObjectURL(new Blob([response.data]));
					const a = document.createElement("a");
					a.style.display = "none";
					a.href = url;
					a.download = `${data.name}.${data.type}`;

					document.body.appendChild(a);
					a.click();
					document.body.removeChild(a);

					window.URL.revokeObjectURL(url);
				}),
		onMutate: () => {
			setShowExportOptionsModal({ isOpen: false, onConfirm: null });
		},
		onSuccess: () => {
			toast.success("Exported successfully");
		},
		onError: () => {
			toast.error("Failed to export");
		},
		networkMode: "always",
	});

	const filterableNumberFields = group.fields.filter(
		(f) => f.type === FieldType.NUMBER,
	);
	const filterableStringFields = group.fields.filter(
		(f) => f.type === FieldType.TEXT,
	);

	const numberOperators = ["=", "!=", ">", "<"];
	const stringOperators = ["=", "!="];

	useEffect(() => {
		const applicableFilters = filters.filter(
			(f) => f.value !== null && f.value !== "" && f.fieldId !== "",
		);

		const equal =
			JSON.stringify(params.filters) === JSON.stringify(applicableFilters);
		if (equal) {
			return;
		}
		console.log("setting filters", applicableFilters);
		setParams({ ...params, filters: applicableFilters });
	}, [filters, params, setParams]);

	const sortOptions: SearchSort[] = filterableNumberFields.reduce(
		(acc, curr: Field) => {
			return [
				...acc,
				{ fieldId: curr.id, order: -1 },
				{ fieldId: curr.id, order: 1 },
			] as SearchSort[];
		},
		[] as SearchSort[],
	);

	return (
		<div>
			<div className="flex items-end gap-4 flex-wrap mb-2">
				<TextInput
					labelClassName="label"
					className="input input-sm input-bordered flex items-center gap-2 w-56"
					wrapperClassName="form-control"
					label="Search"
					name="q"
					id="q"
					value={q}
					onChange={(e) => {
						setQ(e.target.value);
						setParams({ ...params, q: e.target.value });
					}}
					required
				/>
				<TextInput
					labelClassName="label"
					className="input input-sm input-bordered flex items-center gap-2 w-24"
					wrapperClassName="form-control"
					label="Limit"
					name="limit"
					id="limit"
					type="number"
					value={limit}
					min={10}
					onChange={(e) => {
						setLimit(Number.parseInt(e.target.value));
						setParams({ ...params, limit: Number.parseInt(e.target.value) });
					}}
					required
				/>
				<div className="form-control w-40">
					<label className="label">
						<span className="label-text">Sorting</span>
					</label>
					<select
						className="select select-bordered select-sm w-full"
						value={`${params.sort.fieldId}:${params.sort.order}`}
						onChange={(e) => {
							const [fieldId, order] = e.target.value.split(":");
							console.log({ fieldId, order });
							setParams({
								...params,
								sort: {
									fieldId,
									order: Number.parseInt(order) as SearchSort["order"],
								},
							});
						}}
					>
						<option value=":1">Sort By</option>
						{sortOptions.map((sort) => (
							<option
								key={sort.fieldId + sort.order}
								value={`${sort.fieldId}:${sort.order}`}
								selected={
									sort.fieldId === params.sort.fieldId &&
									sort.order === params.sort.order
								}
							>
								{group.fields.find((f) => f.id === sort.fieldId)?.name}:{" "}
								{sort.order === 1 ? "ascending" : "descending"}
							</option>
						))}
					</select>
				</div>
				<button
					className="btn btn-primary btn-outline btn-sm"
					onClick={() => {
						setShowExportOptionsModal({
							isOpen: true,
							onConfirm: (name, type) => {
								exportMutation.mutate({ name, type });
							},
						});
					}}
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						strokeWidth={1.5}
						stroke="currentColor"
						className="size-5 mt-[-3px]"
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M9 8.25H7.5a2.25 2.25 0 0 0-2.25 2.25v9a2.25 2.25 0 0 0 2.25 2.25h9a2.25 2.25 0 0 0 2.25-2.25v-9a2.25 2.25 0 0 0-2.25-2.25H15m0-3-3-3m0 0-3 3m3-3V15"
						/>
					</svg>
					Export Results
				</button>
				<button
					className="btn btn-success btn-sm btn-outline"
					onClick={() => {
						setFilters([
							...filters,
							{ fieldId: "", value: null, operator: "=" },
						]);
					}}
				>
					+ Filter
				</button>
			</div>
			{!!filters.length && (
				<div className="flex flex-wrap gap-3 pb-4 pt-2">
					{filters.map((filter, index) => {
						const fieldType = group.fields.find(
							(f) => f.id === filter.fieldId,
						)?.type;
						return (
							<div
								key={index}
								onMouseEnter={() => setHoveredFilter(index)}
								onMouseLeave={() => setHoveredFilter(null)}
								className="flex items-center flex-wrap border-2 rounded-lg border-base-content border-opacity-25 shadow-sm shrink-0 relative"
							>
								{hoveredFilter === index && (
									<button
										className="absolute btn-circle btn btn-xs btn-error"
										onClick={() => {
											setFilters((prev) => prev.filter((_, i) => i !== index));
										}}
										style={{
											top: "-0.5rem",
											right: "-0.5rem",
										}}
									>
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											strokeWidth={1.5}
											stroke="currentColor"
											className="size-3"
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="M6 18 18 6M6 6l12 12"
											/>
										</svg>
									</button>
								)}
								<select
									className="select select-sm  h-9 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0 border-r-2 border-base-content border-opacity-25 rounded-r-none"
									value={filter.fieldId}
									onChange={(e) => {
										const fieldId = e.target.value;
										setFilters((prev) =>
											prev.map((f, i) =>
												index === i
													? ({
															value: null,
															operator: "=",
															fieldId,
														} as SearchFilter)
													: f,
											),
										);
									}}
								>
									<option value="">Select Field</option>
									{[...filterableNumberFields, ...filterableStringFields].map(
										(field) => (
											<option key={field.id} value={field.id}>
												{field.name}
											</option>
										),
									)}
								</select>
								<select
									className="select select-sm w-16 h-9 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0 border-r-2 border-base-content border-opacity-25 rounded-r-none"
									value={filter.operator}
									onChange={(e) => {
										const operator = e.target.value as SearchFilter["operator"];
										setFilters((prev) =>
											prev.map((f, i) =>
												i === index ? { ...f, operator } : f,
											),
										);
									}}
								>
									{fieldType === FieldType.NUMBER ? (
										<>
											{numberOperators.map((op) => (
												<option key={op} value={op}>
													{op}
												</option>
											))}
										</>
									) : (
										<>
											{stringOperators.map((op) => (
												<option key={op} value={op}>
													{op}
												</option>
											))}
										</>
									)}
								</select>
								<input
									className="input input-sm h-9 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0"
									type={fieldType === FieldType.NUMBER ? "number" : "text"}
									value={filter.value ?? ""}
									onChange={(e) => {
										let value: string | number = e.target.value;
										if (fieldType === FieldType.NUMBER) {
											console.log("here");
											value = Number.parseFloat(value);
										}
										setFilters((prev) =>
											prev.map((f, i) => (i === index ? { ...f, value } : f)),
										);
									}}
								/>
								{/* <button
                  onClick={() => {
                    setFilters((prev) => prev.filter((_, i) => i !== index));
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
                      d="M6 18 18 6M6 6l12 12"
                    />
                  </svg>
                </button> */}
							</div>
						);
					})}
				</div>
			)}
			{showExportOptionsModal.isOpen && (
				<ConfigureExportModal
					defaultName={`${group.name}-${new Date().toLocaleDateString()}`}
					isOpen={showExportOptionsModal.isOpen}
					onClose={() =>
						setShowExportOptionsModal({ isOpen: false, onConfirm: null })
					}
					onConfirm={(name, type) => {
						if (showExportOptionsModal.onConfirm) {
							showExportOptionsModal.onConfirm(name, type);
						}
					}}
				/>
			)}
		</div>
	);
};
