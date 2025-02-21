import { FC } from "react";
import { Modal, ModalProps } from "../ui/Modal";

type ConfirmRemoveEndpointResultsProps = Pick<
	ModalProps,
	"isOpen" | "onClose"
> & {
	onConfirm: () => void;
};

export const ConfirmRemoveEndpointResults: FC<
	ConfirmRemoveEndpointResultsProps
> = ({ onConfirm, isOpen, onClose }) => {
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
					onClick: onConfirm,
					className: "bg-blue-500 text-white",
				},
			]}
		>
			<div className="space-y-2 w-[450px]">
				<p className="text-warning mb-3">
					Are you sure you want to remove all scraped results for this endpoint?
					All scraping results will be removed (you can export them before
					deletion) This action cannot be undone.
				</p>
			</div>
		</Modal>
	);
};
