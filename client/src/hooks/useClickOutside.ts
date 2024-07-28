import { RefObject, useEffect } from "react";

export const useClickOutside = <T extends HTMLElement>(
	ref: RefObject<T>,
	callback: (event: MouseEvent) => void,
) => {
	useEffect(() => {
		const handleClick = (event: MouseEvent) => {
			if (ref.current && !ref.current.contains(event.target as Node)) {
				callback(event);
			}
		};

		document.addEventListener("click", handleClick);

		return () => {
			document.removeEventListener("click", handleClick);
		};
	}, [ref, callback]);
};
