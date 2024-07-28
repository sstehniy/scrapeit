/** eslint-disable unused-imports/no-unused-vars */
/** eslint-disable @typescript-eslint/no-explicit-any */
import axios from "axios";
import type React from "react";
import { type FC, useCallback, useEffect, useState } from "react";
import { JsonView, darkStyles } from "react-json-view-lite";
import "react-json-view-lite/dist/index.css";
import { useMutation } from "@tanstack/react-query";
import { toast } from "react-toastify";
import { v4 } from "uuid";
import {
	type Endpoint,
	type Field,
	FieldSelector,
	PaginationConfig,
	type ScrapeGroup,
	type ScrapeResultTest,
	ScrapeStatus,
	ScrapeType,
	SelectorStatus,
} from "../../types";
import { FloatingAIChat } from "../FloatingAIChat";
import { Button } from "../ui/Button";
import { Modal, type ModalProps } from "../ui/Modal";
import { TextInput } from "../ui/TextInput";
type ConfigureGroupEndpointProps = Pick<ModalProps, "isOpen" | "onClose"> & {
	onConfirm: (endpoint: Endpoint) => void | Promise<void>;
	editEndpoint?: Endpoint;
	fields: Field[];
};

const isEndpointPaginationConfig = (
	config: unknown,
): config is PaginationConfig => {
	console.log(config);
	const configCandidate = config as Partial<PaginationConfig>;
	return (
		typeof configCandidate.type === "string" &&
		typeof configCandidate.parameter === "string" &&
		(typeof configCandidate.step === "number" ||
			typeof configCandidate.step === "undefined") &&
		(typeof configCandidate.start === "number" ||
			typeof configCandidate.start === "undefined") &&
		(typeof configCandidate.end === "number" ||
			typeof configCandidate.end === "undefined") &&
		(typeof configCandidate.urlRegexToInsert === "string" ||
			typeof configCandidate.urlRegexToInsert === "undefined")
	);
};

const isEndpoint = (ep: unknown): ep is Endpoint => {
	const epCandidate = ep as Partial<Endpoint>;
	console.log(epCandidate);
	console.log(isEndpointPaginationConfig(epCandidate.paginationConfig));
	return (
		typeof epCandidate.name === "string" &&
		typeof epCandidate.url === "string" &&
		typeof epCandidate.mainElementSelector === "string" &&
		typeof epCandidate.withDetailedView === "boolean" &&
		typeof epCandidate.detailedViewTriggerSelector === "string" &&
		typeof epCandidate.detailedViewMainElementSelector === "string" &&
		typeof epCandidate.interval === "string" &&
		isEndpointPaginationConfig(epCandidate.paginationConfig)
	);
};

const defaultEndpoint: Endpoint = {
	id: "",
	name: "",
	url: "",
	detailFieldSelectors: [],
	mainElementSelector: "",
	withDetailedView: false,
	detailedViewTriggerSelector: "",
	detailedViewMainElementSelector: "",
	interval: "*/5 * * * *",
	active: false,
	status: ScrapeStatus.IDLE,
	paginationConfig: {
		type: "url_parameter",
		parameter: "",
		end: 0,
		start: 0,
		step: 0,
	},
};

const getScrapeType = (ep: Endpoint) => {
	const mainElementsSelector = ep.mainElementSelector.trim();
	const withDetailedView = ep.withDetailedView;
	const detailedViewTriggerSelector = ep.detailedViewTriggerSelector.trim();
	const detailedViewMainElementSelector =
		ep.detailedViewMainElementSelector.trim();

	if (!withDetailedView && mainElementsSelector) {
		return ScrapeType.PREVIEWS;
	}

	if (
		withDetailedView &&
		detailedViewTriggerSelector === "" &&
		detailedViewMainElementSelector !== ""
	) {
		return ScrapeType.PURE_DETAILS;
	}

	if (
		withDetailedView &&
		mainElementsSelector &&
		detailedViewTriggerSelector &&
		detailedViewMainElementSelector
	) {
		return ScrapeType.PREVIEWS_WITH_DETAILS;
	}

	return ScrapeType.PREVIEWS;
};

const scrapeTypes = [
	{
		label: "Previews",
		value: ScrapeType.PREVIEWS,
	},
	{
		label: "Previews with Details",
		value: ScrapeType.PREVIEWS_WITH_DETAILS,
	},
	{
		label: "Pure Details",
		value: ScrapeType.PURE_DETAILS,
	},
] as const;

