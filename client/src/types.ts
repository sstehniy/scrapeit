type ScrapeGroup = {
	id: string;
	originalId?: string;
	name: string;
	fields: Field[];
	endpoints: Endpoint[];
	withThumbnail: boolean;
	versionTag: string;
	created: string;
	updated: string;
	isArchived: boolean;
};

type ArchivedScrapeGroup = {
	id: string;
	originalId: string;
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
	NUMBER = "number",
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
	withDetailedView: boolean;
	detailedViewTriggerSelector: string;
	detailedViewMainElementSelector: string;
	interval?: string;
	active?: boolean;
	lastScraped?: Date;
	status?: ScrapeStatus;
};

type PaginationConfig = {
	type: "url_parameter" | "url_path";
	parameter: string;
	start: number;
	end: number;
	step: number;
	urlRegexToInsert?: string;
};

enum ScrapeStatus {
	RUNNING = "running",
	IDLE = "idle",
}

enum SelectorStatus {
	OK = "ok",
	NEEDS_UPDATE = "needs_update",
	NEW = "new",
}

enum ExportType {
	EXCEL = "xlsx",
	CSV = "csv",
	XML = "xml",
	JSON = "json",
}

type FieldSelector = {
	id: string;
	fieldId: string;
	selector: string;
	attributeToGet: string;
	regex: string;
	regexMatchIndexToUse: number;
	selectorStatus: SelectorStatus;
	lockedForEdit: boolean;
};

type ScrapeResult = {
	id: string;
	uniqueHash: string;
	endpointId: string;
	groupId: string;
	fields: ScrapeResultDetail[];
	timestampLastUpdate: Date;
	TimestampInitial: Date;
	groupVersionTag: string;
};

type ScrapeResultTest = ScrapeResult & {
	fields: ScrapeResultDetailTest[];
};

type ScrapeResultDetail = {
	id: string;
	fieldId: string;
	value: string;
};

type ScrapeResultDetailTest = ScrapeResultDetail & {
	rawData: string;
	regexMatches: string[];
};

type NotificationConfig = {
	id: string;
	groupId: string;
	name: string;
	fieldIdsToNotify: string[];
	conditions: NotificationCondition[];
};

type NotificationCondition = {
	fieldId: string;
	operator: string;
	value: number;
};

enum ScrapeType {
	PREVIEWS = "previews",
	PREVIEWS_WITH_DETAILS = "previews_with_details",
	PURE_DETAILS = "pure_details",
}

export { FieldType, SelectorStatus, ScrapeStatus, ExportType, ScrapeType };

export type {
	NotificationConfig,
	NotificationCondition,
	ScrapeGroup,
	ArchivedScrapeGroup,
	Field,
	Endpoint,
	PaginationConfig,
	FieldSelector,
	ScrapeResult,
	ScrapeResultTest,
	ScrapeResultDetail,
	ScrapeResultDetailTest,
};
