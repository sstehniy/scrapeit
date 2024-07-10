import { FC, useEffect, useState } from "react";
import { ScrapeGroup } from "../types";
import axios from "axios";
import { toast } from "react-toastify";
import { GroupCard } from "../components/GroupCard";
import { CreateGroupModal } from "../components/modals/CreateGroup";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";

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
    <div>
      <Button
        className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
        onClick={() => setShowCreateGroupModal(true)}
      >
        Create new Group
      </Button>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-5">
        {scrapeGroups?.map((group) => (
          <GroupCard key={group.id} group={group} />
        ))}
      </div>
      <CreateGroupModal
        isOpen={showCreateGroupModal}
        onClose={() => setShowCreateGroupModal(false)}
        onConfirm={handleCreateGroup}
      />
    </div>
  );
};