const FirstStepContent: FC<{
	endpoint: Endpoint;
	setEndpoint: React.Dispatch<React.SetStateAction<Endpoint>>;
	firstStepErrors: { [key: string]: string };
	setFirstStepErrors: React.Dispatch<
		React.SetStateAction<{ [key: string]: string }>
	>;
	handleTestGettingElement: () => void;
	testElementLoading: boolean;
	testElementError: string | null;
	testElementResult: string | null;
	validateFirstStep: (ep: Endpoint) => boolean;
	defaultScrapeType: ScrapeType;
}> = ({
	endpoint,
	setEndpoint,
	firstStepErrors,
	handleTestGettingElement,
	testElementLoading,
	testElementError,
	testElementResult,
	validateFirstStep,
	defaultScrapeType,
}) => {
	const [scrapeType, setScrapeType] = useState<ScrapeType>(ScrapeType.PREVIEWS);

	useEffect(() => {
		setScrapeType(defaultScrapeType);
	}, [defaultScrapeType]);

	return (
		<div className="w-[450px]">
			<label className="form-control w-full">
				<div className="label">
					<span className="label-text">Import Endpoint</span>
				</div>
				<input
					type="file"
					className="file-input file-input-bordered "
					accept=".json"
					onChange={(e) => {
						const file = e.target.files?.[0];
						if (!file) return;
						const reader = new FileReader();
						reader.onload = (e) => {
							const content = e.target?.result;
							if (typeof content !== "string") return;
							try {
								const endpoint = JSON.parse(content);
								console.log(endpoint);
								if (isEndpoint(endpoint)) {
									const newEndpoint = {
										...endpoint,
										name: endpoint.name + " (imported)",
										url: endpoint.url,
										mainElementSelector: endpoint.mainElementSelector,
										detailFieldSelectors: endpoint.detailFieldSelectors,
										withDetailedView: endpoint.withDetailedView,
										detailedViewTriggerSelector:
											endpoint.detailedViewTriggerSelector,
										detailedViewMainElementSelector:
											endpoint.detailedViewMainElementSelector,
									};
									setEndpoint(newEndpoint);
									validateFirstStep(newEndpoint);
								} else {
									toast.error("Failed to parse endpoint config");
								}
							} catch (error) {
								console.error(error);
								toast.error("Failed to parse endpoint config");
							}
						};
						reader.readAsText(file);
					}}
				/>
			</label>
			<TextInput
				labelClassName="label"
				className="input input-bordered flex items-center gap-2"
				wrapperClassName="form-control mb-4"
				label="Name"
				name="name"
				id="name"
				value={endpoint.name}
				onChange={(e) => {
					const newEndpoint = { ...endpoint, name: e.target.value };
					validateFirstStep(newEndpoint);
					setEndpoint(newEndpoint);
				}}
				required
				error={firstStepErrors.name}
			/>

			<TextInput
				labelClassName="label"
				className="input input-bordered flex items-center gap-2"
				wrapperClassName="form-control mb-4"
				label="Endpoint URL"
				name="endpoint_name"
				id="endpoint_name"
				value={endpoint.url}
				onChange={(e) => {
					const newEndpoint = { ...endpoint, url: e.target.value };
					validateFirstStep(newEndpoint);
					setEndpoint(newEndpoint);
				}}
				required
				error={firstStepErrors.url}
			/>

			<div role="tablist" className="tabs tabs-boxed px-0">
				{/* <a role="tab" className="tab">
					Tab 1
				</a>
				<a role="tab" className="tab tab-active">
					Tab 2
				</a>
				<a role="tab" className="tab">
					Tab 3
				</a> */}
				{scrapeTypes.map((type) => (
					<button
						key={type.value}
						className={`tab ${scrapeType === type.value ? "tab-active" : ""}`}
						onClick={() => {
							setScrapeType(type.value);
							let newEndpoint = { ...endpoint };

							switch (type.value) {
								case ScrapeType.PREVIEWS:
									newEndpoint = {
										...endpoint,
										withDetailedView: false,
									};
									break;
								case ScrapeType.PREVIEWS_WITH_DETAILS:
									newEndpoint = {
										...endpoint,
										withDetailedView: true,
									};
									break;
								case ScrapeType.PURE_DETAILS:
									newEndpoint = {
										...endpoint,
										withDetailedView: true,
									};
									break;
							}
							setScrapeType(type.value);
							validateFirstStep(newEndpoint);
							setEndpoint(newEndpoint);
						}}
					>
						{type.label}
					</button>
				))}
			</div>

			{scrapeType !== ScrapeType.PURE_DETAILS && (
				<TextInput
					labelClassName="label"
					className="input input-bordered flex items-center gap-2"
					wrapperClassName="form-control mb-4"
					label="Main Element Selector"
					name="main_element_selector"
					id="main_element_selector"
					value={endpoint.mainElementSelector}
					onChange={(e) => {
						const newEndpoint = {
							...endpoint,
							mainElementSelector: e.target.value,
						};
						validateFirstStep(newEndpoint);
						setEndpoint(newEndpoint);
					}}
					required
					error={firstStepErrors.mainElementSelector}
				/>
			)}
			{/* <div className="form-control">
				<label className="label cursor-pointer">
					<span className="label-text">Detailed View Scrape</span>
					<input
						type="checkbox"
						className="toggle"
						checked={endpoint.withDetailedView}
						onChange={() => {
							const newEndpoint = {
								...endpoint,
								withDetailedView: !endpoint.withDetailedView,
							};
							validateFirstStep(newEndpoint);
							setEndpoint(newEndpoint);
						}}
					/>
				</label>
			</div> */}

			{scrapeType === ScrapeType.PREVIEWS_WITH_DETAILS && (
				<TextInput
					labelClassName="label"
					className="input input-bordered flex items-center gap-2"
					wrapperClassName="form-control mb-4"
					label="Detailed View Trigger Selector"
					name="detailed_view_trigger_selector"
					id="detailed_view_trigger_selector"
					value={endpoint.detailedViewTriggerSelector}
					onChange={(e) => {
						const newEndpoint = {
							...endpoint,
							detailedViewTriggerSelector: e.target.value,
						};
						validateFirstStep(newEndpoint);
						setEndpoint(newEndpoint);
					}}
					required
					error={firstStepErrors.detailedViewTriggerSelector}
				/>
			)}
			{scrapeType !== ScrapeType.PREVIEWS && (
				<TextInput
					labelClassName="label"
					className="input input-bordered flex items-center gap-2"
					wrapperClassName="form-control mb-4"
					label="Detailed View Main Element Selector"
					name="detailed_view_main_element_selector"
					id="detailed_view_main_element_selector"
					value={endpoint.detailedViewMainElementSelector}
					onChange={(e) => {
						const newEndpoint = {
							...endpoint,
							detailedViewMainElementSelector: e.target.value,
						};
						validateFirstStep(newEndpoint);
						setEndpoint(newEndpoint);
					}}
					required
					error={firstStepErrors.detailedViewMainElementSelector}
				/>
			)}

			<Button
				className="btn btn-primary btn-sm"
				onClick={handleTestGettingElement}
				disabled={
					testElementLoading || Object.values(firstStepErrors).length > 0
				}
			>
				Test Getting Element
			</Button>

			{testElementLoading && (
				<div className="w-full flex justify-center mb-3">
					<span className="loading loading-spinner loading-lg" />
				</div>
			)}

			{testElementError && (
				<p className="text-red-500 text-xs italic">{testElementError}</p>
			)}

			{testElementResult && (
				<div className="bg-gray-800 px-2 rounded mt-2">
					<pre
						className="text-xs text-gray-300 rounded"
						style={{
							height: "200px",
							overflow: "auto",
							whiteSpace: "pre-wrap",
						}}
					>
						{testElementResult}
					</pre>
				</div>
			)}
		</div>
	);
};

