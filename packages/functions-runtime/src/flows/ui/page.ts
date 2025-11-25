import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  InputElementImplementationResponse,
  InputElementResponse,
  IteratorElementImplementationResponse,
  UiElementApiResponse,
  UiElementApiResponses,
  UIElements,
  ValidateFn,
} from ".";

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
  // When true, let the use go back a step if there is a previous step
  allowBack?: boolean;
  fullWidth?: boolean;
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
  /**
   * Keyboard shortcut for this action.
   *
   * @example
   * // Single key
   * hotkey: "1"
   * hotkey: "enter"
   * hotkey: "space"
   * hotkey: "F1"
   *
   *
   * // With modifier:
   * // shift, alt + mod (cmd on macOS, ctrl on Windows/Linux)
   *
   * hotkey: "mod+s"           // cmd+s on macOS, ctrl+s on Windows/Linux
   * hotkey: "mod+shift+enter" // cmd+shift+enter on macOS, ctrl+shift+enter on Windows/Linux
   *
   * // With behavior control
   * hotkey: {
   *   key: "mod+enter",
   *   behaviour: "includeForms"  // Trigger even when focused on form elements
   * }
   */
  hotkey?: string | HotkeyConfig;
};

type HotkeyConfig = {
  key: string;
  /** If the hotkey should be triggered when the user is on a form element or not
   *
   * @default "excludeForms"
   */
  behaviour?: "includeForms" | "excludeForms";
};

export interface UiPageApiResponse extends BaseUiDisplayResponse<"ui.page"> {
  stage?: string;
  title?: string;
  description?: string;
  actions?: PageActionConfig[];
  content: UiElementApiResponses;
  hasValidationErrors: boolean;
  validationError?: string;
  allowBack?: boolean;
  fullWidth?: boolean;
}

export async function callbackFn<T extends UIElements>(
  elements: T,
  elementName: string,
  callbackName: string,
  data: any
): Promise<any> {
  // Find the element by name
  const element = elements.find(
    (el) => (el as any).uiConfig && (el as any).uiConfig.name === elementName
  );

  if (!element) {
    throw new Error(`Element with name ${elementName} not found`);
  }

  // Check if the callback exists on the element
  const cb = (element as any)[callbackName];
  if (typeof cb !== "function") {
    throw new Error(
      `Callback ${callbackName} not found on element ${elementName}`
    );
  }

  // Call the callback with provided arguments
  return await cb(data);
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

  // Validate actions - all action validation errors are thrown as bad requests
  if (options.actions && options.actions.length > 0) {
    // Actions are defined on the page
    // Normalize action - convert "undefined" string or undefined value to null
    const normalizedAction =
      action === "undefined" || action === undefined ? null : action;

    // Only validate action when data is being submitted (data is not null)
    if (data !== null) {
      if (normalizedAction === null) {
        // No action provided when actions are required during submission
        const validValues = options.actions
          .map((a) => {
            if (typeof a === "string") return a;
            return a && typeof a === "object" && "value" in a ? a.value : "";
          })
          .filter(Boolean);
        throw new Error(
          `action is required. Valid actions are: ${validValues.join(", ")}`
        );
      }

      // Action provided, check if it's valid
      const isValidAction = options.actions.some((a) => {
        if (typeof a === "string") return a === normalizedAction;
        return (
          a &&
          typeof a === "object" &&
          "value" in a &&
          a.value === normalizedAction
        );
      });

      if (!isValidAction) {
        // Build list of valid action values
        const validValues = options.actions
          .map((a) => {
            if (typeof a === "string") return a;
            return a && typeof a === "object" && "value" in a ? a.value : "";
          })
          .filter(Boolean);

        throw new Error(
          `invalid action "${normalizedAction}". Valid actions are: ${validValues.join(
            ", "
          )}`
        );
      }
    }
  } else if (
    action !== null &&
    action !== undefined &&
    action !== "undefined"
  ) {
    // No actions defined but a real action value was provided - this is a bad request
    throw new Error(
      `invalid action "${action}". No actions are defined for this page`
    );
  }

  const contentUiConfig = await Promise.all(
    content.map(async (c) => {
      const resolvedC = await c;

      const elementData =
        data && typeof data === "object" && resolvedC.uiConfig.name in data
          ? data[resolvedC.uiConfig.name]
          : undefined;

      const { uiConfig, validationErrors } = await recursivelyProcessElement(
        c,
        elementData,
        options.actions && options.actions.length > 0 ? action : null
      );

      if (validationErrors) hasValidationErrors = true;

      return uiConfig;
    })
  );

  // If there is page level validation, validate the data
  if (data && options.validate) {
    const validationResult =
      options.actions && action !== null
        ? await (options.validate as any)(data, action)
        : await (options.validate as any)(data);
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
      content: contentUiConfig,
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
      allowBack: options.allowBack,
      fullWidth: options.fullWidth,
    },
    hasValidationErrors,
  };
}

