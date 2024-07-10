import { FC } from "react";
import { ScrapeGroup } from "../../types";
import { NavLink } from "react-router-dom";
import { getBaseUrl } from "../../util/url";

export const GroupCard: FC<{ group: ScrapeGroup }> = ({ group }) => {
  return (
    <div
      key={group.id}
      className="bg-white shadow-md rounded p-4 flex flex-col"
    >
      <NavLink to={`/group/${group.id}`}>
        <h2 className="text-xl font-bold text-base-300">{group.name}</h2>
        {group.endpoints.map((endpoint) => (
          <div key={endpoint.id} className="mt-2">
            <h3 className="text-lg font-bold inline-block me-3">
              {endpoint.name}
            </h3>
            <span>({getBaseUrl(endpoint.url)})</span>
          </div>
        ))}
      </NavLink>
    </div>
  );
};
