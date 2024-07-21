import { FC, useEffect, useState } from "react";
import { SearchConfig, SearchFilter, SearchSort } from "../../views/GroupView";
import {
  Endpoint,
  ExportType,
  Field,
  FieldType,
  ScrapeGroup,
} from "../../types";
import { TextInput } from "../ui/TextInput";
import "./index.css";
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
  const [filters, setFilters] = useState<SearchFilter[]>([]);
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

  const filterableNumberFields = group.fields.filter(
    (f) => f.type === FieldType.NUMBER,
  );
  const filterableStringFields = group.fields.filter(
    (f) => f.type === FieldType.TEXT,
  );

  const numberOperators = ["=", "!=", ">", "<"];
  const stringOperators = ["=", "!="];

  useEffect(() => {
    const appliableFilters = filters.filter(
      (f) => f.value !== null && f.value !== "" && f.fieldId !== "",
    );

    const equal =
      JSON.stringify(params.filters) === JSON.stringify(appliableFilters);
    if (equal) {
      return;
    }
    console.log("setting filters", appliableFilters);
    setParams({ ...params, filters: appliableFilters });
  }, [filters, params, setParams]);

  const sortOptions: SearchSort[] = filterableNumberFields.reduce(
    (acc, curr: Field) => {
      return [
        ...acc,
        { fieldId: curr.id, order: -1 },
        { fieldId: curr.id, order: 1 },
      ] as SearchSort[];
    },
    [] as SearchSort[],
  );

  return (
    <div>
      <div className="flex items-end gap-4 mb-4 flex-wrap">
        <TextInput
          labelClassName="label"
          className="input input-sm input-bordered flex items-center gap-2 w-56"
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
          className="input input-sm input-bordered flex items-center gap-2 w-24"
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
        <select
          className="select select-bordered select-sm w-32"
          value={params.sort.fieldId + ":" + params.sort.order}
          onChange={(e) => {
            const [fieldId, order] = e.target.value.split(":");
            console.log({ fieldId, order });
            setParams({
              ...params,
              sort: {
                fieldId,
                order: parseInt(order) as SearchSort["order"],
              },
            });
          }}
        >
          <option value=":1">Sort By</option>
          {sortOptions.map((sort) => (
            <option
              key={sort.fieldId + sort.order}
              value={`${sort.fieldId}:${sort.order}`}
              selected={
                sort.fieldId === params.sort.fieldId &&
                sort.order === params.sort.order
              }
            >
              {group.fields.find((f) => f.id === sort.fieldId)?.name}:{" "}
              {sort.order === 1 ? "ascending" : "descending"}
            </option>
          ))}
        </select>
        <button
          className="btn btn-primary btn-outline btn-sm"
          onClick={() => {
            setShowExportOptionsModal({
              isOpen: true,
              onConfirm: (name, type) => {
                exportMutation.mutate({ name, type });
              },
            });
          }}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={1.5}
            stroke="currentColor"
            className="size-5 mt-[-3px]"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M9 8.25H7.5a2.25 2.25 0 0 0-2.25 2.25v9a2.25 2.25 0 0 0 2.25 2.25h9a2.25 2.25 0 0 0 2.25-2.25v-9a2.25 2.25 0 0 0-2.25-2.25H15m0-3-3-3m0 0-3 3m3-3V15"
            />
          </svg>
          Export
        </button>
        <button
          className="btn btn-success btn-sm btn-outline"
          onClick={() => {
            setFilters([
              ...filters,
              { fieldId: "", value: null, operator: "=" },
            ]);
          }}
        >
          + Filter
        </button>
      </div>
      <div className="flex flex-wrap gap-3 mb-4">
        {filters.map((filter, index) => {
          const fieldType = group.fields.find(
            (f) => f.id === filter.fieldId,
          )?.type;
          return (
            <div
              key={index}
              className="flex items-center gap-3 flex-wrap p-3 border border-gray-200 border-opacity-30 rounded-md shadow-sm"
            >
              <select
                className="select select-bordered select-sm w-32"
                value={filter.fieldId}
                onChange={(e) => {
                  const fieldId = e.target.value;
                  setFilters((prev) =>
                    prev.map((f, i) =>
                      index === i
                        ? ({
                            value: null,
                            operator: "=",
                            fieldId,
                          } as SearchFilter)
                        : f,
                    ),
                  );
                }}
              >
                <option value="">Select Field</option>
                {[...filterableNumberFields, ...filterableStringFields].map(
                  (field) => (
                    <option key={field.id} value={field.id}>
                      {field.name}
                    </option>
                  ),
                )}
              </select>
              <select
                className="select select-bordered select-sm w-16"
                value={filter.operator}
                onChange={(e) => {
                  const operator = e.target.value as SearchFilter["operator"];
                  setFilters((prev) =>
                    prev.map((f, i) => (i === index ? { ...f, operator } : f)),
                  );
                }}
              >
                {fieldType === FieldType.NUMBER ? (
                  <>
                    {numberOperators.map((op) => (
                      <option key={op} value={op}>
                        {op}
                      </option>
                    ))}
                  </>
                ) : (
                  <>
                    {stringOperators.map((op) => (
                      <option key={op} value={op}>
                        {op}
                      </option>
                    ))}
                  </>
                )}
              </select>
              <TextInput
                labelClassName=""
                className="input input-sm input-bordered w-32"
                wrapperClassName="form-control"
                label=""
                name="value"
                id="value"
                type={fieldType === FieldType.NUMBER ? "number" : "text"}
                value={filter.value ?? ""}
                onChange={(e) => {
                  let value: string | number = e.target.value;
                  if (fieldType === FieldType.NUMBER) {
                    console.log("here");
                    value = parseFloat(value);
                  }
                  setFilters((prev) =>
                    prev.map((f, i) => (i === index ? { ...f, value } : f)),
                  );
                }}
                required
              />
              <button
                onClick={() => {
                  setFilters((prev) => prev.filter((_, i) => i !== index));
                }}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                  className="size-5"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M6 18 18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
          );
        })}
      </div>
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
