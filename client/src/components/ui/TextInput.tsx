import { FC } from "react";
import { WithTooltip } from "./WithTooltip";

type TextInputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  labelClassName?: string;
  labelStyle?: React.CSSProperties;
  wrapperClassName?: string;
  wrapperStyle?: React.CSSProperties;
  tooltip?: string;
  error?: string;
};

export const TextInput: FC<TextInputProps> = ({
  label,
  labelClassName,
  labelStyle,
  wrapperClassName,
  wrapperStyle,
  tooltip,
  ...props
}) => {
  return (
    <div className={wrapperClassName} style={wrapperStyle}>
      <label className={labelClassName} style={labelStyle}>
        <span className="label-text">{label}</span>
      </label>
      <label className={props.className}>
        <input
          type={props.type === undefined ? "text" : props.type}
          value={
            props.value === undefined || props.value === null ? "" : props.value
          }
          onChange={props.onChange === undefined ? () => {} : props.onChange}
          {...props}
          className="grow w-full"
        />
        {tooltip && (
          <WithTooltip tooltip={tooltip}>
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
                d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"
              />
            </svg>
          </WithTooltip>
        )}
      </label>
      {props.error && (
        <p className="text-red-500 text-xs italic mt-2">{props.error}</p>
      )}
    </div>
  );
};
