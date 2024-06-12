import React, { FC, useEffect, useRef } from "react";
import { useClickOutside } from "../../hooks/useClickOutside";

export type ModalProps = {
  isOpen: boolean;
  onClose: () => void;
  actions: ModalAction[];
  title: string | React.ReactNode | null;
  content: string | React.ReactNode | null;
};

type ModalAction = {
  label: string;
  onClick: () => void;
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
};

export const Modal: FC<ModalProps> = ({
  actions,
  content,
  isOpen,
  onClose,
  title,
}) => {
  const dialogRef = useRef<HTMLDivElement>(null);
  const timeSinceOpened = useRef<number>(0);

  useEffect(() => {
    if (isOpen) {
      timeSinceOpened.current = Date.now();
    }

    if (!isOpen && timeSinceOpened.current > 0) {
      timeSinceOpened.current = 0;
    }

    return () => {
      timeSinceOpened.current = 0;
    };
  }, [isOpen]);

  useClickOutside(dialogRef, () => {
    if (
      isOpen &&
      timeSinceOpened.current > 0 &&
      Date.now() - timeSinceOpened.current > 100
    ) {
      onClose();
    }
  });

  const getModalTitle = () => {
    if (!title) {
      return null;
    }

    if (typeof title === "string") {
      return <h3 className="font-bold text-lg">{title}</h3>;
    }

    return title;
  };

  return (
    <dialog
      className="modal"
      open={isOpen}
      style={{
        backgroundColor: "rgba(0, 0, 0, 0.5)",
      }}
    >
      <div className="modal-box" ref={dialogRef}>
        <button
          className="btn btn-sm btn-circle btn-ghost absolute right-2 top-2"
          onClick={onClose}
        >
          ✕
        </button>
        <div className="border-b-2 border-secondary">{getModalTitle()}</div>
        <div className="p-4">{content}</div>
        <div className="modal-action gap-2">
          {actions.map((action, index) => (
            <button
              key={index}
              className={`btn ${action.className || ""}`}
              onClick={action.onClick}
              style={action.style}
              disabled={action.disabled}
            >
              {action.label}
            </button>
          ))}
        </div>
      </div>
    </dialog>
  );
};
