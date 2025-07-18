import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  InputElementImplementationResponse,
  InputElementResponse,
  IteratorElementImplementationResponse,
  IteratorElementResponse,
  UiElementApiResponse,
  UiElementApiResponses,
  UIElements,
  UIElement,
  ValidateFn,
} from ".";
import { UiElementIteratorApiResponse } from "./elements/iterator";

type PageOptions<
  C extends FlowConfig,
  A extends PageActions[],
  T extends UIElements,
> = {
  stage?: ExtractStageKeys<C>;
  title?: string;
  description?: string;
  content: T;
  validate?: ValidateFn<ExtractFormData<T>>;
  actions?: A;
};

export type UiPage<C extends FlowConfig> = <
  T extends UIElements,
  const A extends PageActions[] = [],
>(
  name: string,
  options: PageOptions<C, A, T>
) => Promise<
  A["length"] extends 0
    ? ExtractFormData<T>
    : { data: ExtractFormData<T>; action: ActionValue<A[number]> }
>;

type PageActions = string | PageActionConfig;
type PageActionConfig = {
  label: string;
  value: string;
  mode?: "primary" | "secondary" | "destructive";
};

// Extract the key from a custom action, supporting either a string or an object with a value property
type ActionValue<T> = T extends string
  ? T
  : T extends { value: infer V }
  ? V
  : never;

// Extract the data from elements and return a key-value object based on the name of the element
// Either from extracting directly from input elements or by extracting the already extracted types from an iterator element
export type ExtractFormData<T extends UIElements> = {
  [K in Extract<T[number], InputElementResponse<string, any>>["name"]]: Extract<
    T[number],
    InputElementResponse<K, any>
  >["valueType"];
} & {
  [K in Extract<
    T[number],
    IteratorElementResponse<string, any>
  >["name"]]: Extract<
    T[number],
    IteratorElementResponse<K, any>
  >["contentData"];
};

export interface UiPageApiResponse extends BaseUiDisplayResponse<"ui.page"> {
  stage?: string;
  title?: string;
  description?: string;
  actions?: PageActionConfig[];
  content: UiElementApiResponses;
  hasValidationErrors: boolean;
  validationError?: string;
}

export async function page<
  C extends FlowConfig,
  A extends PageActions[],
  T extends UIElements,
>(
  options: PageOptions<C, A, T>,
  data: any,
  action: string | null
): Promise<{ page: UiPageApiResponse; hasValidationErrors: boolean }> {
  // Turn these back into the actual response types
  const content = options.content as unknown as ImplementationResponse<
    any,
    any
  >[];
  let hasValidationErrors = false;
  let validationError: string | undefined;

  // if we have actions defined, validate that the given action exists
  if (options.actions && action !== null) {
    const isValidAction = options.actions.some((a) => {
      if (typeof a === "string") return a === action;
      return a && typeof a === "object" && "value" in a && a.value === action;
    });

    if (!isValidAction) {
      hasValidationErrors = true;
      validationError = "invalid action";
    }
  }

  const ret = await Promise.all(
    content.map(async (c) => {
      const resolvedC = await c;

      const elementData =
        data && typeof data === "object" && resolvedC.uiConfig.name in data
          ? data[resolvedC.uiConfig.name]
          : undefined;

      const { uiConfig, validationErrors } = await recursivelyProcessElement(
        c,
        elementData
      );

      if (validationErrors) hasValidationErrors = true;

      return uiConfig;
    })
  );

  // If there is page level validation, validate the data
  if (data && options.validate) {
    const validationResult = await options.validate(data);
    if (typeof validationResult === "string") {
      hasValidationErrors = true;
      validationError = validationResult;
    }
  }

  return {
    page: {
      __type: "ui.page",
      stage: options.stage,
      title: options.title,
      description: options.description,
      content: ret,
      actions: options.actions?.map((a) => {
        if (typeof a === "string") {
          return { label: a, value: a, mode: "primary" };
        } else if (typeof a === "object") {
          a.mode = a.mode || "primary";
        }
        return a;
      }),
      hasValidationErrors,
      validationError,
    },
    hasValidationErrors,
  };
}

