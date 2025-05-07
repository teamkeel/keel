import { createFlowContext, FlowConfig } from "..";
import {
  UiElementSelectOne,
  UiElementSelectOneApiResponse,
} from "./elements/select/one";
import {
  UiElementInputText,
  UiElementInputTextApiResponse,
} from "./elements/input/text";
import {
  UiElementInputNumber,
  UiElementInputNumberApiResponse,
} from "./elements/input/number";
import {
  UiElementInputBoolean,
  UiElementInputBooleanApiResponse,
} from "./elements/input/boolean";
import {
  UiElementMarkdown,
  UiElementMarkdownApiResponse,
} from "./elements/display/markdown";
import {
  UiElementTable,
  UiElementTableApiResponse,
} from "./elements/display/table";
import {
  UiElementDivider,
  UiElementDividerApiResponse,
} from "./elements/display/divider";
import { UiPage, UiPageApiResponse } from "./page";
import {
  UiElementImage,
  UiElementImageApiResponse,
} from "./elements/display/image";
import {
  UiElementBanner,
  UiElementBannerApiResponse,
} from "./elements/display/banner";
import {
  UiElementHeader,
  UiElementHeaderApiResponse,
} from "./elements/display/header";
import {
  UiElementCode,
  UiElementCodeApiResponse,
} from "./elements/display/code";
import {
  UiElementGrid,
  UiElementGridApiResponse,
} from "./elements/display/grid";
import {
  UiElementList,
  UiElementListApiResponse,
} from "./elements/display/list";

export interface UI<C extends FlowConfig> {
  page: UiPage<C>;
  display: UiDisplayElements;
  inputs: UiInputsElements;
  select: UiSelectElements;
}

// Input elements that are named and return values
type UiInputsElements = {
  text: UiElementInputText;
  number: UiElementInputNumber;
  boolean: UiElementInputBoolean;
};

// Select elements that are named and return values
type UiSelectElements = {
  one: UiElementSelectOne;
};

// Display elements that do not return values
type UiDisplayElements = {
  divider: UiElementDivider;
  markdown: UiElementMarkdown;
  header: UiElementHeader;
  banner: UiElementBanner;
  image: UiElementImage;
  code: UiElementCode;
  grid: UiElementGrid;
  list: UiElementList;
  table: UiElementTable;
};

// The base input element function. All inputs must be named and can optionally have a config
export type InputElement<TValueType, TConfig extends any = never> = <
  N extends string,
>(
  name: N,
  options?: BaseInputConfig<TValueType> & TConfig
) => InputElementResponse<N, TValueType>;

// The base display element function. Display elements do not have a name but optionally have a config
export type DisplayElement<TConfig extends any = never> = (
  options?: TConfig
) => DisplayElementResponse;

// Union of all element function shapes
export type UIElements = (
  | InputElementResponse<string, any>
  | DisplayElementResponse
)[];

// We create internal _type properties to help identity inputs vs display elements
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

// Config that applied to all inputs
export interface BaseInputConfig<T> {
  label?: string;
  defaultValue?: T;
  helpText?: string;
  optional?: boolean;
  disabled?: boolean;
  validate?: (data: T) => Promise<boolean | string>;
}

// Base response for all inputs
export interface BaseUiInputResponse<K, TData> {
  __type: K;
  name: string;
  label: string;
  defaultValue?: TData;
  helpText?: string;
  optional?: boolean;
  disabled?: boolean;
}

// Base response for all display elements
export interface BaseUiDisplayResponse<K> {
  __type: K;
}

export type UIApiResponses = {
  display: {
    divider: UiElementDividerApiResponse;
    markdown: UiElementMarkdownApiResponse;
    header: UiElementHeaderApiResponse;
    banner: UiElementBannerApiResponse;
    image: UiElementImageApiResponse;
    code: UiElementCodeApiResponse;
    grid: UiElementGridApiResponse;
    list: UiElementListApiResponse;
    table: UiElementTableApiResponse;
  };
  input: {
    text: UiElementInputTextApiResponse;
    number: UiElementInputNumberApiResponse;
    boolean: UiElementInputBooleanApiResponse;
  };
  select: {
    one: UiElementSelectOneApiResponse;
  };
};

export type UiElementApiResponses = // Display elements
  (
    | UiElementDividerApiResponse
    | UiElementMarkdownApiResponse
    | UiElementHeaderApiResponse
    | UiElementBannerApiResponse
    | UiElementImageApiResponse
    | UiElementCodeApiResponse
    | UiElementGridApiResponse
    | UiElementListApiResponse
    | UiElementTableApiResponse

    // Input elements
    | UiElementInputTextApiResponse
    | UiElementInputNumberApiResponse
    | UiElementInputBooleanApiResponse

    // Select elements
    | UiElementSelectOneApiResponse
  )[];

// The root API response. Used to generate the OpenAPI schema
export type UiApiUiConfig = UiPageApiResponse;

/* ********************
 * Implementations
 ******************* */

export type InputElementImplementation<
  TData,
  TConfig extends (...args: any) => any,
  TApiResponse,
> = (
  ...args: Parameters<TConfig>
) => InputElementImplementationResponse<TApiResponse, TData>;

export type InputElementImplementationResponse<TApiResponse, TData> = {
  uiConfig: TApiResponse;
  getData: (data: TData) => TData;
  validate?: (data: TData) => Promise<boolean | string>;
};

export type DisplayElementImplementation<
  TConfig extends (...args: any) => any,
  TApiResponse,
> = (
  ...args: Parameters<TConfig>
) => DisplayElementImplementationResponse<TApiResponse>;

export type DisplayElementImplementationResponse<TApiResponse> = {
  uiConfig: TApiResponse;
};
