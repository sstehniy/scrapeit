import { FC, MouseEventHandler } from "react";

type ButtonProps = React.InputHTMLAttributes<HTMLButtonElement>;
export const Button: FC<ButtonProps> = ({
  children,
  disabled,
  onClick,
  ...props
}) => {
  const disabledStyles: React.CSSProperties = disabled
    ? {
        cursor: "not-allowed",
        opacity: 0.5,
        pointerEvents: "none",
      }
    : {};

  const handleChange: MouseEventHandler<HTMLButtonElement> = (e) => {
    if (!disabled) {
      onClick && onClick(e);
    }
  };

  return (
    <div
      {...(props as React.InputHTMLAttributes<HTMLDivElement>)}
      onClick={
        handleChange as unknown as React.InputHTMLAttributes<HTMLDivElement>["onClick"]
      }
      style={{ ...props.style, ...disabledStyles }}
    >
      {children}
    </div>
  );
};