const recursivelyProcessElement = async (
  c: ImplementationResponse<any, any>,
  data: any,
  action: string | null
): Promise<{ uiConfig: UiElementApiResponse; validationErrors: boolean }> => {
  const resolvedC = await c;
  const elementType = "__type" in resolvedC ? resolvedC.__type : null;

  switch (elementType) {
    case "input":
      return processInputElement(
        resolvedC as InputElementImplementationResponse<any, any>,
        data,
        action
      );
    case "iterator":
      return processIteratorElement(
        resolvedC as IteratorElementImplementationResponse<any, any>,
        data,
        action
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
  data: any,
  action: string | null
): Promise<{ uiConfig: UiElementApiResponse; validationErrors: boolean }> => {
  const hasData = data !== undefined && data !== null;

  if (!hasData || !element.validate) {
    return {
      uiConfig: { ...element.uiConfig },
      validationErrors: false,
    };
  }

  const validationError =
    action !== null
      ? await (element.validate as any)(data, action)
      : await (element.validate as any)(data);
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
  data: any,
  action: string | null
): Promise<{ uiConfig: UiElementApiResponse; validationErrors: boolean }> => {
  const elements = element.uiConfig.content as ImplementationResponse<
    any,
    any
  >[];
  const dataArr = data as any[] | undefined;

  // Process the UI config content
  const ui: UiElementApiResponse[] = [];
  for (const el of elements) {
    const result = await recursivelyProcessElement(el, undefined, action);
    ui.push(result.uiConfig);
  }

  // Check for validation errors if we have data
  const validationErrors = await validateIteratorData(
    elements,
    dataArr,
    action
  );
  let hasValidationErrors = validationErrors.length > 0;

  let validationError: string | undefined = undefined;
  if (dataArr && element.validate) {
    const v =
      action !== null
        ? await (element.validate as any)(dataArr, action)
        : await (element.validate as any)(dataArr);
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
  dataArr: any[] | undefined,
  action: string | null
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

      if (
        "__type" in resolvedEl &&
        resolvedEl.__type === "input" &&
        "validate" in resolvedEl &&
        resolvedEl.validate &&
        typeof resolvedEl.validate === "function"
      ) {
        const fieldName = resolvedEl.uiConfig.name;

        if (rowData && typeof rowData === "object" && fieldName in rowData) {
          const validationError =
            action !== null
              ? await (resolvedEl as any).validate(rowData[fieldName], action)
              : await (resolvedEl as any).validate(rowData[fieldName]);

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

/* ********************
 * Helper functions
 ******************* */

// Extract the key from a custom action, supporting either a string or an object with a value property
type ActionValue<T> = T extends string
  ? T
  : T extends { value: infer V }
  ? V
  : never;

// Extract the data from elements and return a key-value object based on the name of the element
export type ExtractFormData<T extends UIElements> = {
  [K in Extract<T[number], InputElementResponse<string, any>>["name"]]: Extract<
    T[number],
    InputElementResponse<K, any>
  >["returnType"];
};
