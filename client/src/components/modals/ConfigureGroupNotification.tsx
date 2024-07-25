import { type FC, useState } from "react";
import { Modal, type ModalProps } from "../ui/Modal";
import {
	FieldType,
	type NotificationConfig,
	type ScrapeGroup,
} from "../../types";

type ConfigureGroupNotificationModalProps = Pick<
	ModalProps,
	"isOpen" | "onClose"
> & {
	onConfirm: (name: NotificationConfig) => void;
	groupNotificationConfig?: NotificationConfig;
	group: ScrapeGroup;
};

export const ConfigureGroupNotificationModal: FC<
	ConfigureGroupNotificationModalProps
> = ({ onConfirm, isOpen, onClose, groupNotificationConfig, group }) => {
	const [config, setConfig] = useState<NotificationConfig>(
		groupNotificationConfig || {
			groupId: group.id,
			fieldIdsToNotify: [],
			conditions: [],
		},
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

	const validateConfig = (config: NotificationConfig) => {
		if (
			config.conditions.length === 0 ||
			config.fieldIdsToNotify.length === 0 ||
			!config.conditions.every((condition) => {
				return condition.fieldId && condition.operator && condition.value;
			})
		) {
			return false;
		}
		return true;
	};

	const handleConfirm = (config: NotificationConfig) => {
		if (validateConfig(config)) {
			onConfirm(config);
		}
	};

	return (
		<Modal
			isOpen={isOpen}
			onClose={onClose}
			title="Create new Group"
			actions={[
				{
					label: "Cancel",
					onClick: onClose,
					className: "bg-gray-500 text-white",
				},
				{
					label: "Create",
					onClick: () => {
						handleConfirm(config);
					},
					className: "bg-blue-500 text-white",
					disabled: !validateConfig(config),
				},
			]}
		>
			<div className="space-y-2 w-[450px]">
				<label className="label">Notification Config</label>
				<label className="label">Fields to notify</label>
				{group.fields.map((field) => {
					return (
						<div key={field.id} className="flex items-center space-x-2">
							<input
								type="checkbox"
								id={field.id}
								checked={config.fieldIdsToNotify.includes(field.id)}
								onChange={(e) => {
									const newConfig = {
										...config,
									};
									if (e.target.checked) {
										newConfig.fieldIdsToNotify.push(field.id);
									} else {
										newConfig.fieldIdsToNotify =
											newConfig.fieldIdsToNotify.filter(
												(id) => id !== field.id,
											);
									}
									setConfig(newConfig);
									validateConfig(config);
								}}
							/>
							<label htmlFor={field.id}>{field.name}</label>
						</div>
					);
				})}
				<div className="flex gap-2 items-center justify-between">
					<label className="label">Conditions</label>
					<button
						className="btn btn-sm btn-primary"
						type="button"
						onClick={() => {
							const newConfig = {
								...config,
							};
							newConfig.conditions.push({
								fieldId: "",
								operator: "=",
								value: 0,
							});
							setConfig(newConfig);
							validateConfig(config);
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
							className="flex items-center flex-wrap border-2 rounded-lg border-base-content border-opacity-25 shadow-sm shrink-0 relative"
						>
							{hoveredCondition === index && (
								<button
									type="button"
									className="absolute btn-circle btn btn-xs btn-error"
									onClick={() => {
										const newConfig = {
											...config,
											conditions: config.conditions.filter(
												(_, i) => i !== index,
											),
										};
										setConfig(newConfig);
										validateConfig(config);
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
									const newConfig = {
										...config,
										conditions: config.conditions.map((c, i) =>
											i === index ? { ...c, fieldId: e.target.value } : c,
										),
									};
									setConfig(newConfig);
									validateConfig(newConfig);
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
									const newConfig = {
										...config,
										conditions: config.conditions.map((c, i) =>
											i === index ? { ...c, operator: e.target.value } : c,
										),
									};
									setConfig(newConfig);
									validateConfig(newConfig);
								}}
							>
								{operatorOptions.map(({ label, value }) => (
									<option key={value} value={value}>
										{label}
									</option>
								))}
							</select>
							<input
								className="input input-sm h-10 rounded-lg border-0 focus:ring-0 focus:outline-none focus:ring-offset-0"
								type="number"
								value={condition.value ?? 0}
								onChange={(e) => {
									const value: number = Number.parseFloat(e.target.value);
									const newConfig = {
										...config,
										conditions: config.conditions.map((c, i) =>
											i === index ? { ...c, value } : c,
										),
									};
									setConfig(newConfig);
									validateConfig(newConfig);
								}}
							/>
						</div>
					);
				})}
				{!config.conditions.length && <p>No conditions added</p>}
			</div>
		</Modal>
	);
};
