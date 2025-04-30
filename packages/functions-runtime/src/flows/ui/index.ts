import { FlowConfig } from "..";
import { UiElementSelectOne } from "./elements/select/single";
import { UiElementInputText } from "./elements/input/text";
import { UiElementInputNumber } from "./elements/input/number";
import { UiElementInputBoolean } from "./elements/input/boolean";
import { UiElementMarkdown } from "./elements/display/markdown";
import { UiElementTable } from "./elements/display/table";
import { UiElementDivider } from "./elements/display/divider";
import { UiPage } from "./page";

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
  single: UiElementSelectOne;
};

// Display elements that do not return values
type UiDisplayElements = {
  divider: UiElementDivider;
  markdown: UiElementMarkdown;
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
