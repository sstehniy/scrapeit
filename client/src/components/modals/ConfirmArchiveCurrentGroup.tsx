import { FC, useCallback, useEffect, useState } from "react";
import { Modal, ModalProps } from "../ui/Modal";
import { TextInput } from "../ui/TextInput";
import {
  keepPreviousData,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import axios from "axios";

type ConfirmArchiveCurrentGroupProps = Pick<
  ModalProps,
  "isOpen" | "onClose"
> & {
  onConfirm: (versionTag: string) => void;
};

export const ConfirmArchiveCurrentGroup: FC<
  ConfirmArchiveCurrentGroupProps
> = ({ onConfirm, isOpen, onClose }) => {
  const [versionTag, setVersionTag] = useState("");
  const [versionTagError, setVersionTagError] = useState("");

  const validateVersionTag = useCallback((tag: string) => {
    return tag.trim().length > 0;
  }, []);

  const queryClient = useQueryClient();

  const { data } = useQuery<{ data: { exists: boolean } }>({
    queryKey: ["versionTagExists", versionTag],
    queryFn: async () => {
      return axios.get(`/api/scrape-groups/version-tag-exists/${versionTag}`);
    },
    enabled: validateVersionTag(versionTag),
    placeholderData: keepPreviousData,
  });

  const versionExists = data?.data.exists;

  useEffect(() => {
    if (isOpen) {
      setVersionTag("");
    }
  }, [isOpen]);

  console.log(versionExists);

  const getConfirmArchiveCurrentGroupContent = () => {
    return (
      <div className="space-y-2 w-[450px]">
        <p className="text-gray-500 mb-3">
          Warning: as you have added new fields to current group that already
          has scraped data, this group will be archived. You still can access
          scraped data for this version of group, however, you will only be able
          to scrape new data with the updated group. Please enter a version tag
          for the new ARCHIVED group.
        </p>
        <TextInput
          labelClassName="block font-medium text-gray-500 mb-1"
          className="input input-bordered w-full"
          wrapperClassName="mb-4"
          label="Version Tag"
          name="version_tag"
          id="version_tag"
          placeholder="e.g v1.0.0"
          value={versionTag}
          onChange={(e) => {
            setVersionTag(e.target.value);
            if (!validateVersionTag(versionTag)) {
              setVersionTagError("Version Tag is required");
            } else {
              setVersionTagError("");
            }
            queryClient.invalidateQueries({
              queryKey: ["versionTagExists", versionTag],
            });
          }}
          required
        />
        {versionExists && (
          <p className="text-red-500 text-sm mb-2">
            Provided Version Tag exists already
          </p>
        )}
        {versionTagError && (
          <p className="text-red-500 text-sm">{versionTagError}</p>
        )}
      </div>
    );
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create new Group"
      content={getConfirmArchiveCurrentGroupContent()}
      actions={[
        {
          label: "Cancel",
          onClick: onClose,
          className: "bg-gray-500 text-white",
        },
        {
          label: "Create",
          onClick: () => {
            if (!validateVersionTag(versionTag)) {
              setVersionTagError("Version Tag is required");
              return;
            }
            if (versionExists) {
              return;
            }
            onConfirm(versionTag);
          },
          className: "bg-blue-500 text-white",
          disabled: versionExists || !validateVersionTag(versionTag),
        },
      ]}
    />
  );
};
