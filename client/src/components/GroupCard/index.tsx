import { FC } from "react";
import { ScrapeGroup } from "../../types";
import { NavLink } from "react-router-dom";
import { getBaseUrl } from "../../util/url";

export const GroupCard: FC<{ group: ScrapeGroup }> = ({ group }) => {
  return (
    <div className="card bg-base-300 shadow-xl hover:shadow-2xl transition-shadow duration-300">
      <NavLink to={`/group/${group.id}`} className="card-body flex flex-col">
        <h2 className="card-title text-2xl">{group.name}</h2>
        <div className="flex-1 ps-5">
          {group.endpoints.map((endpoint) => (
            <div key={endpoint.id} className="mt-1">
              <h3 className="font-bold inline-block mr-2">{endpoint.name}</h3>
              <span className="text-sm text-gray-500">
                ({getBaseUrl(endpoint.url)})
              </span>
            </div>
          ))}
        </div>
      </NavLink>
    </div>
  );
};
