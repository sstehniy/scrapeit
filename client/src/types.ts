export type ScrapeGroup = {
  id: string;
  name: string;
  url: string;
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
  searchConfig: SearchConfig[];
  paginationConfig: PaginationConfig;
  mainElementSelector: string;
  detailFieldSelectors: FieldSelector[];
  interval?: number;
  active?: boolean;
  lastScraped?: Date;
  status?: ScrapeStatus;
};

type PaginationConfig =
  | {
      type: "url";
      parameter: string;
      start: number;
      end: number;
    }
  | {
      type: "selectors";
      nextSelector: string;
      prevSelector: string;
      maxPages: number;
    };

enum ScrapeStatus {
  SUCCESS = "success",
  FAILED = "failed",
}

type FieldSelector = {
  fieldId: string;
  selector: string;
  // e.g href, src, innerText, data-*
  attributeToGet: string;
};

type SearchConfig = {
  id: string;
  name: string;
  param: string;
  value: string;
};