type Remark = {
	fieldId: string;
	remark: string;
};

const SecondStepContent: FC<{
	endpoint: Endpoint;
	setEndpoint: React.Dispatch<React.SetStateAction<Endpoint>>;
	fields: Field[];
	handleExtractSelectorsForAllFields: (remarks: Remark[]) => void;
	handleExtractSelectorForField: (field: Field, remark: string) => void;
	totalCost: number;
	fieldsWithLoadingSelectors: string[];
	validateSecondStep: (ep: Endpoint) => boolean;
	loadingSampleData: boolean;
	handleTestScrape: (ep: Endpoint) => void;
	sampleData: ScrapeResultTest[];
}> = ({
	endpoint,
	setEndpoint,
	fields,
	handleExtractSelectorsForAllFields,
	handleExtractSelectorForField,
	fieldsWithLoadingSelectors,
	loadingSampleData,
	sampleData,
	totalCost,
	handleTestScrape,
	validateSecondStep,
}) => {
	const [remarks, setRemarks] = useState<Remark[]>(
		fields.map((field) => ({ fieldId: field.id, remark: "" })),
	);
	const intervals = [
		{
			label: "Every minute",
			value: "* * * * *",
		},
		{
			label: "Every 5 minutes",
			value: "*/5 * * * *",
		},
		{
			label: "Every 15 minutes",
			value: "*/15 * * * *",
		},
		{
			label: "Every 30 minutes",
			value: "*/30 * * * *",
		},
		{
			label: "Every hour",
			value: "0 * * * *",
		},
		{
			label: "Every 6 hours",
			value: "0 */6 * * *",
		},
		{
			label: "Every 12 hours",
			value: "0 */12 * * *",
		},
		{
			label: "Every day",
			value: "0 0 * * *",
		},
	];
	return (
		<div className="w-full mx-auto px-4">
			<div className="space-y-4">
				<div>
					<label className="label cursor-pointer w-full sm:w-[600px]">
						<span className="label-text text-lg">Name:</span>{" "}
						<strong className="text-xl">{endpoint.name}</strong>
					</label>
				</div>
				<div className="flex flex-col sm:flex-row items-start sm:items-center gap-2 sm:gap-5 mb-2 w-full sm:w-[600px] justify-between">
					<label className="label cursor-pointer">
						<span className="label-text text-lg">URL:</span>
					</label>
					<label className="input input-bordered input-sm flex items-center gap-1 flex-1 w-full">
						<input
							type="text"
							readOnly
							className="grow"
							defaultValue={endpoint.url}
						/>
						<button
							onClick={() => {
								navigator.clipboard.writeText(endpoint.url).then(() => {
									toast.success(
										"Endpoint URL successfully copied to clipboard",
										{
											autoClose: 1000,
										},
									);
								});
							}}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								strokeWidth={1}
								stroke="currentColor"
								className="size-5"
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"
								/>
							</svg>
						</button>
					</label>
				</div>
				<div>
					<label className="label cursor-pointer w-full sm:w-[600px] mb-2">
						<span className="label-text text-lg">Main Element Selector:</span>
						<strong>{endpoint.mainElementSelector}</strong>
					</label>
				</div>
				<div className="form-control w-full sm:w-[600px] mb-2">
					<label className="label cursor-pointer">
						<span className="label-text text-lg">Scrape in background</span>
						<input
							type="checkbox"
							className="toggle"
							checked={endpoint.active}
							onChange={() => {
								const newEndpoint = {
									...endpoint,
									active: !endpoint.active,
								};
								validateSecondStep(newEndpoint);
								setEndpoint(newEndpoint);
							}}
						/>
					</label>
				</div>
				{endpoint.active && (
					<>
						<label className="label">Scrape Interval</label>
						<select
							className="select select-bordered w-full mb-4"
							value={endpoint.interval}
							onChange={(e) => {
								const newEndpoint = {
									...endpoint,
									interval: e.target.value,
								};
								validateSecondStep(newEndpoint);
								setEndpoint(newEndpoint);
							}}
						>
							{intervals.map((interval) => (
								<option key={interval.value} value={interval.value}>
									{interval.label}
								</option>
							))}
						</select>
					</>
				)}
				{!(
					endpoint.withDetailedView && !endpoint.detailedViewTriggerSelector
				) && (
					<>
						<label className="label">Pagination Config</label>
						<select
							className="select select-bordered w-full mb-4"
							value={endpoint.paginationConfig.type}
							onChange={(e) => {
								const newEndpoint = {
									...endpoint,
									paginationConfig: {
										...endpoint.paginationConfig,
										type: e.target.value as "url_parameter",
									},
								};
								validateSecondStep(newEndpoint);
								setEndpoint(newEndpoint);
							}}
						>
							<option value="url_parameter" selected>
								URL Parameter
							</option>
							<option value="url_path">URL Path</option>
						</select>
						<label className="label">Parameter</label>
						<input
							type="text"
							name="name"
							id="name"
							className="input input-bordered flex items-center gap-2 w-full mb-4"
							value={endpoint.paginationConfig.parameter}
							onChange={(e) => {
								const newEndpoint = {
									...endpoint,
									paginationConfig: {
										...endpoint.paginationConfig,
										parameter: e.target.value,
									},
								};
								validateSecondStep(newEndpoint);
								setEndpoint(newEndpoint);
							}}
							required
						/>

						<div className="flex flex-col sm:flex-row gap-5 mb-10 w-full">
							<div className="flex-1">
								<label className="label">Start</label>
								<input
									type="number"
									name="name"
									id="name"
									min={0}
									className="input input-bordered flex items-center gap-2 w-full"
									value={endpoint.paginationConfig.start}
									onChange={(e) => {
										const newEndpoint = {
											...endpoint,
											paginationConfig: {
												...endpoint.paginationConfig,
												start: Number.parseInt(e.target.value),
											},
										};
										validateSecondStep(newEndpoint);
										setEndpoint(newEndpoint);
									}}
									required
								/>
							</div>
							<div className="flex-1">
								<label className="label">End</label>
								<input
									type="number"
									name="name"
									id="name"
									min={0}
									className="input input-bordered flex items-center gap-2 w-full"
									value={endpoint.paginationConfig.end}
									onChange={(e) => {
										const newEndpoint = {
											...endpoint,
											paginationConfig: {
												...endpoint.paginationConfig,
												end: Number.parseInt(e.target.value),
											},
										};
										validateSecondStep(newEndpoint);
										setEndpoint(newEndpoint);
									}}
									required
								/>
							</div>
							<div className="flex-1">
								<label className="label">Step</label>
								<input
									type="number"
									name="name"
									id="name"
									min={1}
									step={1}
									className="input input-bordered flex items-center gap-2 w-full"
									value={endpoint.paginationConfig.step}
									onChange={(e) => {
										const newEndpoint = {
											...endpoint,
											paginationConfig: {
												...endpoint.paginationConfig,
												step: Number.parseInt(e.target.value),
											},
										};
										validateSecondStep(newEndpoint);
										setEndpoint(newEndpoint);
									}}
									required
								/>
							</div>
							{endpoint.paginationConfig.type === "url_path" && (
								<div className="flex-1">
									<TextInput
										name="urlRegexToInsert"
										id="urlRegexToInsert"
										label="URL Regex To Insert"
										labelClassName="label"
										className="input input-bordered flex items-center gap-2 w-full"
										value={endpoint.paginationConfig.urlRegexToInsert}
										onChange={(e) => {
											const newEndpoint = {
												...endpoint,
												paginationConfig: {
													...endpoint.paginationConfig,
													urlRegexToInsert: e.target.value,
												},
											};
											validateSecondStep(newEndpoint);
											setEndpoint(newEndpoint);
										}}
									/>
								</div>
							)}
						</div>
					</>
				)}

				<div className="w-full flex justify-end">
					<div className="flex flex-col items-end mb-2">
						<Button
							className="btn btn-primary btn-sm mb-1"
							onClick={() => {
								handleExtractSelectorsForAllFields(remarks);
							}}
							disabled={
								loadingSampleData || !!fieldsWithLoadingSelectors.length
							}
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
									d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z"
								/>
							</svg>
							Smart Extract Selectors{" "}
						</Button>
						<div className="text-base-content text-xs text-opacity-55">
							{totalCost > 0 && `$${totalCost.toFixed(4)}`}
						</div>
					</div>
				</div>

				<div className="w-full overflow-x-auto">
					{fields.map((field, idx) => (
						<div key={field.id} className="mb-4">
							<div className="flex flex-col sm:flex-row gap-3">
								<button
									className="btn btn-square btn-sm btn-outline border-0 self-start sm:self-center"
									onClick={() => {
										const newEndpoint = {
											...endpoint,
											detailFieldSelectors: endpoint.detailFieldSelectors.map(
												(selector) =>
													selector.fieldId === field.id
														? {
																...selector,
																lockedForEdit: !selector.lockedForEdit,
															}
														: selector,
											),
										};
										setEndpoint(newEndpoint);
									}}
								>
									{endpoint.detailFieldSelectors.find(
										(selector) => selector.fieldId === field.id,
									)?.lockedForEdit ? (
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											strokeWidth={1.5}
											stroke="currentColor"
											className="size-5 text-error"
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
											/>
										</svg>
									) : (
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
												d="M13.5 10.5V6.75a4.5 4.5 0 1 1 9 0v3.75M3.75 21.75h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H3.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
											/>
										</svg>
									)}
								</button>
								<div
									className="flex flex-col sm:flex-row gap-3 w-full"
									style={{
										opacity: endpoint.detailFieldSelectors.find(
											(selector) => selector.fieldId === field.id,
										)?.lockedForEdit
											? 0.5
											: 1,
										pointerEvents: endpoint.detailFieldSelectors.find(
											(selector) => selector.fieldId === field.id,
										)?.lockedForEdit
											? "none"
											: "auto",
									}}
								>
									<div className="w-full sm:w-36">
										<TextInput
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full"
											wrapperClassName="form-control mb-2 sm:mb-0"
											label={idx > 0 ? "" : "Field Name"}
											readOnly
											disabled
											value={field.name}
										/>
									</div>
									<div className="w-full sm:w-36">
										<TextInput
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full"
											wrapperClassName="form-control mb-2 sm:mb-0"
											label={idx > 0 ? "" : "Field Key"}
											readOnly
											disabled
											value={field.key}
										/>
									</div>
									<div className="flex-1">
										<TextInput
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full mb-1"
											wrapperClassName="form-control"
											label={idx > 0 ? "" : "Selector"}
											value={
												endpoint.detailFieldSelectors.find(
													(selector) => selector.fieldId === field.id,
												)?.selector
											}
											onChange={(e) => {
												const newEndpoint = {
													...endpoint,
													detailFieldSelectors:
														endpoint.detailFieldSelectors.map((selector) =>
															selector.fieldId === field.id
																? { ...selector, selector: e.target.value }
																: selector,
														),
												};
												setEndpoint(newEndpoint);
												validateSecondStep(newEndpoint);
											}}
											required
										/>
										{endpoint.detailFieldSelectors.find(
											(selector) =>
												selector.fieldId === field.id &&
												selector.selector === "" &&
												selector.selectorStatus === SelectorStatus.NEW,
										) && (
											<div className="text-yellow-500 text-xs italic">
												Selector not extracted
											</div>
										)}
										{endpoint.detailFieldSelectors.find(
											(selector) => selector.fieldId === field.id,
										)?.selectorStatus === SelectorStatus.NEEDS_UPDATE && (
											<div className="text-yellow-500 text-xs italic">
												Selector may need to be updated
											</div>
										)}
									</div>
									<div className="w-full sm:w-28">
										<TextInput
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full"
											wrapperClassName="form-control mb-2 sm:mb-0"
											label={idx > 0 ? "" : "Attribute"}
											value={
												endpoint.detailFieldSelectors.find(
													(selector) => selector.fieldId === field.id,
												)?.attributeToGet
											}
											onChange={(e) => {
												const newEndpoint = {
													...endpoint,
													detailFieldSelectors:
														endpoint.detailFieldSelectors.map((selector) =>
															selector.fieldId === field.id
																? {
																		...selector,
																		attributeToGet: e.target.value,
																	}
																: selector,
														),
												};
												setEndpoint(newEndpoint);
												validateSecondStep(newEndpoint);
											}}
										/>
									</div>
									<div className="w-full sm:w-36">
										<TextInput
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full"
											wrapperClassName="form-control mb-2 sm:mb-0"
											label={idx > 0 ? "" : "Regex"}
											value={
												endpoint.detailFieldSelectors.find(
													(selector) => selector.fieldId === field.id,
												)?.regex
											}
											onChange={(e) => {
												const newEndpoint = {
													...endpoint,
													detailFieldSelectors:
														endpoint.detailFieldSelectors.map((selector) =>
															selector.fieldId === field.id
																? { ...selector, regex: e.target.value }
																: selector,
														),
												};
												setEndpoint(newEndpoint);
												validateSecondStep(newEndpoint);
											}}
										/>
									</div>
									<div className="w-full sm:w-24">
										<TextInput
											type="number"
											min={0}
											step={1}
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full"
											wrapperClassName="form-control mb-2 sm:mb-0"
											label={idx > 0 ? "" : "Regex Match"}
											value={
												endpoint.detailFieldSelectors.find(
													(selector) => selector.fieldId === field.id,
												)?.regexMatchIndexToUse
											}
											onChange={(e) => {
												setEndpoint((prev) => ({
													...prev,
													detailFieldSelectors: prev.detailFieldSelectors.map(
														(selector) =>
															selector.fieldId === field.id
																? {
																		...selector,
																		regexMatchIndexToUse: +e.target.value,
																	}
																: selector,
													),
												}));
											}}
										/>
									</div>
									<div className="flex-1">
										<TextInput
											labelClassName="label"
											className="input input-bordered flex items-center gap-2 w-full"
											wrapperClassName="form-control mb-2 sm:mb-0"
											label={idx > 0 ? "" : "Remarks"}
											tooltip={
												"This data is not saved, only helpful for smart extract)"
											}
											value={
												remarks.find((r) => r.fieldId === field.id)!.remark
											}
											onChange={(e) => {
												setRemarks((prev) =>
													prev.map((pr) =>
														pr.fieldId === field.id
															? { ...pr, remark: e.target.value }
															: pr,
													),
												);
											}}
										/>
									</div>
								</div>
								<button
									className="btn btn-square btn-sm btn-outline border-0 self-start sm:self-center"
									onClick={() => {
										handleExtractSelectorForField(
											field,
											remarks.find((r) => r.fieldId === field.id)!.remark,
										);
									}}
									disabled={
										!!fieldsWithLoadingSelectors.length ||
										endpoint.detailFieldSelectors.find(
											(selector) => selector.fieldId === field.id,
										)?.lockedForEdit ||
										loadingSampleData
									}
									style={{
										userSelect: fieldsWithLoadingSelectors.includes(field.id)
											? "none"
											: "auto",
										cursor: fieldsWithLoadingSelectors.includes(field.id)
											? "not-allowed"
											: "pointer",
									}}
								>
									{!fieldsWithLoadingSelectors.includes(field.id) ? (
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
												d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z"
											/>
										</svg>
									) : (
										<span className="loading loading-spinner loading-md"></span>
									)}
								</button>
							</div>
						</div>
					))}
				</div>
				<div className="w-full flex justify-end">
					<Button
						className="btn btn-outline btn-primary btn-sm mb-2"
						onClick={() => {
							handleTestScrape(endpoint);
						}}
						disabled={loadingSampleData || !!fieldsWithLoadingSelectors.length}
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
						Test Scrape
					</Button>
				</div>
				{loadingSampleData && (
					<div className="w-full flex justify-center mb-3">
						<span className="loading loading-spinner loading-lg"></span>
					</div>
				)}
				{sampleData?.length > 0 && (
					<div className="w-full">
						<div
							className="mb-5 overflow-x-auto"
							style={{
								maxHeight: 500,
								overflowY: "auto",
							}}
						>
							<table className="table table-compact w-full">
								<thead className="bg-base-300 sticky top-0">
									<tr>
										{fields.map((field) => (
											<th
												key={field.id}
												className="px-4 py-3 text-left text-xs font-medium text-base-content uppercase tracking-wider cursor-pointer hover:bg-base-200 transition whitespace-nowrap"
											>
												{field.name}
											</th>
										))}
									</tr>
								</thead>
								<tbody>
									{sampleData.map((row) => (
										<tr key={row.id} className="hover:bg-base-200 transition">
											{fields.map((field) => (
												<td key={field.id} className="text-sm">
													<div
														className="px-4"
														style={{
															overflow: "hidden",
															textOverflow: "ellipsis",
															maxHeight: 95,
															maxWidth: 300,
															display: "-webkit-box",
															WebkitLineClamp: 4,
															WebkitBoxOrient: "vertical",
														}}
													>
														{
															row.fields.find((r) => r.fieldId === field.id)
																?.value
														}
													</div>
												</td>
											))}
										</tr>
									))}
								</tbody>
							</table>
						</div>
						<div
							className="w-full"
							style={{
								maxHeight: 400,
								overflow: "auto",
							}}
						>
							<JsonView
								data={sampleData}
								clickToExpandNode
								shouldExpandNode={() => true}
								style={darkStyles}
							/>
						</div>
					</div>
				)}
			</div>
		</div>
	);
};

