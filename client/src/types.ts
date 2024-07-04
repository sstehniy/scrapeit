export type ScrapeGroup = {
  id: string;
  name: string;
  fields: Field[];
  endpoints: Endpoint[];
};

type Field = {
  id: string;
  name: string;
  type: FieldType;
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
};

enum ScrapeStatus {
  SUCCESS = "success",
  FAILED = "failed",
}

type FieldSelector = {
  id: string;
  fieldId: string;
  selector: string;
  // e.g href, src, innerText, data-*
  attributeToGet: string;
};
