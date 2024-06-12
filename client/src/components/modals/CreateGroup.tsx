import { FC, useState } from "react";
import { Modal, ModalProps } from "../ui/Modal";

type CreateGroupModalProps = Pick<ModalProps, "isOpen" | "onClose"> & {
  onConfirm: (name: string) => void;
};

export const CreateGroupModal: FC<CreateGroupModalProps> = ({
  onConfirm,
  isOpen,
  onClose,
}) => {
  const [name, setName] = useState("");

  const getCreateGroupContent = () => {
    return (
      <div className="space-y-4">
        <label className="block text-sm font-medium text-gray-700">Name</label>
        <input
          type="text"
          name="name"
          id="name"
          className="mt-1 p-2 border border-gray-300 rounded w-full"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </div>
    );
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create new Group"
      content={getCreateGroupContent()}
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
        },
      ]}
    />
  );
};