export const ConfigureGroupEndpoint: FC<ConfigureGroupEndpointProps> = ({
	onConfirm,
	isOpen,
	onClose,
	editEndpoint,
	fields,
}) => {
	const [endpoint, setEndpoint] = useState<Endpoint>(defaultEndpoint);
	const [currentStep, setCurrentStep] = useState(0);
	const [firstStepErrors, setFirstStepErrors] = useState<{
		[key: string]: string;
	}>({});
	const [secondStepErrors, setSecondStepErrors] = useState<{
		[key: string]: string;
	}>({});
	const [testElementLoading, setTestElementLoading] = useState(false);
	const [testElementError, setTestElementError] = useState<string | null>(null);
	const [testElementResult, setTestElementResult] = useState<string | null>(
		null,
	);
	const [fieldsWithLoadingSelectors, setFieldsWithLoadingSelectors] = useState<
		string[]
	>([]);
	const [loadingSampleData, setLoadingSampleData] = useState(false);
	const [sampleData, setSampleData] = useState<ScrapeResultTest[]>([]);
	const [totalCost, setTotalCost] = useState(0);

	useEffect(() => {
		if (editEndpoint) {
			setEndpoint(editEndpoint);
		} else {
			const prefilledDetailSelectors: Endpoint["detailFieldSelectors"] =
				fields.map((field) => ({
					id: v4(),
					fieldId: field.id,
					attributeToGet: "",
					selector: "",
					regex: "",
					regexMatchIndexToUse: 0,
					lockedForEdit: false,
					selectorStatus: SelectorStatus.NEW,
				}));
			setEndpoint({
				...defaultEndpoint,
				id: v4(),
				detailFieldSelectors: prefilledDetailSelectors,
			});
		}
	}, [editEndpoint, fields]);

	const testGettingElementMutation = useMutation({
		mutationFn: ({ endpoint }: { endpoint: Endpoint }) =>
			axios
				.post("/api/selectors/test", {
					endpoint,
				})
				.then((resp) => resp.data as { html: string }),
		onSuccess: (data) => setTestElementResult(data.html),
		onMutate: () => {
			setTestElementError(null);
			setTestElementResult(null);
			setTestElementLoading(true);
		},
		onError: (error) => {
			console.error(error);
			setTestElementError("Error while testing element");
		},
		onSettled: () => setTestElementLoading(false),
	});

	const handleTestGettingElement = useCallback(() => {
		if (!endpoint.url.trim() || !endpoint.mainElementSelector.trim()) {
			setTestElementError("Please enter a URL");
			return;
		}
		testGettingElementMutation.mutate({ endpoint });
	}, [testGettingElementMutation, endpoint]);

	const handleTestScrapeMutation = useMutation({
		mutationFn: ({ group }: { group: ScrapeGroup }) =>
			axios
				.post("/api/scrape/endpoint-test", {
					group,
				})
				.then((resp) => resp.data as ScrapeResultTest[]),
		onSuccess: (data) => setSampleData(data),
		onMutate: () => {
			setSampleData([]);
			setLoadingSampleData(true);
		},
		onError: (error) => {
			console.error(error);
			toast.error("Failed to test scrape");
		},
		onSettled: () => {
			setLoadingSampleData(false);
		},
	});

	const handleTestScrape = useCallback(
		(ep: Endpoint) => {
			const endpointCopy: Endpoint = {
				...ep,
				paginationConfig: {
					...ep.paginationConfig,
					end: ep.paginationConfig.start,
				},
			};

			const group: ScrapeGroup = {
				id: "1",
				name: "Test Group",
				fields,
				endpoints: [endpointCopy],
				withThumbnail: false,
				created: new Date().toISOString(),
				updated: new Date().toISOString(),
				isArchived: false,
				versionTag: "",
			};

			handleTestScrapeMutation.mutate({ group });
		},
		[fields, handleTestScrapeMutation],
	);

	const extractSelectorForFieldMutation = useMutation({
		mutationFn: ({ field, remark }: { field: Field; remark: string }) =>
			axios
				.post("/api/selectors/extract", {
					endpoint,
					fieldsToExtractSelectorsFor: [
						{
							key: field.key,
							name: field.name,
							type: field.type,
							remark,
						},
					],
				})
				.then(
					(res) => res.data as { fields: FieldSelector[]; totalCost: number },
				),
		onSuccess: ({ fields, totalCost }, { field }) => {
			console.log("Extracted selectors", fields);
			const newEndpoint: Endpoint = {
				...endpoint,
				detailFieldSelectors: endpoint.detailFieldSelectors.map((selector) =>
					selector.fieldId === field.id
						? {
								...selector,
								regexMatchIndexToUse: fields[0]?.regexMatchIndexToUse,
								selector: fields[0]?.selector || "",
								attributeToGet: fields[0]?.attributeToGet || "",
								regex: fields[0]?.regex || "",
							}
						: selector,
				),
			};
			setEndpoint(newEndpoint);
			setTotalCost((prev) => prev + totalCost);
			setFieldsWithLoadingSelectors((prev) =>
				prev.filter((id) => id !== field.id),
			);
			handleTestScrape(newEndpoint);
		},
		onMutate: ({ field }) => {
			setFieldsWithLoadingSelectors((prev) => [...prev, field.id]);
			console.log("Extracting selector for field", field);
		},
		onError: (error) => {
			setFieldsWithLoadingSelectors([]);
			console.error(error);
			toast.error("Failed to extract selector for field");
		},
	});

	const handleExtractSelectorForField = useCallback(
		(field: Field, remark: string) => {
			if (fieldsWithLoadingSelectors.includes(field.id)) {
				return;
			}

			extractSelectorForFieldMutation.mutate({
				field,
				remark,
			});
		},
		[fieldsWithLoadingSelectors, extractSelectorForFieldMutation],
	);

	const extractSelectorsForAllFieldsMutation = useMutation({
		mutationFn: ({
			toExtract,
			remarks,
		}: { toExtract: Field[]; remarks: Remark[]; groupFields: Field[] }) =>
			axios
				.post("/api/selectors/extract", {
					endpoint,
					fieldsToExtractSelectorsFor: toExtract.map((field) => ({
						key: field.key,
						name: field.name,
						type: field.type,
						remark: remarks.find((r) => r.fieldId === field.id)?.remark,
					})),
				})
				.then(
					(res) => res.data as { fields: FieldSelector[]; totalCost: number },
				),
		onMutate: ({ toExtract }) =>
			setFieldsWithLoadingSelectors(toExtract.map((f) => f.id)),
		onSuccess: ({ fields, totalCost }, { groupFields }) => {
			console.log("Extracted selectors", fields);
			const newEndpoint: Endpoint = {
				...endpoint,
				detailFieldSelectors: endpoint.detailFieldSelectors.map((selector) => {
					const groupField = groupFields.find((f) => f.id === selector.fieldId);
					console.log("Field", groupField, selector);
					const extractedField = fields.find(
						// biome-ignore lint/suspicious/noExplicitAny: <explanation>
						(extractedField: any) => extractedField.field === groupField?.key,
					);
					if (!extractedField) {
						return selector;
					}
					return {
						...selector,
						selector: extractedField.selector,
						attributeToGet: extractedField.attributeToGet,
						regex: extractedField.regex,
						regexMatchIndexToUse: extractedField.regexMatchIndexToUse,
					};
				}),
			};
			setTotalCost((prev) => prev + totalCost);
			setEndpoint(newEndpoint);

			setFieldsWithLoadingSelectors([]);
			handleTestScrape(newEndpoint);
		},
		onError: (error) => {
			console.error(error);
			toast.error("Failed to extract all selectors");
		},
	});

	const handleExtractSelectorsForAllFields = useCallback(
		(remarks: Remark[]) => {
			const toExtract = fields.filter(
				(f) =>
					!endpoint.detailFieldSelectors.find((df) => df.fieldId === f.id)
						?.lockedForEdit,
			);

			if (toExtract.length === 0) {
				return;
			}
			extractSelectorsForAllFieldsMutation.mutate({
				groupFields: fields,
				remarks,
				toExtract,
			});
		},
		[fields, extractSelectorsForAllFieldsMutation, endpoint],
	);

	const validateFirstStep = useCallback((ep: Endpoint) => {
		setFirstStepErrors({});
		const errors: { [key: string]: string } = {};
		if (ep.name.trim() === "") {
			errors.name = "Name is required";
		}
		if (ep.url === "") {
			errors.url = "URL is required";
		}
		// if (ep.mainElementSelector.trim() === "") {
		// 	errors.mainElementSelector = "Main Element Selector is required";
		// }
		const urlRegex = /^(http|https):\/\/[^\s\/$.?#].[^\s]*$/g;
		if (!urlRegex.test(ep.url)) {
			errors.url = "URL is not valid";
		}

		// Validate based on scrape type logic
		const mainElementSelector = ep.mainElementSelector.trim();
		const withDetailedView = ep.withDetailedView;
		const detailedViewTriggerSelector = ep.detailedViewTriggerSelector.trim();
		const detailedViewMainElementSelector =
			ep.detailedViewMainElementSelector.trim();

		// PureDetails case
		if (!withDetailedView) {
			if (mainElementSelector === "") {
				errors.mainElementSelector =
					"Main Element Selector is required for Previews";
			}
		}
		// Configs with Detailed View
		else {
			// Config 2: With Detailed View and Trigger
			if (detailedViewTriggerSelector !== "") {
				if (mainElementSelector === "") {
					errors.mainElementSelector =
						"Main Element Selector is required when Detailed View Trigger is present";
				}
				if (detailedViewMainElementSelector === "") {
					errors.detailedViewMainElementSelector =
						"Detailed View Main Element Selector is required when Detailed View Trigger is present";
				}
			}
			// Config 3: With Detailed View, no Trigger
			else {
				if (detailedViewMainElementSelector === "") {
					errors.detailedViewMainElementSelector =
						"Detailed View Main Element Selector is required when there's no Detailed View Trigger";
				}
				// Note: mainElementSelector is ignored in this case, so we don't validate it
			}
		}
		setFirstStepErrors(errors);
		return Object.entries(errors).length === 0;
	}, []);

	const validateSecondStep = useCallback((ep: Endpoint) => {
		setSecondStepErrors({});
		const errors: { [key: string]: string } = {};
		if (ep.paginationConfig.parameter.trim() === "") {
			errors.parameter = "Parameter is required";
		}
		if (ep.paginationConfig.start < 0) {
			errors.start = "Start is required";
		}
		if (ep.paginationConfig.end === 0) {
			errors.end = "End is required";
		}
		if (ep.paginationConfig.step === 0) {
			errors.step = "Step is required";
		}
		if (
			ep.detailFieldSelectors.some(
				(selector) => selector.selector.trim() === "",
			)
		) {
			errors.selector = "Selector is required for all fields";
		}
		setSecondStepErrors(errors);
		return Object.entries(errors).length === 0;
	}, []);

	const validateField = useCallback(
		(name: "name" | "url" | "mainElementSelector"): boolean => {
			let isValid = true;
			switch (name) {
				case "name":
					if (endpoint.name.trim() === "") {
						isValid = false;
						setFirstStepErrors((prev) => ({
							...prev,
							name: "Name is required",
						}));
					} else {
						setFirstStepErrors((prev) => {
							const newErrors = { ...prev };
							delete newErrors.name;
							return newErrors;
						});
					}
					break;
				case "url":
					if (endpoint.url === "") {
						isValid = false;
						setFirstStepErrors((prev) => ({ ...prev, url: "URL is required" }));
					} else if (
						!/^(http|https):\/\/[^\s\/$.?#].[^\s]*$/g.test(endpoint.url)
					) {
						isValid = false;
						setFirstStepErrors((prev) => ({
							...prev,
							url: "URL is not valid",
						}));
					} else {
						setFirstStepErrors((prev) => {
							const newErrors = { ...prev };
							delete newErrors.url;
							return newErrors;
						});
					}
					break;
				case "mainElementSelector":
					if (endpoint.mainElementSelector.trim() === "") {
						isValid = false;
						setFirstStepErrors((prev) => ({
							...prev,
							mainElementSelector: "Main Element Selector is required",
						}));
					} else {
						setFirstStepErrors((prev) => {
							const newErrors = { ...prev };
							delete newErrors.mainElementSelector;
							return newErrors;
						});
					}
					break;
				default:
					break;
			}
			return isValid;
		},
		[endpoint.mainElementSelector, endpoint.name, endpoint.url],
	);

	return (
		<>
			{isOpen && <FloatingAIChat />}
			{currentStep === 0 ? (
				<Modal
					isOpen={isOpen}
					onClose={onClose}
					closeOnClickOutside={false}
					title={editEndpoint ? "Edit Endpoint" : "Create new Endpoint"}
					actions={[
						{
							label: "Cancel",
							onClick: onClose,
							className: "bg-gray-500 text-white",
						},

						{
							label: "Next",
							onClick: () => {
								if (
									!validateField("name") ||
									!validateField("url") ||
									!validateField("mainElementSelector")
								) {
									return;
								}
								setCurrentStep(currentStep + 1);
							},
							className: "bg-blue-500 text-white",
							disabled: Object.entries(firstStepErrors).length > 0,
						},
					]}
				>
					<FirstStepContent
						endpoint={endpoint}
						defaultScrapeType={getScrapeType(endpoint)}
						setEndpoint={setEndpoint}
						firstStepErrors={firstStepErrors}
						setFirstStepErrors={setFirstStepErrors}
						handleTestGettingElement={handleTestGettingElement}
						testElementLoading={testElementLoading}
						testElementError={testElementError}
						testElementResult={testElementResult}
						validateFirstStep={validateFirstStep}
					/>
				</Modal>
			) : (
				<Modal
					isOpen={isOpen}
					onClose={onClose}
					closeOnClickOutside={false}
					title={editEndpoint ? "Edit Endpoint" : "Create new Endpoint"}
					actions={[
						{
							label: "Cancel",
							onClick: onClose,
							className: "bg-gray-500 text-white",
						},

						{
							label: "Back",
							onClick: () => setCurrentStep(currentStep - 1),
							className: "bg-blue-500 text-white",
						},
						{
							label: editEndpoint ? "Save" : "Create",
							disabled:
								Object.entries(firstStepErrors).length > 0 ||
								Object.entries(secondStepErrors).length > 0 ||
								endpoint.detailFieldSelectors.some(
									(d) =>
										d.selectorStatus === SelectorStatus.NEW &&
										d.selector === "",
								),
							onClick: () => {
								if (
									!validateFirstStep(endpoint) ||
									!validateSecondStep(endpoint)
								) {
									return;
								}
								if (!editEndpoint) {
									endpoint.detailFieldSelectors =
										endpoint.detailFieldSelectors.map((selector) => {
											return {
												...selector,
												selectorStatus: SelectorStatus.OK,
											};
										});
								}
								onConfirm(endpoint);
								onClose();
							},
							className: "bg-green-500 text-white",
						},
					]}
				>
					<SecondStepContent
						endpoint={endpoint}
						setEndpoint={setEndpoint}
						fields={fields}
						handleExtractSelectorsForAllFields={
							handleExtractSelectorsForAllFields
						}
						totalCost={totalCost}
						handleTestScrape={handleTestScrape}
						validateSecondStep={validateSecondStep}
						handleExtractSelectorForField={handleExtractSelectorForField}
						fieldsWithLoadingSelectors={fieldsWithLoadingSelectors}
						loadingSampleData={loadingSampleData}
						sampleData={sampleData}
					/>
				</Modal>
			)}
		</>
	);
};
