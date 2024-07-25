import { type FC } from "react";
import { Modal, type ModalProps } from "../ui/Modal";

type ConfirmDeleteGroupProps = Pick<ModalProps, "isOpen" | "onClose"> & {
	onConfirm: () => void;
};

export const ConfirmDeleteGroup: FC<ConfirmDeleteGroupProps> = ({
	onConfirm,
	isOpen,
	onClose,
}) => {
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
					label: "Confirm",
					onClick: onConfirm,
					className: "bg-blue-500 text-white",
				},
			]}
		>
			<div className="space-y-2 w-[450px]">
				<p className="text-warning mb-3">
					Are you sure you want to delete this group? All scraping results
					including archived groups will be removed. This action cannot be
					undone.
				</p>
			</div>
		</Modal>
	);
};
