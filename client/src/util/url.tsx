export const getBaseUrl = (endpointFullUrl: string, withProtocol?: boolean) => {
  const url = new URL(endpointFullUrl);
  if (!withProtocol) return url.host;
  return `${url.protocol}//${url.host}`;
};
