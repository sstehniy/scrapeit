import { type FC, useState } from "react";
import { v4 } from "uuid";
import {
	FieldType,
	NotificationCondition,
	type NotificationConfig,
	type ScrapeGroup,
} from "../../types";
import { Modal, type ModalProps } from "../ui/Modal";
import { TextInput } from "../ui/TextInput";

type ConfigureGroupNotificationModalProps = Pick<
	ModalProps,
	"isOpen" | "onClose"
> & {
	onConfirm: (configs: NotificationConfig[]) => void;
	groupNotificationConfigs?: NotificationConfig[];
	group: ScrapeGroup;
};

export const ConfigureGroupNotificationModal: FC<
	ConfigureGroupNotificationModalProps
> = ({ onConfirm, isOpen, onClose, groupNotificationConfigs, group }) => {
	const [configs, setConfigs] = useState<NotificationConfig[]>(
		groupNotificationConfigs || [
			{
				id: v4(),
				name: group.name + " notification 1",
				conditions: [],
				fieldIdsToNotify: [],
				groupId: group.id,
			},
		],
	);

	const [hoveredCondition, setHoveredCondition] = useState<number | null>(null);

	const fieldsForConditions = group.fields.filter((field) => {
		return field.type === FieldType.NUMBER;
	});

	const operatorOptions = [
		{ label: ">", value: ">" },
		{ label: "<", value: "<" },
		{ label: "=", value: "=" },
		{ label: "!=", value: "!=" },
	];

	const validateConfig = (configs: NotificationConfig[]) => {
		return configs.every((config) => {
			if (
				!config.name.trim() ||
				config.conditions.length === 0 ||
				config.fieldIdsToNotify.length === 0 ||
				!config.conditions.every((condition) => {
					return condition.fieldId && condition.operator && condition.value;
				})
			) {
				return false;
			}
			return true;
		});
	};

	const handleConfirm = (configs: NotificationConfig[]) => {
		if (validateConfig(configs)) {
			onConfirm(configs);
		}
	};

	return (
		<Modal
			isOpen={isOpen}
			onClose={onClose}
			title="Configure Notifications"
			actions={[
				{
					label: "Cancel",
					onClick: onClose,
					className: "bg-gray-500 text-white",
				},
				{
					label: "Save",
					onClick: () => {
						handleConfirm(configs);
					},
					className: "bg-blue-500 text-white",
					disabled: !validateConfig(configs),
				},
			]}
		>
			<div className="space-y-2 w-[500px]">
				{configs.map((config) => (
					<div
						className="collapse collapse-plus  bg-base-200 border-2 border-secondary-content relative"
						key={config.id}
					>
						<button
							type="button"
							className="absolute btn btn-xs btn-square btn-error border-0 cursor-pointer"
							onClick={() => {
								const newConfigs = configs.filter((c) => c.id !== config.id);
								setConfigs(newConfigs);
								validateConfig(newConfigs);
							}}
							style={{
								zIndex: 100,
								top: "1.1rem",
								right: "3rem",
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
						<input type="radio" name="my-accordion-3" defaultChecked />
						<div className="collapse-title text-lg font-medium">
							{config.name}
						</div>
						<div className="collapse-content">
							<TextInput
								labelClassName="label"
								className="input input-bordered flex items-center gap-2"
								wrapperClassName="form-control mb-4"
								label="Config Name"
								name="name"
								id="name"
								value={config.name}
								onChange={(e) => {
									const newConfigs = configs.map((c) =>
										c.id === config.id
											? {
													...c,
													name: e.target.value,
												}
											: c,
									);
									setConfigs(newConfigs);
									validateConfig(newConfigs);
								}}
								required
							/>
							<label className="label text-lg px-0">Fields to notify</label>
							<div className="grid grid-cols-2 gap-2 mb-3">
								{[...group.fields].map((field) => {
									const isChecked = config.fieldIdsToNotify.includes(field.id);
									return (
										<div
											key={field.id + config.id}
											className="flex items-center"
										>
											<input
												type="checkbox"
												id={field.id + config.id}
												checked={isChecked}
												onChange={(e) => {
													console.log(config.id);
													const newConfig = configs.map((c) =>
														c.id === config.id
															? {
																	...c,
																	fieldIdsToNotify: e.target.checked
																		? [...c.fieldIdsToNotify, field.id]
																		: c.fieldIdsToNotify.filter(
																				(ftn) => ftn !== field.id,
																			),
																}
															: c,
													);

													setConfigs(newConfig);
													validateConfig(newConfig);
												}}
												className="hidden"
											/>
											<label
												htmlFor={field.id + config.id}
												className={`cursor-pointer btn btn-sm  w-full rounded-sm border ${
													isChecked ? "btn-accent" : "btn-secondary btn-outline"
												} transition-colors duration-300`}
												title={field.name}
											>
												<span
													style={{
														overflow: "hidden",
														textOverflow: "ellipsis",
														whiteSpace: "nowrap",
														maxWidth: 100,
													}}
												>
													{field.name}
												</span>
											</label>
										</div>
									);
								})}
							</div>

							<div className="flex gap-2 items-center justify-between mb-2">
								<label className="label text-lg px-0">Conditions</label>
								<button
									className="btn btn-sm btn-primary btn-outline"
									type="button"
									onClick={() => {
										const newCondition: NotificationCondition = {
											fieldId: "",
											operator: "=",
											value: 0,
										};
										const newConfigs = configs.map((c) =>
											c.id === config.id
												? {
														...c,
														conditions: [...c.conditions, newCondition],
													}
												: c,
										);
										setConfigs(newConfigs);
										validateConfig(newConfigs);
									}}
								>
									+ Add condition
								</button>
							</div>
							{config.conditions.map((condition, index) => {
								return (
									<div
										key={index}
										onMouseEnter={() => setHoveredCondition(index)}
										onMouseLeave={() => setHoveredCondition(null)}
										className="flex items-center flex-wrap border-2 rounded-lg border-base-content border-opacity-25 shadow-sm shrink-0 relative mb-2"
									>
										{hoveredCondition === index && (
											<button
												type="button"
												className="absolute btn-circle btn btn-xs btn-error"
												onClick={() => {
													const newConfigs = configs.map((c) =>
														c.id === config.id
															? {
																	...c,
																	conditions: c.conditions.filter(
																		(_, i) => i !== index,
																	),
																}
															: c,
													);
													setConfigs(newConfigs);
													validateConfig(newConfigs);
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
											className="select select-sm  h-10 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0 border-r-2 border-base-content border-opacity-25 rounded-r-none"
											value={condition.fieldId}
											onChange={(e) => {
												const newConfigs = configs.map((c) =>
													c.id === config.id
														? {
																...c,
																conditions: c.conditions.map((cnd, i) =>
																	i === index
																		? {
																				...cnd,
																				fieldId: e.target.value,
																			}
																		: cnd,
																),
															}
														: c,
												);
												setConfigs(newConfigs);
												validateConfig(newConfigs);
											}}
										>
											<option value="">Select Field</option>
											{fieldsForConditions.map((field) => (
												<option key={field.id} value={field.id}>
													{field.name}
												</option>
											))}
										</select>
										<select
											className="select select-sm w-16 h-10 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0 border-r-2 border-base-content border-opacity-25 rounded-r-none"
											value={condition.operator}
											onChange={(e) => {
												const newConfigs = configs.map((c) =>
													c.id === config.id
														? {
																...c,
																conditions: c.conditions.map((cnd, i) =>
																	i === index
																		? {
																				...cnd,
																				operator: e.target.value,
																			}
																		: cnd,
																),
															}
														: c,
												);
												setConfigs(newConfigs);
												validateConfig(newConfigs);
											}}
										>
											{operatorOptions.map(({ label, value }) => (
												<option key={value} value={value}>
													{label}
												</option>
											))}
										</select>
										<input
											className="input input-sm h-10 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0 flex-1"
											type="number"
											value={condition.value ?? 0}
											onChange={(e) => {
												const value: number = Number.parseFloat(e.target.value);
												const newConfigs = configs.map((c) =>
													c.id === config.id
														? {
																...c,
																conditions: c.conditions.map((cnd, i) =>
																	i === index
																		? {
																				...cnd,
																				value,
																			}
																		: cnd,
																),
															}
														: c,
												);
												setConfigs(newConfigs);
												validateConfig(newConfigs);
											}}
										/>
									</div>
								);
							})}
							{!config.conditions.length && <p>No conditions added</p>}
						</div>
					</div>
				))}
				<button
					className="w-full btn btn-success"
					onClick={() => {
						setConfigs((prev) => [
							...prev,
							{
								id: v4(),
								name: group.name + " Notification " + (configs.length + 1),
								conditions: [],
								fieldIdsToNotify: [],
								groupId: group.id,
							},
						]);
					}}
				>
					+ Add Config
				</button>
			</div>
		</Modal>
	);
};
