export type IDWhereCondition = {
  equals?: string | null;
  notEquals?: string | null;
  oneOf?: string[] | null;
};

export type StringWhereCondition = {
  startsWith?: string | null;
  endsWith?: string | null;
  oneOf?: string[] | null;
  contains?: string | null;
  equals?: string | null;
  notEquals?: string | null;
};

export type BooleanWhereCondition = {
  equals?: boolean | null;
  notEquals?: boolean | null;
};

export type NumberWhereCondition = {
  greaterThan?: number | null;
  greaterThanOrEquals?: number | null;
  lessThan?: number | null;
  lessThanOrEquals?: number | null;
  equals?: number | null;
  notEquals?: number | null;
};

export type DurationWhereCondition = {
  greaterThan?: DurationString | null;
  greaterThanOrEquals?: DurationString | null;
  lessThan?: DurationString | null;
  lessThanOrEquals?: DurationString | null;
  equals?: DurationString | null;
  notEquals?: DurationString | null;
};

export type DateWhereCondition = {
  equals?: Date | string | null;
  equalsRelative?: RelativeDateString | null;
  before?: Date | string | null;
  beforeRelative?: RelativeDateString | null;
  onOrBefore?: Date | string | null;
  after?: Date | string | null;
  afterRelative?: RelativeDateString | null;
  onOrAfter?: Date | string | null;
};

export type DateQueryInput = {
  equals?: string | null;
  before?: string | null;
  onOrBefore?: string | null;
  after?: string | null;
  onOrAfter?: string | null;
};

export type TimestampQueryInput = {
  before: string | null;
  after: string | null;
  equalsRelative?: RelativeDateString | null;
  beforeRelative?: RelativeDateString | null;
  afterRelative?: RelativeDateString | null;
};

export type StringArrayWhereCondition = {
  equals?: string[] | null;
  notEquals?: string[] | null;
  any?: StringArrayQueryWhereCondition | null;
  all?: StringArrayQueryWhereCondition | null;
};

export type StringArrayQueryWhereCondition = {
  equals?: string | null;
  notEquals?: string | null;
};

export type NumberArrayWhereCondition = {
  equals?: number[] | null;
  notEquals?: number[] | null;
  any?: NumberArrayQueryWhereCondition | null;
  all?: NumberArrayQueryWhereCondition | null;
};

export type NumberArrayQueryWhereCondition = {
  greaterThan?: number | null;
  greaterThanOrEquals?: number | null;
  lessThan?: number | null;
  lessThanOrEquals?: number | null;
  equals?: number | null;
  notEquals?: number | null;
};

export type BooleanArrayWhereCondition = {
  equals?: boolean[] | null;
  notEquals?: boolean[] | null;
  any?: BooleanArrayQueryWhereCondition | null;
  all?: BooleanArrayQueryWhereCondition | null;
};

export type BooleanArrayQueryWhereCondition = {
  equals?: boolean | null;
  notEquals?: boolean | null;
};

export type DateArrayWhereCondition = {
  equals?: Date[] | null;
  notEquals?: Date[] | null;
  any?: DateArrayQueryWhereCondition | null;
  all?: DateArrayQueryWhereCondition | null;
};

export type DateArrayQueryWhereCondition = {
  greaterThan?: Date | null;
  greaterThanOrEquals?: Date | null;
  lessThan?: Date | null;
  lessThanOrEquals?: number | null;
  equals?: Date | null;
  notEquals?: Date | null;
};

export type ContextAPI = {
  headers: RequestHeaders;
  response: Response;
  isAuthenticated: boolean;
  now(): Date;
};

export type Response = {
  headers: Headers;
  status?: number;
};

export type PageInfo = {
  startCursor: string;
  endCursor: string;
  totalCount: number;
  hasNextPage: boolean;
  count: number;
  pageNumber?: number;
};

type MimeType =
  | "application/json"
  | "application/gzip"
  | "application/pdf"
  | "application/rtf"
  | "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
  | "application/vnd.openxmlformats-officedocument.presentationml.presentation"
  | "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  | "application/vnd.ms-excel"
  | "application/vnd.ms-powerpoint"
  | "application/msword"
  | "application/zip"
  | "application/xml"
  | "application/x-7z-compressed"
  | "application/x-tar"
  | "image/gif"
  | "image/jpeg"
  | "image/svg+xml"
  | "image/png"
  | "text/html"
  | "text/csv"
  | "text/javascript"
  | "text/plain"
  | "text/calendar"
  | (string & {});

