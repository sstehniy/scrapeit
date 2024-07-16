/** eslint-disable prettier/prettier */
/** eslint-disable prettier/prettier */
/** eslint-disable prettier/prettier */
/** eslint-disable prettier/prettier */
/** eslint-disable prettier/prettier */
/** eslint-disable prettier/prettier */
/** eslint-disable prettier/prettier */
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { useState } from "react";
import { toast } from "react-toastify";
import { Button } from "../ui/Button";
import { MultilineTextInput } from "../ui/MultilineTextInput";

export const FloatingAIChat = () => {
  const [isChatOpen, setIsChatOpen] = useState(false);

  const handleChatClick = () => {
    setIsChatOpen(!isChatOpen);
  };

  return (
    <>
      <ChatWindow // eslint-disable-next-line prettier/prettier
        isOpen={isChatOpen}
        onClose={() => {
          setIsChatOpen(false);
        }}
      />

      {!isChatOpen && <FloatingButton onClick={handleChatClick} />}
    </>
  );
};

const FloatingButton = ({ onClick }: { onClick: () => void }) => {
  return (
    <div
      onClick={onClick}
      className="fixed bottom-0 right-0 p-3 m-3 bg-blue-500 text-white rounded-full cursor-pointer z-[100000]"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="size-5"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z"
        />
      </svg>
    </div>
  );
};

const ChatWindow = ({
  onClose,
  isOpen,
}: {
  onClose: () => void;
  isOpen: boolean;
}) => {
  const [prompt, setPrompt] = useState("");
  const [messages, setMessages] = useState<string[]>([]);

  const { mutate } = useMutation<{ data: { response: string } }>({
    mutationKey: ["ai-completion", prompt],
    mutationFn: () => {
      return axios.post("/api/ai/completion", { prompt });
    },
    onSuccess: (data) => {
      console.log(data);
      setMessages([data.data.response]);
    },
    onError: (err) => {
      console.log(err);
      toast.error("Failed to get response from AI");
      setPrompt("");
    },
  });

  return (
    <div
      className="fixed bottom-0 right-0 m-4 bg-base-300 shadow-xl  rounded-lg z-[100000] px-5 py-5 w-[600px]"
      style={{
        display: isOpen ? "block" : "none",
      }}
    >
      <div className="absolute top-4 right-4 cursor-pointer" onClick={onClose}>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
          stroke="currentColor"
          className="size-4"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M6 18 18 6M6 6l12 12"
          />
        </svg>
      </div>
      <div className="px-3">
        <MultilineTextInput
          labelClassName="label"
          className="textarea textarea-bordered w-full text-lg"
          wrapperClassName="form-control mb-4"
          label="Prompt"
          name="prompt"
          id="prompt"
          value={prompt}
          onChange={(e) => {
            setPrompt(e.target.value);
          }}
        />

        <Button
          className="btn btn-primary w-full"
          onClick={() => {
            mutate();
          }}
          disabled={!prompt.trim()}
        >
          Ask AI
        </Button>
      </div>

      <div className="mt-4">
        <div
          className="chat chat-start"
          style={{
            letterSpacing: "1.25px",
          }}
        >
          {messages.map((message, index) => (
            <div key={index} className="chat-bubble">
              {message}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
