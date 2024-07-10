import { FC, useEffect, useRef, useState } from "react";
import { Modal, ModalProps } from "../ui/Modal";
import { Field, FieldType } from "../../types";
import { v4 } from "uuid";
import { TextInput } from "../ui/TextInput";
import { Button } from "../ui/Button";

type ConfigureGroupSchemaProps = Pick<ModalProps, "isOpen" | "onClose"> & {
  onConfirm: (fields: Field[]) => void;
  fieldsToEdit: Field[];
};

const defaultFields: Field[] = [
  {
    id: v4(),
    isFullyEditable: false,
    name: "",
    key: "link",
    type: FieldType.LINK,
  },
  {
    id: v4(),
    isFullyEditable: false,
    name: "",
    key: "image",
    type: FieldType.IMAGE,
  },
];

export const ConfigureGroupSchema: FC<ConfigureGroupSchemaProps> = ({
  onConfirm,
  isOpen,
  onClose,
  fieldsToEdit,
}) => {
  const [fields, setFields] = useState<Field[]>([]);
  const [fieldsValid, setFieldsValid] = useState(true);
  const [fieldsUnique, setFieldsUnique] = useState(true);

  useEffect(() => {
    if (fieldsToEdit.length) {
      setFields(fieldsToEdit);
    } else {
      setFields(defaultFields);
    }
  }, [fieldsToEdit]);

  const listRef = useRef<HTMLDivElement>(null);

  const validateFields = () => {
    setFieldsValid(true);
    setFieldsUnique(true);
    const uniqueKeys = new Set<string>();
    const uniqueNames = new Set<string>();
    let valid = true;
    let unique = true;
    fields.forEach((field) => {
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
  };

  const keyAndNameEditable = (field: Field) => {
    if (!field.isFullyEditable) {
      return false;
    }
    return fieldsToEdit.length;
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Define Fields"
      content={
        <>
          <label className="block font-medium text-xl text-gray-500 px-1 mb-2">
            Fields
          </label>
          <div
            className="space-y-2 w-[450px] max-h-[325px] overflow-y-auto px-1 mb-2"
            ref={listRef}
          >
            {fields.map((field, index) => (
              <div key={field.id} className="flex items-center space-x-2">
                <TextInput
                  labelClassName="block font-medium text-gray-500 mb-1"
                  className="input input-bordered w-full"
                  wrapperClassName="mb-4"
                  label="Name"
                  name="name"
                  id="name"
                  placeholder="Field Name"
                  value={field.name}
                  onChange={(e) => {
                    const newFields = [...fields];
                    newFields[index].name = e.target.value;
                    setFields(newFields);
                  }}
                  required
                />
                <TextInput
                  labelClassName="block font-medium text-gray-500 mb-1"
                  className="input input-bordered w-full"
                  wrapperClassName="mb-4"
                  label="Key"
                  name="key"
                  id="key"
                  placeholder="Field Key"
                  value={field.key}
                  onChange={(e) => {
                    const newFields = [...fields];
                    newFields[index].key = e.target.value;
                    setFields(newFields);
                  }}
                  required
                  disabled={!keyAndNameEditable(field)}
                />

                <div className="w-40">
                  <label className="block font-medium text-gray-500 mb-1">
                    Type
                  </label>
                  <select
                    value={field.type}
                    onChange={(e) => {
                      const newFields = [...fields];
                      newFields[index].type = e.target.value as FieldType;
                      setFields(newFields);
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
                      const newFields = [...fields];
                      newFields.splice(index, 1);
                      setFields(newFields);
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
          <div className="flex justify-end">
            <Button
              onClick={() => {
                setFields([
                  ...fields,
                  {
                    id: v4(),
                    name: "",
                    key: "",
                    type: FieldType.TEXT,
                    isFullyEditable: true,
                  },
                ]);
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
        </>
      }
      actions={[
        {
          label: "Cancel",
          onClick: onClose,
          className: "bg-gray-500 text-white",
        },
        {
          label: "Create",
          onClick: () => {
            if (!validateFields()) {
              return;
            }
            onConfirm(fields);
            onClose();
          },
          className: "bg-blue-500 text-white",
          disabled: !fields.length || !fieldsValid || !fieldsUnique,
        },
      ]}
    />
  );
};
