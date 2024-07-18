import { FC, useMemo } from "react";
import { ScrapeGroup, ScrapeResult } from "../../types";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { getBaseUrl } from "../../util/url";
import { getColoredEndpointPill } from "../ColoredEndpointPill";

type ResultsTableProps = {
  group: ScrapeGroup | null;
  results: ScrapeResult[] | null;
  loading: boolean;
  hasMore: boolean;
  loadMore: () => void;
  height: number;
};

const columnHelper = createColumnHelper<{
  id: string;
  [key: string]: any;
}>();

const parseTime = (time: string) => {
  const date = new Date(time);
  return date.toLocaleString("en-US");
};

export const ResultsTable: FC<ResultsTableProps> = ({
  hasMore,
  loadMore,
  loading,
  results,
  group,
  height,
}) => {
  const renderResults = useMemo(() => {
    return (
      results?.map((result) => {
        return {
          ...result,
          fields: result.fields.filter((f) => {
            return group?.fields.some((gf) => gf.id === f.fieldId);
          }),
        };
      }) || []
    );
  }, [group?.fields, results]);

  const columns = useMemo(() => {
    if (!group) {
      return;
    }
    return [
      columnHelper.accessor("timestampInitial", {
        header: "Created",
        cell: (row) => parseTime(row.getValue()),
      }),
      columnHelper.accessor("timestampLastUpdated", {
        header: "Updated",
        cell: (row) => parseTime(row.getValue()),
      }),
      columnHelper.accessor("endpointId", {
        header: "Endpoint",
        cell: (row) => {
          return (
            <div className="py-2">
              {getColoredEndpointPill(row.row.original.endpointId, group)}
            </div>
          );
        },
      }),
      ...group.fields.map((field) =>
        columnHelper.accessor(field.name, {
          header: field.name,
          cell: (row) => {
            const value = row.row.original.fields.find(
              (result: any) => result.fieldId === field.id,
            );

            switch (field.type) {
              case "image": {
                let imageUrl = value?.value;
                const endpoint = group.endpoints.find(
                  (e) => e.id === row.row.original.endpointId,
                );
                if (imageUrl && imageUrl.startsWith("/")) {
                  const baseLink = getBaseUrl(endpoint?.url || "", true);
                  imageUrl = `${baseLink}${imageUrl}`;
                }
                return (
                  <div className="shrink-0 w-28">
                    <img
                      src={
                        imageUrl ||
                        "https://via.assets.so/img.jpg?w=135&h=100&tc=grey&bg=lightgrey&t=thumbnail"
                      }
                      alt={field.name}
                      loading="lazy"
                      style={{
                        width: "100%",
                        height: "100%",
                        objectFit: "contain",
                      }}
                    />
                  </div>
                );
              }
              case "link": {
                const endpoint = group.endpoints.find(
                  (e) => e.id === row.row.original.endpointId,
                );
                if (value?.value.startsWith("http")) {
                  return (
                    <a href={value?.value} target="_blank">
                      Link
                    </a>
                  );
                }
                const baseLink = getBaseUrl(endpoint?.url || "", true);

                return (
                  <a href={`${baseLink}${value?.value}`} target="_blank">
                    Link
                  </a>
                );
              }
              default:
                return value?.value;
            }
          },
        }),
      ),
    ];
  }, [group, renderResults]);

  const table = useReactTable({
    data: renderResults || [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  });
  return (
    <>
      <div className="text-lg pb-2 bg-base-100">
        <span>Displaying </span>
        <span className="font-bold">{table.getRowModel().rows.length}</span>
        <span> results</span>
      </div>
      <div
        className="overflow-x-auto w-100 flex flex-col  overflow-y-auto align-center no-scrollbar"
        style={{ height: `${height}px` }}
      >
        <table className="w-full border-collapse bg-gray-800 shadow-md rounded-lg">
          <thead className="bg-gray-700 sticky top-0">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    onClick={header.column.getToggleSortingHandler()}
                    className="px-4 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-600 transition whitespace-nowrap"
                  >
                    {flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="bg-gray-800 divide-y divide-gray-700">
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id} className="hover:bg-gray-700 transition">
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id} className="text-sm text-gray-300">
                    <div
                      className="px-4"
                      style={{
                        // max 3 lines of text
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        // if field type is text, max 3 lines of text
                        maxWidth: 300,
                        maxHeight: 100,
                        display: "-webkit-box",
                        WebkitLineClamp: 4,
                        WebkitBoxOrient: "vertical",
                      }}
                    >
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </div>
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        {loading && (
          <span className="loading loading-spinner loading-xs"></span>
        )}
        {hasMore && !loading && (
          <div className="w-100 flex justify-center">
            <button
              onClick={loadMore}
              disabled={loading}
              className="btn btn-primary my-3"
            >
              Load more
            </button>
          </div>
        )}
      </div>
    </>
  );
};
