import { FC, useEffect, useState } from "react";
import { ScrapeGroup } from "../types";
import axios from "axios";
import { toast } from "react-toastify";
import { GroupCard } from "../components/GroupCard";
import { CreateGroupModal } from "../components/modals/CreateGroup";
import { useNavigate } from "react-router-dom";

export const GroupsOverview: FC = () => {
  const [scrapeGroups, setScrapeGroups] = useState<ScrapeGroup[] | null>(null);
  const [showCreateGroupModal, setShowCreateGroupModal] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    axios.get(`/api/scrape-groups`).then((data) => {
      setScrapeGroups(data.data);
    });
  }, []);

  const handleCreateGroup = (name: string) => {
    axios
      .post(`/api/scrape-groups`, { name })
      .then((response) => {
        toast.success("Group created");
        setShowCreateGroupModal(false);
        setScrapeGroups((prev) =>
          prev ? [...prev, response.data] : [response.data],
        );
        navigate(`/group/${response.data.id}`);
      })
      .catch((error) => {
        console.error(error);
        toast.error("Failed to create group");
      });
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Scrape Groups</h1>
        <button
          className="btn btn-primary"
          onClick={() => setShowCreateGroupModal(true)}
        >
          Create new Group
        </button>
      </div>

      {scrapeGroups === null ? (
        <div className="flex justify-center items-center h-64">
          <span className="loading loading-spinner loading-lg text-primary"></span>
        </div>
      ) : scrapeGroups.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-xl text-gray-600">
            No groups found. Create your first group!
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {scrapeGroups.map((group) => (
            <GroupCard key={group.id} group={group} />
          ))}
        </div>
      )}

      {showCreateGroupModal && (
        <CreateGroupModal
          isOpen={showCreateGroupModal}
          onClose={() => setShowCreateGroupModal(false)}
          onConfirm={handleCreateGroup}
        />
      )}
    </div>
  );
};
