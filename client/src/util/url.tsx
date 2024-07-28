export const getBaseUrl = (endpointFullUrl: string, withProtocol?: boolean) => {
	try {
		const url = new URL(endpointFullUrl);
		if (!withProtocol) return url.host;
		return `${url.protocol}//${url.host}`;
	} catch (_err) {
		return endpointFullUrl;
	}
};
