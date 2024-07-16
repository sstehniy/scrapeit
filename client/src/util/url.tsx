export const getBaseUrl = (endpointFullUrl: string, withProtocol?: boolean) => {
  console.log(endpointFullUrl);

  try {
    const url = new URL(endpointFullUrl);
    if (!withProtocol) return url.host;
    return `${url.protocol}//${url.host}`;
  } catch (err) {
    return endpointFullUrl;
  }
};
