/** eslint-disable prettier/prettier */
/** eslint-disable unused-imports/no-unused-vars */
/** eslint-disable no-empty */
import { FC, useCallback, useEffect, useRef, useState } from "react";
import Sortable from "sortablejs";
import { v4 } from "uuid";
import { Field, FieldType } from "../../types";
import { FieldChange } from "../../views/GroupView";
import { Button } from "../ui/Button";
import { Modal, ModalProps } from "../ui/Modal";
import { TextInput } from "../ui/TextInput";

type ConfigureGroupSchemaProps = Pick<ModalProps, "isOpen" | "onClose"> & {
	onConfirm: (fields: Field[], fieldChanges: FieldChange[]) => void;
	fieldsToEdit: Field[];
};

const defaultFields: Field[] = [
	{
		id: v4(),
		isFullyEditable: false,
		key: "unique_identifier",
		name: "ID",
		type: FieldType.TEXT,
		order: 0,
	},
	{
		id: v4(),
		isFullyEditable: false,
		name: "Link",
		key: "link",
		type: FieldType.LINK,
		order: 1,
	},
	{
		id: v4(),
		isFullyEditable: true,
		name: "Image",
		key: "image",
		type: FieldType.IMAGE,
		order: 2,
	},
];

// eslint-disable-next-line prettier/prettier
export const ConfigureGroupSchema: FC<ConfigureGroupSchemaProps> = ({
	onConfirm,
	isOpen,
	onClose,
	fieldsToEdit,
}) => {
	const [fields, setFields] = useState<Field[]>(
		fieldsToEdit.length ? fieldsToEdit : defaultFields,
	);
	const [fieldsValid, setFieldsValid] = useState(true);
	const [fieldsUnique, setFieldsUnique] = useState(true);
	const [fieldChanges, setFieldChanges] = useState<FieldChange[]>([]);

	const sortableRef = useRef<Sortable | null>(null);

	useEffect(() => {
		if (listRef.current) {
			sortableRef.current = new Sortable(listRef.current, {
				handle: ".handle",
				onEnd: (evt) => {
					const oldIndex = evt.oldIndex || 0;
					const newIndex = evt.newIndex || 0;

					setFields((prev) => {
						const newFields = [...prev];
						const [removed] = newFields.splice(oldIndex, 1);
						newFields.splice(newIndex, 0, removed);
						newFields.forEach((field, idx) => {
							field.order = idx + 1;
						});
						return newFields;
					});
				},
			});
		}
		return () => {
			sortableRef.current?.destroy();
		};
	}, [fields, isOpen]);

	const listRef = useRef<HTMLDivElement>(null);

	const validateFields = useCallback((fieldsToValidate: Field[]) => {
		setFieldsValid(true);
		setFieldsUnique(true);
		const uniqueKeys = new Set<string>();
		const uniqueNames = new Set<string>();
		let valid = true;
		let unique = true;
		fieldsToValidate.forEach((field) => {
			if (!field.name || !field.key) {
				valid = false;
			}
			if (uniqueKeys.has(field.key) || uniqueNames.has(field.name)) {
				unique = false;
			}
			uniqueKeys.add(field.key);
			uniqueNames.add(field.name);
		});
		setFieldsValid(valid);
		setFieldsUnique(unique);
		return valid && unique;
	}, []);

	const keyEditable = (field: Field) => {
		if (!field.isFullyEditable) {
			return false;
		}
		return true;
	};

	// eslint-disable-next-line unused-imports/no-unused-vars
	const handleFieldChange = useCallback(
		(fieldId: string, type: FieldChange["type"], value?: string) => {
			let newFields = [...fields];

			const fieldIsNewSinceLastSave = !fieldsToEdit.find(
				(field) => field.id === fieldId,
			);
			switch (type) {
				case "add_field": {
					if (!fieldIsNewSinceLastSave) {
						return;
					}
					const newField = {
						id: v4(),
						name: "",
						key: "",
						type: FieldType.TEXT,
						isFullyEditable: true,
						order: newFields.length + 1,
					};
					newFields.push(newField);
					setFieldChanges([
						...fieldChanges,
						{ fieldId: newField.id, type, fieldIsNewSinceLastSave: true },
					]);
					break;
				}
				// eslint-disable-next-line no-empty
				case "delete_field": {
					/* empty */
					if (fieldIsNewSinceLastSave) {
						newFields = newFields.filter((field) => field.id !== fieldId);
						setFieldChanges((prev) =>
							prev.filter((change) => change.fieldId !== fieldId),
						);
					} else {
						const field = fieldsToEdit.find((field) => field.id === fieldId);
						if (!field) {
							return;
						}
						newFields = newFields.filter((field) => field.id !== fieldId);
						setFieldChanges([
							...fieldChanges,
							{ fieldId, type, fieldIsNewSinceLastSave },
						]);
					}
					break;
				}
				case "change_field_name": {
					const field = newFields.find((field) => field.id === fieldId);
					if (!field || value === undefined) {
						return;
					}

					field.name = value;

					const sameChangeIdx = fieldChanges.findIndex(
						(change) => change.fieldId === fieldId && change.type === type,
					);
					if (sameChangeIdx !== -1) {
						setFieldChanges((prev) => {
							const newChanges = [...prev];
							newChanges[sameChangeIdx] = {
								fieldId,
								type,
								fieldIsNewSinceLastSave,
							};
							return newChanges;
						});
						break;
					}

					setFieldChanges([
						...fieldChanges,
						{ fieldId, type, fieldIsNewSinceLastSave },
					]);

					break;
				}
				case "change_field_key":
				case "change_field_type": {
					const field = newFields.find((field) => field.id === fieldId);
					if (!field || !field.isFullyEditable || value === undefined) {
						return;
					}

					if (type === "change_field_key") {
						field.key = value;
					} else if (type === "change_field_type") {
						field.type = value as FieldType;
					}

					const sameChangeIdx = fieldChanges.findIndex(
						(change) => change.fieldId === fieldId && change.type === type,
					);
					if (sameChangeIdx !== -1) {
						setFieldChanges((prev) => {
							const newChanges = [...prev];
							newChanges[sameChangeIdx] = {
								fieldId,
								type,
								fieldIsNewSinceLastSave,
							};
							return newChanges;
						});
						break;
					}

					setFieldChanges([
						...fieldChanges,
						{ fieldId, type, fieldIsNewSinceLastSave },
					]);

					break;
				}
			}
			validateFields(newFields);
			setFields(newFields);
		},
		[fieldChanges, fields, fieldsToEdit, validateFields],
	);

	return (
		<Modal
			isOpen={isOpen}
			onClose={onClose}
			title="Define Fields"
			actions={[
				{
					label: "Cancel",
					onClick: onClose,
					className: "bg-gray-500 text-white",
				},
				{
					label: fieldsToEdit.length ? "Save" : "Create",
					onClick: () => {
						if (!validateFields(fields)) {
							return;
						}
						onConfirm(fields, fieldChanges);
						onClose();
					},
					className: "bg-blue-500 text-white",
					disabled: !fields.length || !fieldsValid || !fieldsUnique,
				},
			]}
		>
			<label className="block font-medium text-xl text-gray-500 mb-2">
				Fields
			</label>
			<div
				className="space-y-2 w-[500px] max-h-[400px] overflow-y-auto px-1 mb-2"
				ref={listRef}
			>
				{fields.map((field) => (
					<div key={field.id} className="flex items-center space-x-2">
						<div className="handle cursor-move mt-3 -ms-2 me-2">
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
									d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
								/>
							</svg>
						</div>
						<TextInput
							labelClassName="label"
							className="input input-bordered flex items-center gap-2"
							wrapperClassName="form-control mb-4"
							label="Name"
							name="name"
							id="name"
							placeholder="Field Name"
							value={field.name}
							onChange={(e) => {
								handleFieldChange(
									field.id,
									"change_field_name",
									e.target.value,
								);
							}}
							disabled={!keyEditable(field)}
							required
						/>
						<TextInput
							labelClassName="label"
							className="input input-bordered flex items-center gap-2"
							wrapperClassName="form-control mb-4"
							label="Key"
							name="key"
							id="key"
							placeholder="Field Key"
							value={field.key}
							onChange={(e) => {
								handleFieldChange(field.id, "change_field_key", e.target.value);
							}}
							required
							disabled={!keyEditable(field)}
						/>

						<div className="w-40">
							<label className="label">Type</label>
							<select
								value={field.type}
								onChange={(e) => {
									handleFieldChange(
										field.id,
										"change_field_type",
										e.target.value,
									);
								}}
								className="select select-bordered w-full mb-4"
								disabled={!field.isFullyEditable}
							>
								{Object.values(FieldType).map((type) => (
									<option key={type} value={type}>
										{type}
									</option>
								))}
							</select>
						</div>

						{field.isFullyEditable ? (
							<Button
								className="btn btn-sm btn-outline border-0 btn-square btn-secondary mt-2"
								onClick={() => {
									handleFieldChange(field.id, "delete_field");
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
							</Button>
						) : (
							<Button
								className="btn btn-sm btn-outline border-0 btn-square btn-secondary mt-2 hover:bg-transparent"
								style={{
									visibility: "hidden",
									pointerEvents: "none",
									opacity: 0,
								}}
							></Button>
						)}
					</div>
				))}
			</div>
			{!fieldsValid && (
				<p className="text-red-500 text-xs">All fields are required</p>
			)}
			{!fieldsUnique && (
				<p className="text-red-500 text-xs">
					Field names and keys must be unique
				</p>
			)}
			{fieldChanges.filter(
				(change) =>
					change.type === "delete_field" &&
					fieldsToEdit.find((f) => f.id === change.fieldId),
			).length > 0 && (
				<p className="text-red-500 text-xs">
					Following existing fields will be deleted:{" "}
					{fieldChanges
						.filter(
							(change) =>
								change.type === "delete_field" &&
								fieldsToEdit.find((f) => f.id === change.fieldId),
						)
						.map(
							(change) =>
								fieldsToEdit.find((f) => f.id === change.fieldId)?.name,
						)
						.join(", ")}
				</p>
			)}
			<div className="flex justify-end mt-3">
				<Button
					onClick={() => {
						handleFieldChange("", "add_field");
						setTimeout(() => {
							listRef.current?.scrollTo({
								top: listRef.current.scrollHeight,
								behavior: "smooth",
							});
						}, 50);
					}}
					className="btn btn-primary btn-sm"
				>
					Add Field
				</Button>
			</div>
		</Modal>
	);
};
