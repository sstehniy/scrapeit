import { useEffect, useState } from "react";

import axios from "axios";
import { ScrapeGroup } from "./types";

function App() {
  const [scrapeGroups, setScrapeGroups] = useState<ScrapeGroup[] | null>(null);
  useEffect(() => {
    axios.get(`/api/scrape-groups`).then((data) => {
      setScrapeGroups(data.data);
    });
  }, []);

  return (
    <div className="container mx-auto py-10">
      <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        Create new Group
      </button>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-5">
        {scrapeGroups?.map((group) => (
          <div
            key={group.id}
            className="bg-white shadow-md rounded p-4 flex flex-col"
          >
            <h2 className="text-xl font-bold">{group.name}</h2>
            <p className="text-gray-500">{group.url}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
