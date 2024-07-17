import React, { FC, PropsWithChildren, useRef } from "react";
import { createPortal } from "react-dom";
import { useClickAway } from "react-use";
import { Button } from "./Button";

export type ModalProps = {
  isOpen: boolean;
  onClose: () => void;
  actions: ModalAction[];
  title: string | React.ReactNode | null;
  closeOnClickOutside?: boolean;
};

export type ModalAction = {
  label: string;
  onClick: () => void;
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
};

export const Modal: FC<PropsWithChildren<ModalProps>> = ({
  actions,
  isOpen,
  children,
  onClose,
  title,
  closeOnClickOutside = true,
}) => {
  const modalContentRef = useRef<HTMLDivElement>(null);
  useClickAway(modalContentRef, closeOnClickOutside ? onClose : () => {});
  const getModalTitle = () => {
    if (!title) {
      return null;
    }

    if (typeof title === "string") {
      return <h3 className="font-bold text-xl mb-2">{title}</h3>;
    }

    return title;
  };

  if (!isOpen) return null;

  const portalRoot = document.getElementById("portal");
  if (!portalRoot) return null;

  return createPortal(
    <div
      className="modal modal-open"
      style={{
        backgroundColor: "rgba(0, 0, 0, 0.5)",
        zIndex: 1000,
      }}
    >
      <div
        className="modal-box max-w-max px-14 overflow-x-hidden"
        ref={modalContentRef}
      >
        <Button
          className="btn btn-sm btn-circle btn-ghost absolute right-2 top-2"
          onClick={onClose}
        >
          ✕
        </Button>
        {getModalTitle()}
        <div className="py-4">{children}</div>
        <div className="modal-action gap-2">
          {actions.map((action, index) => (
            <Button
              key={index}
              className={`btn ${action.className || ""}`}
              onClick={action.onClick}
              style={action.style}
              disabled={action.disabled}
            >
              {action.label}
            </Button>
          ))}
        </div>
      </div>
    </div>,
    portalRoot,
  );
};
