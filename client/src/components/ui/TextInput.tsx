import { FC } from "react";

type TextInputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  labelClassName?: string;
  labelStyle?: React.CSSProperties;
  wrapperClassName?: string;
  wrapperStyle?: React.CSSProperties;
  error?: string;
};

export const TextInput: FC<TextInputProps> = ({
  label,
  labelClassName,
  labelStyle,
  wrapperClassName,
  wrapperStyle,
  ...props
}) => {
  return (
    <div className={wrapperClassName} style={wrapperStyle}>
      <label className={labelClassName} style={labelStyle}>
        {label}
      </label>
      <input
        type={props.type === undefined ? "text" : props.type}
        className={props.className}
        value={
          props.value === undefined || props.value === null ? "" : props.value
        }
        onChange={props.onChange === undefined ? () => {} : props.onChange}
        {...props}
      />
      {props.error && (
        <p className="text-red-500 text-xs italic mt-2">{props.error}</p>
      )}
    </div>
  );
};
