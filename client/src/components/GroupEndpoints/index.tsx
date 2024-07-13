import { FC, PropsWithChildren, useCallback, useState } from "react";
import { Endpoint, ScrapeGroup, SelectorStatus } from "../../types";
import { getColoredEndpointPill } from "../ColoredEndpointPill";
import { getBaseUrl } from "../../util/url";
import { ConfigureGroupEndpoint } from "../modals/ConfigureGroupEndpoint";
import axios from "axios";
import { toast } from "react-toastify";
import { Button } from "../ui/Button";

type GroupEndpointsProps = {
  group: ScrapeGroup;
  onEndpointChange: () => void | Promise<void>;
  allowScrapeAllEndpoints: boolean;
  disabledTooltip: string;
  onScrapeAllEndpoints: () => void | Promise<void>;
};

export const GroupEndpoints: FC<GroupEndpointsProps> = ({
  group,
  onEndpointChange,
  allowScrapeAllEndpoints,
  disabledTooltip,
  onScrapeAllEndpoints,
}) => {
  const { endpoints } = group;
  const [showCreateEndpointModal, setShowCreateEndpointModal] = useState(false);
  const [showEditEndpointModal, setShowEditEndpointModal] = useState<{
    isOpen: boolean;
    endpoint: Endpoint | null;
  }>({
    isOpen: false,
    endpoint: null,
  });

  const handleDeleteEndpoint = useCallback(
    async (endpointId: string) => {
      try {
        await axios.delete(
          `/api/scrape-groups/${group.id}/endpoints/${endpointId}`,
        );
        onEndpointChange();
      } catch (e) {
        toast.error("Failed to delete endpoint");
        console.error(e);
      }
    },
    [group.id, onEndpointChange],
  );

  const handleCreateEndpoint = useCallback(
    async (endpoint: Endpoint, type: "create" | "save") => {
      if (type === "create") {
        try {
          await axios.post(`/api/scrape-groups/${group.id}/endpoints`, {
            endpoint,
          });
        } catch (e) {
          toast.error("Failed to create endpoint");
          console.error(e);
        }
      } else {
        try {
          await axios.put(
            `/api/scrape-groups/${group.id}/endpoints/${endpoint.id}`,
            {
              endpoint,
            },
          );
          onEndpointChange();
        } catch (e) {
          toast.error("Failed to save endpoint");
          console.error(e);
        }
      }
      onEndpointChange();
    },
    [group.id, onEndpointChange],
  );

  return (
    <div className="mb-5">
      <h3 className="text-lg font-semibold mb-2">Endpoints</h3>
      {endpoints.length === 0 && (
        <p className="text-base-content">No endpoints configured yet</p>
      )}
      <div className="flex flex-wrap gap-4">
        {endpoints.map((endpoint) => (
          <div
            key={endpoint.id}
            className="mr-2 mb-2 px-3 py-2 border-2 rounded-lg border-base-content bg-base-100 text-base-content flex items-center"
          >
            <div className="mr-10 flex">
              <div className="w-24">
                {getColoredEndpointPill(endpoint.id, group)}
              </div>
              <div className="w-48">{getBaseUrl(endpoint.url, false)}</div>
            </div>
            <div className="flex gap-5">
              {endpoint.detailFieldSelectors.some(
                (ds) => ds.selectorStatus !== SelectorStatus.OK,
              ) && (
                <div
                  className="tooltip"
                  data-tip="This endpoint needs update since last schema change"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    className="size-6 text-warning"
                  >
                    <path
                      fillRule="evenodd"
                      d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003ZM12 8.25a.75.75 0 0 1 .75.75v3.75a.75.75 0 0 1-1.5 0V9a.75.75 0 0 1 .75-.75Zm0 8.25a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
              )}
              <Button
                onClick={() => {
                  setShowEditEndpointModal({
                    isOpen: true,
                    endpoint,
                  });
                }}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                  className="size-6 hover:text-primary"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z"
                  />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
                  />
                </svg>
              </Button>
              <Button>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                  className="size-6 hover:text-error"
                  onClick={() => handleDeleteEndpoint(endpoint.id)}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
                  />
                </svg>
              </Button>
            </div>
          </div>
        ))}
      </div>
      <div className="flex gap-4 mt-2">
        <Button
          className="btn btn-primary btn-sm"
          onClick={() => {
            setShowCreateEndpointModal(true);
          }}
          disabled={group.fields.length === 0}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={2.5}
            stroke="currentColor"
            className="size-4"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M12 4.5v15m7.5-7.5h-15"
            />
          </svg>
          New Endpoint
        </Button>
        {!!group.endpoints.length && (
          <WithTooltip
            tooltip={!allowScrapeAllEndpoints ? disabledTooltip : ""}
          >
            <Button
              className="btn btn-primary btn-sm"
              onClick={onScrapeAllEndpoints}
              disabled={!allowScrapeAllEndpoints}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={2.5}
                stroke="currentColor"
                className="size-4"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
                />
              </svg>
              Scrape All Endpoints
            </Button>
          </WithTooltip>
        )}
      </div>
      <ConfigureGroupEndpoint
        fields={group.fields}
        isOpen={showCreateEndpointModal}
        onClose={() => setShowCreateEndpointModal(false)}
        onConfirm={(endpoint) => handleCreateEndpoint(endpoint, "create")}
      />
      <ConfigureGroupEndpoint
        fields={group.fields}
        isOpen={showEditEndpointModal.isOpen}
        onClose={() =>
          setShowEditEndpointModal({
            isOpen: false,
            endpoint: null,
          })
        }
        onConfirm={(endpoint) => handleCreateEndpoint(endpoint, "save")}
        editEndpoint={showEditEndpointModal.endpoint ?? undefined}
      />
    </div>
  );
};

const WithTooltip: FC<
  PropsWithChildren<{
    tooltip: string;
  }>
> = ({ children, tooltip, ...props }) => {
  return tooltip ? (
    <div className="tooltip" data-tip={tooltip}>
      {children}
    </div>
  ) : (
    <>{children}</>
  );
};
