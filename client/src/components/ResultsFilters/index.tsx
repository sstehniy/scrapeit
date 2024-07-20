import { FC, useEffect, useState } from "react";
import { SearchConfig } from "../../views/GroupView";
import { Endpoint, ExportType, ScrapeGroup } from "../../types";
import { TextInput } from "../ui/TextInput";
import Select from "react-select";
import "./index.css";
import { getBgAndTextColor } from "../ColoredEndpointPill";
import axios from "axios";
import { ConfigureExportModal } from "../modals/ConfigureExport";
import { useMutation } from "@tanstack/react-query";
import { toast } from "react-toastify";

type ResultsFiltersProps = {
  params: SearchConfig;
  group: ScrapeGroup;
  setParams: (params: SearchConfig) => void;
};

export const ResultsFilters: FC<ResultsFiltersProps> = ({
  params,
  group,
  setParams,
}) => {
  const [q, setQ] = useState("");
  const [limit, setLimit] = useState(30);
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [showExportOptionsModal, setShowExportOptionsModal] = useState<{
    isOpen: boolean;
    onConfirm: ((name: string, type: ExportType) => void) | null;
  }>({
    isOpen: false,
    onConfirm: null,
  });

  useEffect(() => {
    const endpoints = group.endpoints.filter((e) =>
      params.endpointIds.includes(e.id),
    );
    const q = params.q;
    const limit = params.limit;
    setEndpoints(endpoints);
    setQ(q);
    setLimit(limit);
  }, [group.endpoints, params.endpointIds, params.limit, params.q]);

  const exportMutation = useMutation({
    mutationFn: (data: { name: string; type: ExportType }) =>
      axios
        .post(
          `/api/scrape/results/export/${group.id}`,
          {
            type: data.type,
            fileName: data.name,
            deleteAfterExport: false,
          },
          {
            responseType: "blob",
          },
        )
        .then((response) => {
          const url = window.URL.createObjectURL(new Blob([response.data]));
          const a = document.createElement("a");
          a.style.display = "none";
          a.href = url;
          a.download = `${data.name}.${data.type}`;

          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);

          window.URL.revokeObjectURL(url);
        }),
    onMutate: () => {
      setShowExportOptionsModal({ isOpen: false, onConfirm: null });
    },
    onSuccess: () => {
      toast.success("Exported successfully");
    },
    onError: () => {
      toast.error("Failed to export");
    },
    networkMode: "always",
  });

  return (
    <div className="flex justify-end items-end gap-4 mb-4">
      <TextInput
        labelClassName="label"
        className="input input-bordered flex items-center gap-2"
        wrapperClassName="form-control"
        label="Search"
        name="q"
        id="q"
        value={q}
        onChange={(e) => {
          setQ(e.target.value);
          setParams({ ...params, q: e.target.value });
        }}
        required
      />
      <TextInput
        labelClassName="label"
        className="input input-bordered flex items-center gap-2"
        wrapperClassName="form-control"
        label="Limit"
        name="limit"
        id="limit"
        type="number"
        value={limit}
        min={10}
        onChange={(e) => {
          setLimit(parseInt(e.target.value));
          setParams({ ...params, limit: parseInt(e.target.value) });
        }}
        required
      />
      <div className="form-control">
        <div className="label">
          <span className="label-text">Endpoints</span>
        </div>
        <Select
          isMulti
          isSearchable={false}
          options={group.endpoints.map((e) => ({
            value: e.id,
            label: e.name,
          }))}
          classNames={{
            container: () => "select-container",
            control: () => "select-control",
            valueContainer: () => "select-value-container",
            multiValue: () => "select-multivalue",
            menu: () => "select-menu",
          }}
          styles={{
            multiValue: (base, props) => {
              const { color, textColor } = getBgAndTextColor(props.data.label);
              return {
                ...base,
                backgroundColor: color,
                color: textColor,
                padding: "0rem 0.5rem",
                borderRadius: "0.9rem",
              };
            },
            multiValueLabel: (base, props) => {
              const { textColor } = getBgAndTextColor(props.data.label);
              return {
                ...base,
                color: textColor,
                fontWeight: "bold",
              };
            },
          }}
          value={endpoints.map((e) => ({
            value: e.id,
            label: e.name,
          }))}
          onChange={(selected) => {
            const ids = selected.map((s) => s.value);
            if (selected.length === 0) {
              setEndpoints([]);
              setParams({
                ...params,
                endpointIds: group.endpoints.map((e) => e.id),
              });
            } else {
              setEndpoints(
                selected.map(
                  (s) => group.endpoints.find((e) => e.id === s.value)!,
                ),
              );
              setParams({ ...params, endpointIds: ids });
            }
          }}
        />
      </div>
      <button
        className="btn btn-primary"
        onClick={() => {
          setShowExportOptionsModal({
            isOpen: true,
            onConfirm: (name, type) => {
              exportMutation.mutate({ name, type });
            },
          });
        }}
      >
        Export
      </button>
      {showExportOptionsModal.isOpen && (
        <ConfigureExportModal
          defaultName={group.name + "-" + new Date().toLocaleDateString()}
          isOpen={showExportOptionsModal.isOpen}
          onClose={() =>
            setShowExportOptionsModal({ isOpen: false, onConfirm: null })
          }
          onConfirm={(name, type) => {
            if (showExportOptionsModal.onConfirm) {
              showExportOptionsModal.onConfirm(name, type);
            }
          }}
        />
      )}
    </div>
  );
};
