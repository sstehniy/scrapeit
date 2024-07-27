import { type FC, useEffect, useState } from "react";
import { Modal, type ModalProps } from "../ui/Modal";
import { TextInput } from "../ui/TextInput";

type CreateGroupModalProps = Pick<ModalProps, "isOpen" | "onClose"> & {
	onConfirm: (data: { name: string }) => void;
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
						onConfirm({ name });
						setName("");
					},
					className: "bg-blue-500 text-white",
					disabled: !name.trim(),
				},
			]}
		>
			<div className="space-y-2 w-[450px]">
				{/* <label className="label">Name</label>
				<input
					type="text"
					name="name"
					id="name"
					className="input input-bordered flex items-center gap-2 w-full"
					value={name}
					onChange={(e) => setName(e.target.value)}
					required
				/> */}
				<TextInput
					labelClassName="label"
					className="input input-bordered flex items-center gap-2"
					wrapperClassName="form-control mb-4"
					label="Name"
					name="name"
					id="name"
					value={name}
					onChange={(e) => {
						setName(e.target.value);
					}}
					required
				/>
			</div>
		</Modal>
	);
};
