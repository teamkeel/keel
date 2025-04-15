import { UiElementSelectOne } from "./elements/select/single";
import { UiElementInputText } from "./elements/input/text";
import { UiElementInputNumber } from "./elements/input/number";
import { UiElementInputBoolean } from "./elements/input/boolean";
import { UiElementMarkdown } from "./elements/display/markdown";
import { UiElementTable } from "./elements/display/table";
import { UiElementDivider } from "./elements/display/divider";
import { FlowConfig } from "..";

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

// Input elements that are named and return values
type UiSelectElements = {
  single: UiElementSelectOne;
};

// Display elements that do not return values
type UiDisplayElements = {
  divider: UiElementDivider;
  markdown: UiElementMarkdown;
  table: UiElementTable;
};

type UiPage<C extends FlowConfig> = <
  T extends UIElements,
  const A extends PageActions[] = [],
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

export type InputElement<TValueType, TConfig extends any = never> = <
  N extends string,
>(
  name: N,
  options?: BaseInputConfig<TValueType> & TConfig
) => InputElementResponse<N, TValueType>;

export type DisplayElement<TConfig extends any = never> = (
  options?: TConfig
) => DisplayElementResponse;

export interface BaseInputConfig<T> {
  label?: string;
  defaultValue?: T;
  helpText?: string;
  optional?: boolean;
  disabled?: boolean;
  validate?: (data: T) => Promise<boolean | string>;
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

export interface BaseUiInputResponse<K, TData> {
  __type: K;
  name: string;
  label: string;
  defaultValue?: TData;
  helpText?: string;
  optional?: boolean;
  disabled?: boolean;
}

export interface BaseUiDisplayResponse<K> {
  __type: K;
}

export type InputElementImplementation<
  TData,
  TConfig extends (...args: any) => any,
  TApiResponse,
> = (...args: Parameters<TConfig>) => {
  uiConfig: TApiResponse;
  getData: (data: TData) => TData;
  validate?: (data: TData) => Promise<boolean | string>;
};

export type DisplayElementImplementation<
  TConfig extends (...args: any) => any,
  TApiResponse,
> = (...args: Parameters<TConfig>) => {
  uiConfig: TApiResponse;
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