const recursivelyProcessElement = async (
  c: ImplementationResponse<any, any>,
  data: any
): Promise<{ uiConfig: UiElementApiResponse; validationErrors: boolean }> => {
  const resolvedC = await c;
  const elementType = "__type" in resolvedC ? resolvedC.__type : null;

  switch (elementType) {
    case "input":
      return processInputElement(
        resolvedC as InputElementImplementationResponse<any, any>,
        data
      );
    case "iterator":
      return processIteratorElement(
        resolvedC as IteratorElementImplementationResponse<any, any>,
        data
      );
    default:
      return {
        uiConfig: { ...resolvedC.uiConfig },
        validationErrors: false,
      };
  }
};

const processInputElement = async (
  element: InputElementImplementationResponse<any, any>,
  data: any
): Promise<{ uiConfig: UiElementApiResponse; validationErrors: boolean }> => {
  const hasData = data !== undefined && data !== null;

  if (!hasData || !element.validate) {
    return {
      uiConfig: { ...element.uiConfig },
      validationErrors: false,
    };
  }

  const validationError = await element.validate(data);
  const hasValidationErrors = typeof validationError === "string";

  return {
    uiConfig: {
      ...element.uiConfig,
      validationError: hasValidationErrors ? validationError : undefined,
    },
    validationErrors: hasValidationErrors,
  };
};

const processIteratorElement = async (
  element: any,
  data: any
): Promise<{ uiConfig: UiElementApiResponse; validationErrors: boolean }> => {
  const elements = element.uiConfig.content as ImplementationResponse<
    any,
    any
  >[];
  const dataArr = data as any[] | undefined;

  // Process the UI config content
  const ui: UiElementApiResponse[] = [];
  for (const el of elements) {
    const result = await recursivelyProcessElement(el, undefined);
    ui.push(result.uiConfig);
  }

  // Check for validation errors if we have data
  const validationErrors = await validateIteratorData(elements, dataArr);
  let hasValidationErrors = validationErrors.length > 0;

  let validationError: string | undefined = undefined;
  if (dataArr && element.validate) {
    const v = await element.validate(dataArr);
    if (typeof v === "string") {
      hasValidationErrors = true;
      validationError = v;
    }
  }

  return {
    uiConfig: {
      ...element.uiConfig,
      content: ui,
      validationError: validationError,
      contentValidationErrors: hasValidationErrors
        ? validationErrors
        : undefined,
    },
    validationErrors: hasValidationErrors,
  };
};

const validateIteratorData = async (
  elements: ImplementationResponse<any, any>[],
  dataArr: any[] | undefined
): Promise<Array<{ index: number; name: string; validationError: string }>> => {
  const validationErrors: Array<{
    index: number;
    name: string;
    validationError: string;
  }> = [];

  if (!dataArr || dataArr.length === 0) {
    return validationErrors;
  }

  for (let i = 0; i < dataArr.length; i++) {
    const rowData = dataArr[i];

    for (const el of elements) {
      const resolvedEl = await el;

      if (isInputElementWithValidation(resolvedEl)) {
        const fieldName = resolvedEl.uiConfig.name;

        if (hasFieldData(rowData, fieldName)) {
          const validationError = await (resolvedEl as any).validate(
            rowData[fieldName]
          );

          if (typeof validationError === "string") {
            validationErrors.push({
              index: i,
              name: fieldName,
              validationError: validationError,
            });
          }
        }
      }
    }
  }

  return validationErrors;
};

const isInputElementWithValidation = (element: any): boolean => {
  return (
    "__type" in element &&
    element.__type === "input" &&
    "validate" in element &&
    element.validate &&
    typeof element.validate === "function"
  );
};

const hasFieldData = (rowData: any, fieldName: string): boolean => {
  return rowData && typeof rowData === "object" && fieldName in rowData;
};
