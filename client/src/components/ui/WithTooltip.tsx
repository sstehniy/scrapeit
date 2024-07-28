import { FC, PropsWithChildren } from "react";

export const WithTooltip: FC<
	PropsWithChildren<{
		tooltip: string;
	}>
> = ({ children, tooltip }) => {
	return tooltip ? (
		<div className="tooltip" data-tip={tooltip}>
			{children}
		</div>
	) : (
		<>{children}</>
	);
};
