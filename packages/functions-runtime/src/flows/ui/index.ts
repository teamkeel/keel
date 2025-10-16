import { FlowConfig, Hardware, NullableHardware } from "..";
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
  UiElementInputDatePicker,
  UiElementInputDatePickerApiResponse,
} from "./elements/input/datePicker";

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
import { ExtractFormData, UiPage, UiPageApiResponse } from "./page";
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
import {
  UiElementKeyValue,
  UiElementKeyValueApiResponse,
} from "./elements/display/keyValue";
import { UiCompleteApiResponse } from "./complete";
import {
  UiElementSelectTable,
  UiElementSelectTableApiResponse,
} from "./elements/select/table";
import {
  UiElementInputDataGrid,
  UiElementInputDataGridApiResponse,
} from "./elements/input/dataGrid";
import {
  UiElementIterator,
  UiElementIteratorApiResponse,
} from "./elements/iterator";
import {
  UiElementPrint,
  UiElementPrintApiResponse,
} from "./elements/interactive/print";
import {
  UiElementPickList,
  UiElementPickListApiResponse,
} from "./elements/interactive/pickList";
import {
  UiElementScan,
  UiElementInputScanApiResponse,
} from "./elements/input/scan";
import {
  UiElementFile,
  UiElementFileApiResponse,
} from "./elements/display/file";

export interface UI<C extends FlowConfig, H extends NullableHardware> {
  page: UiPage<C>;
  display: UiDisplayElements;
  inputs: UiInputsElements;
  select: UiSelectElements;
  iterator: UiElementIterator;
  interactive: UiInteractiveElements<H>;
}

// Input elements that are named and return values
type UiInputsElements = {
  text: UiElementInputText;
  number: UiElementInputNumber;
  boolean: UiElementInputBoolean;
  dataGrid: UiElementInputDataGrid;
  datePicker: UiElementInputDatePicker;
  scan: UiElementScan;
};

// Select elements that are named and return values
type UiSelectElements = {
  one: UiElementSelectOne;
  table: UiElementSelectTable;
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
  keyValue: UiElementKeyValue;
  file: UiElementFile;
};

// Interactive elements may return values, others do not
type UiInteractiveElements<H extends NullableHardware> = {
  print: UiElementPrint<H>;
  pickList: UiElementPickList;
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

export type DisplayElementWithRequiredConfig<TConfig extends any = never> = (
  options: TConfig
) => DisplayElementResponse;

// Union of all element function shapes
export type UIElement =
  | InputElementResponse<string, any>
  | DisplayElementResponse
  | IteratorElementResponse<string, any>;

export type UIElements = UIElement[];

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

export interface IteratorElementResponse<N extends string, E extends UIElements>
  extends UIElementBase {
  _type: "iterator";
  name: N;
  contentData: ExtractFormData<E>[];
}

// Config that applied to all inputs
export interface BaseInputConfig<T, O extends boolean = boolean> {
  label?: string;
  defaultValue?: T;
  helpText?: string;
  optional?: O;
  disabled?: boolean;
  validate?: ValidateFn<T>;
  onLeave?: CallbackFn<T, T>;
}

export type ValidateFn<T> = (
  data: T
) => Promise<boolean | string> | boolean | string;

export type CallbackFn<InputT, OutputT> = (
  data: InputT
) => Promise<OutputT> | OutputT;

// Base response for all inputs
export interface BaseUiInputResponse<K, TData> {
  __type: K;
  name: string;
  label: string;
  defaultValue?: TData;
  optional: boolean;
  disabled: boolean;
  helpText?: string;
  validationError?: string;
}

export interface BaseUiMinimalInputResponse<K> {
  __type: K;
  name: string;
  validationError?: string;
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
    keyValue: UiElementKeyValueApiResponse;
    file: UiElementFileApiResponse;
  };
  input: {
    text: UiElementInputTextApiResponse;
    number: UiElementInputNumberApiResponse;
    boolean: UiElementInputBooleanApiResponse;
    dataGrid: UiElementInputDataGridApiResponse;
    datePicker: UiElementInputDatePickerApiResponse;
    scan: UiElementInputScanApiResponse;
  };
  select: {
    one: UiElementSelectOneApiResponse;
    table: UiElementSelectTableApiResponse;
  };
  iterator: UiElementIteratorApiResponse;
  interactive: {
    print: UiElementPrintApiResponse;
    pickList: UiElementPickListApiResponse;
  };
};

export type UiElementApiResponse =
  // Display elements
  | UiElementDividerApiResponse
  | UiElementMarkdownApiResponse
  | UiElementHeaderApiResponse
  | UiElementBannerApiResponse
  | UiElementImageApiResponse
  | UiElementCodeApiResponse
  | UiElementGridApiResponse
  | UiElementListApiResponse
  | UiElementTableApiResponse
  | UiElementKeyValueApiResponse
  | UiElementFileApiResponse

  // Input elements
  | UiElementInputTextApiResponse
  | UiElementInputNumberApiResponse
  | UiElementInputBooleanApiResponse
  | UiElementInputDataGridApiResponse
  | UiElementInputScanApiResponse

  // Select elements
  | UiElementSelectOneApiResponse
  | UiElementSelectTableApiResponse

  // Special
  | UiElementIteratorApiResponse

  // Interactive elements
  | UiElementPrintApiResponse
  | UiElementPickListApiResponse;

export type UiElementApiResponses = UiElementApiResponse[];

// The root API response. Used to generate the OpenAPI schema
export type UiApiUiConfig = UiPageApiResponse | UiCompleteApiResponse;

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
  __type: "input";
  uiConfig: TApiResponse;
  getData: (data: TData) => TData;
  validate?: ValidateFn<TData>;
  onLeave?: CallbackFn<TData, TData>;
};

export type IteratorElementImplementation<
  TData,
  TConfig extends (...args: any) => any,
  TApiResponse,
> = (
  ...args: Parameters<TConfig>
) => IteratorElementImplementationResponse<TApiResponse, TData>;

export type IteratorElementImplementationResponse<TApiResponse, TData> = {
  __type: "iterator";
  uiConfig: TApiResponse;
  getData: (data: TData) => TData;
  validate?: ValidateFn<TData>;
  onLeave?: CallbackFn<TData, TData>;
};

export type DisplayElementImplementation<
  TConfig extends (...args: any) => any,
  TApiResponse,
> = (
  ...args: Parameters<TConfig>
) => DisplayElementImplementationResponse<TApiResponse>;

export type DisplayElementImplementationResponse<TApiResponse> =
  | {
      uiConfig: TApiResponse;
    }
  | Promise<{
      uiConfig: TApiResponse;
    }>;

export type ImplementationResponse<TApiResponse, TData> =
  | InputElementImplementationResponse<TApiResponse, TData>
  | DisplayElementImplementationResponse<TApiResponse>
  | IteratorElementImplementationResponse<TApiResponse, TData>;
