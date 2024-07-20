import { FC, useState } from "react";
import { Modal, ModalProps } from "../ui/Modal";
import { ExportType } from "../../types";

type ConfigureExportModalProps = Pick<ModalProps, "isOpen" | "onClose"> & {
  onConfirm: (name: string, type: ExportType) => void;
  defaultName?: string;
};

const exportTypeOptions = [
  { label: ".xml", value: ExportType.XML },
  { label: ".json", value: ExportType.JSON },
  { label: ".csv", value: ExportType.CSV },
  { label: ".xlsx", value: ExportType.EXCEL },
] as const;

export const ConfigureExportModal: FC<ConfigureExportModalProps> = ({
  onConfirm,
  isOpen,
  onClose,
  defaultName,
}) => {
  const [name, setName] = useState(defaultName ?? "");
  const [type, setType] = useState<(typeof exportTypeOptions)[number]>(
    exportTypeOptions[0],
  );
  const allowSubmit = name.trim() !== "" && name.indexOf(".") === -1;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Export Results"
      actions={[
        {
          label: "Cancel",
          onClick: onClose,
          className: "bg-gray-500 text-white",
        },
        {
          label: "Create",
          onClick: () => {
            onConfirm(name.trim(), type.value);
            setName("");
          },
          className: "bg-blue-500 text-white",
          disabled: !allowSubmit,
        },
      ]}
    >
      <div className="space-y-2 w-[450px]">
        <label className="label">File Name</label>
        <label className="flex gap-3">
          <input
            type="text"
            name="name"
            id="name"
            className="input input-bordered w-full"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <select
            value={type.value}
            onChange={(e) =>
              setType(
                exportTypeOptions.find((o) => o.value === e.target.value)!,
              )
            }
            className="select select-bordered"
          >
            {exportTypeOptions.map((option) => (
              <option
                key={option.value}
                value={option.value}
                selected={option.value === type.value}
              >
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </div>
    </Modal>
  );
};
