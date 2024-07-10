type ScrapeGroup = {
  id: string;
  name: string;
  fields: Field[];
  endpoints: Endpoint[];
  withThumbnail: boolean;
};

type Field = {
  id: string;
  name: string;
  key: string;
  type: FieldType;
  isFullyEditable: boolean;
};

enum FieldType {
  TEXT = "text",
  IMAGE = "image",
  LINK = "link",
}

type Endpoint = {
  id: string;
  name: string;
  url: string;
  paginationConfig: PaginationConfig;
  mainElementSelector: string;
  detailFieldSelectors: FieldSelector[];
  interval?: number;
  active?: boolean;
  lastScraped?: Date;
  status?: ScrapeStatus;
};

type PaginationConfig = {
  type: "url_parameter";
  parameter: string;
  start: number;
  end: number;
  step: number;
  urlRegexToInsert?: string;
};

enum ScrapeStatus {
  SUCCESS = "success",
  FAILED = "failed",
}

type FieldSelector = {
  id: string;
  fieldId: string;
  selector: string;
  attributeToGet: string;
  regex: string;
};

type ScrapeResult = {
  id: string;
  uniqueHash: string;
  endpointId: string;
  groupId: string;
  fields: ScrapeResultDetail[];
  timestamp: Date;
};

type ScrapeResultDetail = {
  id: string;
  fieldId: string;
  value: string;
};

export { FieldType };

export type {
  ScrapeGroup,
  Field,
  Endpoint,
  PaginationConfig,
  ScrapeStatus,
  FieldSelector,
  ScrapeResult,
  ScrapeResultDetail,
};