export type InlineFileConstructor = {
  filename: string;
  contentType: MimeType;
};
export declare class InlineFile {
  constructor(input: InlineFileConstructor);
  static fromDataURL(url: string): InlineFile;
  // Reads the contents of the file as a buffer
  read(): Promise<Buffer>;
  // Write the files contents from a buffer
  write(data: Buffer): void;
  // Persists the file
  store(expires?: Date, isPublic?: boolean): Promise<File>;
  // Gets the name of the file
  get filename(): string;
  // Gets the media type of the file contents
  get contentType(): string;
  // Gets size of the file's contents in bytes
  get size(): number;
}

export declare class File extends InlineFile {
  // Gets the stored key
  get key(): string;
  // Gets size of the file's contents in bytes
  get isPublic(): boolean;
  // Generates a presigned download URL
  getPresignedUrl(): Promise<URL>;
  // Creates a new instance from the database record
  static fromDbRecord(input: FileDbRecord): File;
  // Persists the file
  toDbRecord(): FileDbRecord;
}

export type FileDbRecord = {
  key: string;
  filename: string;
  contentType: string;
  size: number;
};

export declare class Duration {
  constructor(postgresString: string);
  static fromISOString(iso: DurationString): Duration;

  toISOString(): DurationString;
  toPostgres(): string;
}

export type SortDirection = "asc" | "desc" | "ASC" | "DESC";

// Request headers cannot be mutated, so remove any methods that mutate
export type RequestHeaders = Omit<Headers, "append" | "delete" | "set">;

export declare class Permissions {
  constructor();

  // allow() can be used to explicitly permit access to an action
  allow(): void;

  // deny() can be used to explicitly deny access to an action
  deny(): never;
}

declare class NotFoundError extends Error {}
declare class BadRequestError extends Error {}
declare class UnknownError extends Error {}

export type Errors = {
  /**
   * Returns a 404 HTTP status with an optional message.
   * This error indicates that the requested resource could not be found.
   */
  NotFound: typeof NotFoundError;
  /**
   * Returns a 400 HTTP status with an optional message.
   * This error indicates that the request made by the client is invalid or malformed.
   */
  BadRequest: typeof BadRequestError;
  /**
   * Returns a 500 HTTP status with an optional message.
   * This error indicates that an unexpected condition was encountered, preventing the server from fulfilling the request.
   */
  Unknown: typeof UnknownError;
};

export type FunctionConfig = {
  /**
   * All DB calls within the function will be executed within a transaction.
   * The transaction is rolled back if the function throws an error.
   */
  dbTransaction?: boolean;
};

export type FuncWithConfig<T> = T & {
  config: FunctionConfig;
};

type unit =
  | "year"
  | "years"
  | "month"
  | "months"
  | "day"
  | "days"
  | "hour"
  | "hours"
  | "minute"
  | "minutes"
  | "second"
  | "seconds";
type direction = "next" | "last";
type completed = "complete";
type value = number;

export type RelativeDateString =
  | "now"
  | "today"
  | "tomorrow"
  | "yesterday"
  | `this ${unit}`
  | `${direction} ${unit}`
  | `${direction} ${value} ${unit}`
  | `${direction} ${value} ${completed} ${unit}`;

type dateDuration =
  | `${number}Y${number}M${number}D` // Example: 1Y2M10D
  | `${number}Y${number}M` // Example: 1Y2M
  | `${number}Y${number}D` // Example: 1Y10D
  | `${number}M${number}D` // Example: 10M2D
  | `${number}Y` // Example: 1Y
  | `${number}M` // Example: 1M
  | `${number}D`; // Example: 2D

type timeDuration =
  | `${number}H${number}M${number}S` // Example: 2H30M
  | `${number}H${number}M` // Example: 2H30M
  | `${number}M${number}S` // Example: 2M30S
  | `${number}H${number}S` // Example: 2H30S
  | `${number}H` // Example: 2H
  | `${number}M` // Example: 30M
  | `${number}S`; // Example: 30S

export type DurationString =
  | `P${dateDuration}T${timeDuration}`
  | `P${dateDuration}`
  | `PT${timeDuration}`;

