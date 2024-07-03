import { FC } from "react";
import { ScrapeGroup } from "../../types";

export const GroupCard: FC<{ group: ScrapeGroup }> = ({ group }) => {
  return (
    <div
      key={group.id}
      className="bg-white shadow-md rounded p-4 flex flex-col"
    >
      <h2 className="text-xl font-bold">{group.name}</h2>
      <p className="text-gray-500">{group.url}</p>
    </div>
  );
};
