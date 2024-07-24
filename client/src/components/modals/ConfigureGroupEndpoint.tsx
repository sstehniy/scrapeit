/** eslint-disable unused-imports/no-unused-vars */
/** eslint-disable @typescript-eslint/no-explicit-any */
import axios from "axios";
import type React from "react";
import { type FC, useCallback, useEffect, useState } from "react";
import { JsonView, darkStyles } from "react-json-view-lite";
import "react-json-view-lite/dist/index.css";
import { toast } from "react-toastify";
import { v4 } from "uuid";
import {
	type Endpoint,
	type Field,
	type ScrapeGroup,
	type ScrapeResultTest,
	ScrapeStatus,
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

const defaultEndpoint: Endpoint = {
	id: "",
	name: "",
	url: "",
	detailFieldSelectors: [],
	mainElementSelector: "",
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

const FirstStepContent: FC<{
	endpoint: Endpoint;
	setEndpoint: React.Dispatch<React.SetStateAction<Endpoint>>;
	firstStepErrors: { [key: string]: string };
	setFirstStepErrors: React.Dispatch<
		React.SetStateAction<{ [key: string]: string }>
	>;
	handleTestGettingElement: () => Promise<void>;
	testElementLoading: boolean;
	testElementError: string | null;
	testElementResult: string | null;
	validateFirstStep: (ep: Endpoint) => boolean;
}> = ({
	endpoint,
	setEndpoint,
	firstStepErrors,
	handleTestGettingElement,
	testElementLoading,
	testElementError,
	testElementResult,
	validateFirstStep,
}) => {
	return (
		<div className="w-[450px]">
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

			<Button
				className="btn btn-primary btn-sm"
				onClick={handleTestGettingElement}
				disabled={testElementLoading}
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
				<div className="bg-gray-800 p-2 rounded mt-2">
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
	handleExtractSelectorsForAllFields: (remarks: Remark[]) => Promise<void>;
	handleExtractSelectorForField: (
		field: Field,
		remark: string,
	) => Promise<void>;
	totalCost: number;
	fieldsWithLoadingSelectors: string[];
	validateSecondStep: (ep: Endpoint) => boolean;
	loadingSampleData: boolean;
	handleTestScrape: (ep: Endpoint) => Promise<void>;
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
		<div>
			<div>
				<label className="label cursor-pointer w-[600px]">
					<span className="label-text text-lg">Name:</span>{" "}
					<strong className="text-xl">{endpoint.name}</strong>
				</label>
			</div>
			<div className="flex items-center gap-5 mb-2  w-[600px] justify-between">
				<label className="label cursor-pointer ">
					<span className="label-text text-lg">URL:</span>
				</label>
				<label className="input input-bordered input-sm flex items-center gap-1 flex-1 ">
					<input
						type="text"
						readOnly
						className="grow"
						defaultValue={endpoint.url}
					/>
					<button
						onClick={() => {
							navigator.clipboard.writeText(endpoint.url).then(() => {
								toast.success("Endpoit URL successfully copied to clipboard", {
									autoClose: 1000,
								});
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
				<label className="label cursor-pointer  w-[600px] mb-2">
					<span className="label-text text-lg">
						Scrape in backgroundMain Element Selector:
					</span>
					<strong>{endpoint.mainElementSelector}</strong>
				</label>
			</div>
			<div className="form-control  w-[600px] mb-2">
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
			<label className="label">Pagination Config</label>
			<select
				className="select select-bordered w-full mb-4 "
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
				className="input input-bordered flex items-center gap-2 w-fullmb-4"
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
			<div className="flex gap-5 mb-10 w-full">
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
							labelClassName="label label"
							className="input input-bordered flex items-center gap-2"
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

			<div className="w-full flex justify-end">
				<div className="flex flex-col items-end mb-2">
					<Button
						className="btn btn-primary btn-sm mb-1"
						onClick={() => {
							handleExtractSelectorsForAllFields(remarks);
						}}
						disabled={loadingSampleData || !!fieldsWithLoadingSelectors.length}
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

			<div className="w-full">
				{fields.map((field, idx) => (
					<div key={field.id}>
						<div className="flex gap-3">
							<button
								className="btn btn-square btn-sm btn-outline border-0"
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
								style={{
									display: "flex",
									alignItems: "center",
									justifyContent: "center",
									marginTop: idx === 0 ? 43 : 25,
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
										className="size-5  text-error"
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
								className="flex gap-3"
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
								<div className="w-36">
									<TextInput
										labelClassName="label"
										className="input input-bordered flex items-center gap-2"
										wrapperClassName="form-control mb-4"
										label={idx > 0 ? "" : "Field Name"}
										readOnly
										disabled
										value={field.name}
									/>
								</div>
								<div className="w-36">
									<TextInput
										labelClassName="label"
										className="input input-bordered flex items-center gap-2"
										wrapperClassName="form-control mb-4"
										label={idx > 0 ? "" : "Field Key"}
										readOnly
										disabled
										value={field.key}
									/>
								</div>
								<div className="flex-1">
									<TextInput
										labelClassName="label"
										className="input input-bordered flex items-center gap-2 w-fullmb-1"
										wrapperClassName="form-control "
										label={idx > 0 ? "" : "Selector"}
										value={
											endpoint.detailFieldSelectors.find(
												(selector) => selector.fieldId === field.id,
											)?.selector
										}
										onChange={(e) => {
											const newEndpoint = {
												...endpoint,
												detailFieldSelectors: endpoint.detailFieldSelectors.map(
													(selector) =>
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
								<div className="w-28">
									<TextInput
										labelClassName="label"
										className="input input-bordered flex items-center gap-2"
										wrapperClassName="form-control mb-4"
										label={idx > 0 ? "" : "Attribute"}
										value={
											endpoint.detailFieldSelectors.find(
												(selector) => selector.fieldId === field.id,
											)?.attributeToGet
										}
										onChange={(e) => {
											const newEndpoint = {
												...endpoint,
												detailFieldSelectors: endpoint.detailFieldSelectors.map(
													(selector) =>
														selector.fieldId === field.id
															? { ...selector, attributeToGet: e.target.value }
															: selector,
												),
											};
											setEndpoint(newEndpoint);
											validateSecondStep(newEndpoint);
										}}
									/>
								</div>
								<div className="w-36">
									<TextInput
										labelClassName="label"
										className="input input-bordered flex items-center gap-2"
										wrapperClassName="form-control mb-4"
										label={idx > 0 ? "" : "Regex"}
										value={
											endpoint.detailFieldSelectors.find(
												(selector) => selector.fieldId === field.id,
											)?.regex
										}
										onChange={(e) => {
											const newEndpoint = {
												...endpoint,
												detailFieldSelectors: endpoint.detailFieldSelectors.map(
													(selector) =>
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
								<div className="w-24">
									<TextInput
										type="number"
										min={0}
										step={1}
										labelClassName="label"
										className="input input-bordered flex items-center gap-2"
										wrapperClassName="form-control mb-4"
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
										className="input input-bordered flex items-center gap-2"
										wrapperClassName="form-control mb-4"
										label={idx > 0 ? "" : "Remarks"}
										tooltip={
											"This data is not saved, only helpfull for smart extract)"
										}
										value={remarks.find((r) => r.fieldId === field.id)!.remark}
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
								className="btn btn-square btn-sm btn-outline border-0"
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
									display: "flex",
									alignItems: "center",
									justifyContent: "center",
									userSelect: fieldsWithLoadingSelectors.includes(field.id)
										? "none"
										: "auto",
									cursor: fieldsWithLoadingSelectors.includes(field.id)
										? "not-allowed"
										: "pointer",
									marginTop: idx === 0 ? 43 : 25,
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
			{sampleData.length > 0 && (
				<>
					<div
						className="w-full mb-5"
						style={{
							maxHeight: 500,
							overflow: "auto",
						}}
					>
						<table className="border-collapse bg-gray-800 shadow-md rounded-lg w-full">
							<thead className="bg-gray-700 sticky top-0">
								<tr>
									{fields.map((field) => (
										<th
											key={field.id}
											className="px-4 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-600 transition whitespace-nowrap"
										>
											{field.name}
										</th>
									))}
								</tr>
							</thead>
							<tbody className="bg-gray-800 divide-y divide-gray-700">
								{sampleData.map((row) => (
									<tr key={row.id} className="hover:bg-gray-700 transition">
										{fields.map((field) => (
											<td key={field.id} className="text-sm text-gray-300">
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
							// eslint-disable-next-line unused-imports/no-unused-vars
							shouldExpandNode={() => true}
							style={darkStyles}
						/>
					</div>
				</>
			)}
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

	const handleTestGettingElement = useCallback(async () => {
		try {
			setTestElementError(null);
			setTestElementResult(null);
			setTestElementLoading(true);
			const response = await axios.post("/api/selectors/test", {
				url: endpoint.url,
				mainElementSelector: endpoint.mainElementSelector,
			});
			setTestElementResult(response.data.html);
		} catch (error) {
			console.error(error);
			setTestElementError("Error while testing element");
		} finally {
			setTestElementLoading(false);
		}
	}, [endpoint.mainElementSelector, endpoint.url]);

	const handleTestScrape = useCallback(
		async (ep: Endpoint) => {
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
			setSampleData([]);
			setLoadingSampleData(true);
			try {
				const response = await axios.post("/api/scrape/endpoint-test", {
					group,
				});
				setSampleData(response.data);
			} catch (error) {
				console.error(error);
				toast.error("Failed to test scrape");
			} finally {
				setLoadingSampleData(false);
			}
		},
		[fields],
	);

	const handleExtractSelectorForField = useCallback(
		async (field: Field, remark: string) => {
			if (fieldsWithLoadingSelectors.includes(field.id)) {
				return;
			}
			setFieldsWithLoadingSelectors((prev) => [...prev, field.id]);
			console.log("Extracting selector for field", field);
			try {
				const response = await axios.post("/api/selectors/extract", {
					url: endpoint.url,
					mainElementSelector: endpoint.mainElementSelector,
					fieldsToExtractSelectorsFor: [
						{
							key: field.key,
							name: field.name,
							type: field.type,
							remark,
						},
					],
				});
				console.log("Extracted selectors", response.data);
				const newEndpoint: Endpoint = {
					...endpoint,
					detailFieldSelectors: endpoint.detailFieldSelectors.map((selector) =>
						selector.fieldId === field.id
							? {
									...selector,
									regexMatchIndexToUse:
										response.data.fields[0]?.regexMatchIndexToUse,
									selector: response.data.fields[0]?.selector || "",
									attributeToGet: response.data.fields[0]?.attributeToGet || "",
									regex: response.data.fields[0]?.regex || "",
								}
							: selector,
					),
				};
				setEndpoint(newEndpoint);
				setTotalCost((prev) => prev + response.data.totalCost);
				setFieldsWithLoadingSelectors((prev) =>
					prev.filter((id) => id !== field.id),
				);
				handleTestScrape(newEndpoint);
			} catch (error) {
				setFieldsWithLoadingSelectors([]);
				console.error(error);
				toast.error("Failed to extract selector for field");
			}
		},
		[endpoint, fieldsWithLoadingSelectors, handleTestScrape],
	);

	const handleExtractSelectorsForAllFields = useCallback(
		async (remarks: Remark[]) => {
			const toExtract = fields.filter(
				(f) =>
					!endpoint.detailFieldSelectors.find((df) => df.fieldId === f.id)
						?.lockedForEdit,
			);

			if (toExtract.length === 0) {
				return;
			}
			setFieldsWithLoadingSelectors(toExtract.map((f) => f.id));
			const response = await axios.post("/api/selectors/extract", {
				url: endpoint.url,
				mainElementSelector: endpoint.mainElementSelector,
				fieldsToExtractSelectorsFor: toExtract.map((field) => ({
					key: field.key,
					name: field.name,
					type: field.type,
					remark: remarks.find((r) => r.fieldId === field.id)?.remark,
				})),
			});
			console.log("Extracted selectors", response.data);
			const newEndpoint: Endpoint = {
				...endpoint,
				detailFieldSelectors: endpoint.detailFieldSelectors.map((selector) => {
					const field = fields.find((f) => f.id === selector.fieldId);
					console.log("Field", field, selector);
					const extractedField = response.data.fields.find(
						// biome-ignore lint/suspicious/noExplicitAny: <explanation>
						(extractedField: any) => extractedField.field === field?.key,
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
			setTotalCost((prev) => prev + response.data.totalCost);
			setEndpoint(newEndpoint);

			setFieldsWithLoadingSelectors([]);
			handleTestScrape(newEndpoint);
		},
		[fields, endpoint, handleTestScrape],
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
		if (ep.mainElementSelector.trim() === "") {
			errors.mainElementSelector = "Main Element Selector is required";
		}
		const urlRegex = /^(http|https):\/\/[^\s\/$.?#].[^\s]*$/g;
		if (!urlRegex.test(ep.url)) {
			errors.url = "URL is not valid";
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
