import { FC } from "react";
import { Modal, ModalProps } from "../ui/Modal";

type ConfirmRemoveEndpointProps = Pick<ModalProps, "isOpen" | "onClose"> & {
  onConfirm: () => void;
};

export const ConfirmRemoveEndpoint: FC<ConfirmRemoveEndpointProps> = ({
  onConfirm,
  isOpen,
  onClose,
}) => {
  const getConfirmRemoveEndpointContent = () => {
    return (
      <div className="space-y-2 w-[450px]">
        <p className="text-warning mb-3">
          Are you sure you want to remove this endpoint? All scraping results
          will be removed (you can export them before deletion) This action
          cannot be undone.
        </p>
      </div>
    );
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create new Group"
      content={getConfirmRemoveEndpointContent()}
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
    />
  );
};
