import { FC, useRef } from "react";

type MultilineTextInputProps =
	React.TextareaHTMLAttributes<HTMLTextAreaElement> & {
		label: string;
		labelClassName?: string;
		labelStyle?: React.CSSProperties;
		wrapperClassName?: string;
		wrapperStyle?: React.CSSProperties;
		error?: string;
	};

export const MultilineTextInput: FC<MultilineTextInputProps> = ({
	label,
	labelClassName,
	labelStyle,
	wrapperClassName,
	wrapperStyle,
	...props
}) => {
	const ref = useRef<HTMLTextAreaElement>(null);

	const autoGrow = (element: HTMLTextAreaElement) => {
		element.style.height = "5px";
		element.style.height = element.scrollHeight + "px";
	};

	// eslint-disable-next-line prettier/prettier
	return (
		<div className={wrapperClassName} style={wrapperStyle}>
			<label className={labelClassName} style={labelStyle}>
				{label}
			</label>
			<textarea
				className={props.className}
				value={
					props.value === undefined || props.value === null ? "" : props.value
				}
				ref={ref}
				onChange={props.onChange === undefined ? () => {} : props.onChange}
				onInput={(e) => autoGrow(e.currentTarget)}
				{...props}
				style={{
					resize: "none",
					overflow: "hidden",
					maxHeight: "200px",
					...props.style,
				}}
			/>
			{props.error && (
				<p className="text-red-500 text-xs italic mt-2">{props.error}</p>
			)}
		</div>
	);
};
