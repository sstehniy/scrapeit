import { useEffect, useState } from "react";

import axios from "axios";
import { ScrapeGroup } from "./types";
import { CreateGroupModal } from "./components/modals/CreateGroup";
import { toast } from "react-toastify";

function App() {
  const [scrapeGroups, setScrapeGroups] = useState<ScrapeGroup[] | null>(null);
  const [showCreateGroupModal, setShowCreateGroupModal] = useState(false);
  useEffect(() => {
    axios.get(`/api/scrape-groups`).then((data) => {
      setScrapeGroups(data.data);
    });
  }, []);

  const handleCreateGroup = (name: string) => {
    axios
      .post(`/api/scrape-groups`, { name })
      .then((response) => {
        // setScrapeGroups((prev) => (prev ? [...prev, data.data] : [data.data]));
        toast.success("Group created");
        setShowCreateGroupModal(false);
        setScrapeGroups((prev) =>
          prev ? [...prev, response.data] : [response.data],
        );
      })
      .catch((error) => {
        console.error(error);
        toast.error("Failed to create group");
      });
  };

  return (
    <div className="container mx-auto py-10">
      <button
        className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
        onClick={() => setShowCreateGroupModal(true)}
      >
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
      <CreateGroupModal
        isOpen={showCreateGroupModal}
        onClose={() => setShowCreateGroupModal(false)}
        onConfirm={handleCreateGroup}
      />
    </div>
  );
}

export default App;
