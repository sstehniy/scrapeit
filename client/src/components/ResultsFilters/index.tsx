import { FC, useEffect, useState } from "react";
import { SearchConfig } from "../../views/GroupView";
import { Endpoint, ScrapeGroup } from "../../types";
import { TextInput } from "../ui/TextInput";
import Select from "react-select";
import "./index.css";
import { getBgAndTextColor } from "../ColoredEndpointPill";

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

  return (
    <div className="flex justify-end gap-4">
      <TextInput
        labelClassName="label"
        className="input input-bordered w-full"
        wrapperClassName="form-control mb-4"
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
        className="input input-bordered w-full"
        wrapperClassName="form-control mb-4"
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
            container: (_) => "",
            control: (_) => "select-control",
            valueContainer: (_) => "select-value-container",
            multiValue: (_) => "select-multivalue",
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
    </div>
  );
};
