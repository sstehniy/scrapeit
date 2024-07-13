type ScrapeGroup = {
  id: string;
  name: string;
  fields: Field[];
  endpoints: Endpoint[];
  withThumbnail: boolean;
  versionTag: string;
  created: string;
  updated: string;
  isArchived: boolean;
};

type Field = {
  id: string;
  name: string;
  key: string;
  type: FieldType;
  isFullyEditable: boolean;
  order: number;
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

enum SelectorStatus {
  OK = "ok",
  NEEDS_UPDATE = "needs_update",
  NEW = "new",
}

type FieldSelector = {
  id: string;
  fieldId: string;
  selector: string;
  attributeToGet: string;
  regex: string;
  selectorStatus: SelectorStatus;
};

type ScrapeResult = {
  id: string;
  uniqueHash: string;
  endpointId: string;
  groupId: string;
  fields: ScrapeResultDetail[];
  timestamp: Date;
  groupVersionTag: string;
};

type ScrapeResultDetail = {
  id: string;
  fieldId: string;
  value: string;
};

export { FieldType, SelectorStatus };

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
