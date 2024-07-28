import { ScrapeGroup } from "../types";

export const getBgAndTextColor = (name: string) => {
	const hash = name.split("").reduce((acc, char) => {
		return char.charCodeAt(0) + ((acc << 5) - acc);
	}, 0);
	const color = `hsl(${hash % 360}, 70%, 50%)`;

	// Determine text color based on background color brightness
	const getRGB = (c: string) => parseInt(c.slice(1), 16);
	const getsRGB = (c: number) => {
		c /= 255;
		return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
	};
	const getLuminance = (hexColor: string) => {
		const rgb = getRGB(hexColor);
		const r = getsRGB((rgb >> 16) & 0xff);
		const g = getsRGB((rgb >> 8) & 0xff);
		const b = getsRGB(rgb & 0xff);
		return 0.2126 * r + 0.7152 * g + 0.0722 * b;
	};
	const textColor = getLuminance(color) > 0.179 ? "#000000" : "#ffffff";
	return { color, textColor };
};

export const getColoredEndpointPill = (
	endpointId: string,
	group: ScrapeGroup,
) => {
	const endpoint = group.endpoints.find((e) => e.id === endpointId);
	const name = endpoint?.name || "Unknown";

	// Generate a color based on the name
	const { color, textColor } = getBgAndTextColor(name);
	return (
		<span
			className="px-2 py-1 rounded-xl text-sm font-medium text-nowrap"
			style={{
				backgroundColor: color,
				color: textColor,
			}}
		>
			{name}
		</span>
	);
};