export type FileWriteTypes = InlineFile | File;

export enum UiIcon {
  check = "check",
}

export interface UI<C extends FlowConfig> {
  page: UiPage<C>;
  display: UiDisplayElements;
  inputs: UiInputsElements;
}

export type InputElement<TValueType, TConfig extends any = never> = <
  N extends string
>(
  name: N,
  options?: BaseInputConfig<TValueType> & TConfig
) => InputElementResponse<N, TValueType>;

export type DisplayElement<TConfig extends any = never> = (
  options?: TConfig
) => DisplayElementResponse;

// Input elements that are named and return values
type UiInputsElements = {
  text: UiElementInputText;
  number: InputElement<number>;
  toggle: InputElement<
    boolean,
    {
      mode?: "block" | "inline";
      description?: string;
      icon?: UiIcon;
    }
  >;
};

// Display elements that do not return values
type UiDisplayElements = {
  divider: DisplayElement;
};

type UiPage<C extends FlowConfig> = <
  T extends UIElements,
  const A extends PageActions[] = []
>(options: {
  stage?: ExtractStageKeys<C>;
  title?: string;
  description?: string;
  content: T;
  validate?: (data: ExtractFormData<T>) => Promise<true | string>;
  actions?: A;
}) => A["length"] extends 0
  ? ExtractFormData<T>
  : { data: ExtractFormData<T>; action: ActionValue<A[number]> };

type PageActions =
  | string
  | {
      label: string;
      value: string;
      mode?: "primary" | "secondary" | "destructive";
    };

export interface BaseInputConfig<T> {
  label: string;
  defaultValue?: T;
  optional?: boolean;
  validate?: (data: T) => Promise<true | string>;
}

type UIElements = (
  | InputElementResponse<string, any>
  | DisplayElementResponse
)[];

interface UIElementBase {
  _type: string;
}

export interface InputElementResponse<N extends string, V>
  extends UIElementBase {
  _type: "input";
  name: N;
  valueType: V;
}

export interface DisplayElementResponse extends UIElementBase {
  _type: "display";
}

export interface BaseUiInputResponse<K, T> {
  __type: K;
  name: string;
  label: string;
  defaultValue?: T;
  optional?: boolean;
}

export type ElementImplementation<TData, TApiResponse> = (
  ...args: Parameters<UiElementInputText>
) => {
  uiConfig: TApiResponse;
  getData: (data: TData) => TData;
  validate?: (data: TData) => boolean | string;
};


// element:

export type UiElementInputText = InputElement<
  string,
  {
    // element specific properties
    maxLength?: number;
  }
>;

// The shape of the response over the API
export interface UiElementInputTextApiResponse
  extends BaseUiInputResponse<"ui.input.text", string> {
  // Element specific properties
  maxLength?: number;
}

export const textInput: ElementImplementation<
  ReturnType<UiElementInputText>,
  UiElementInputTextApiResponse
> = (name, options) => {
  return {
    uiConfig: {
      __type: "ui.input.text",
      name,
      label: options?.label || name,
      defaultValue: options?.defaultValue,
      optional: options?.optional,
    },
	validate: (x: any) => true,
	getData: (x: any) => x,
  };
};





interface FlowConfig {
  stages?: StageConfig[];
  title?: string;
  description?: string;
}

type StageConfig =
  | string
  | {
      key: string;
      name: string;
      description?: string;
      initiallyHidden?: boolean;
    };

// Helper functions
type ActionValue<T> = T extends string
  ? T
  : T extends { value: infer V }
  ? V
  : never;

type ExtractFormData<T extends UIElements> = {
  [K in Extract<T[number], InputElementResponse<string, any>>["name"]]: Extract<
    T[number],
    InputElementResponse<K, any>
  >["valueType"];
};

type ExtractStageKeys<T extends FlowConfig> = T extends { stages: infer S }
  ? S extends ReadonlyArray<infer U>
    ? U extends string
      ? U
      : U extends { key: infer K extends string }
      ? K
      : never
    : never
  : never;

// Function overloads
export function flow<const C extends FlowConfig>(
  flowName: string,
  config: C,
  flow: FlowFunction<C>
): (inputs: FlowInputs) => any;
export function flow(
  flowName: string,
  flow: FlowFunction
): (inputs: FlowInputs) => any;