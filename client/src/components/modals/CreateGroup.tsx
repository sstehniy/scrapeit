import { type FC, useEffect, useState } from "react";
import { Modal, type ModalProps } from "../ui/Modal";

type CreateGroupModalProps = Pick<ModalProps, "isOpen" | "onClose"> & {
	onConfirm: (name: string) => void;
};

export const CreateGroupModal: FC<CreateGroupModalProps> = ({
	onConfirm,
	isOpen,
	onClose,
}) => {
	const [name, setName] = useState("");

	useEffect(() => {
		if (isOpen) {
			setName("");
		}
	}, [isOpen]);

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
						onConfirm(name);
						setName("");
					},
					className: "bg-blue-500 text-white",
					disabled: !name.trim(),
				},
			]}
		>
			<div className="space-y-2 w-[450px]">
				<label className="label">Name</label>
				<input
					type="text"
					name="name"
					id="name"
					className="p-2 border border-gray-300 rounded w-full "
					value={name}
					onChange={(e) => setName(e.target.value)}
					required
				/>
			</div>
		</Modal>
	);
};
