import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  InputElementResponse,
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

  const contentUiConfig = (await Promise.all(
    content
      .map(async (content) => {
        let c = await content;
        const isInput = "__type" in c && c.__type == "input";
        const hasData = data && c.uiConfig.name in data;
        if (isInput && hasData && "validate" in c && c.validate) {
          const validationError = await c.validate(data[c.uiConfig.name]);
          if (typeof validationError === "string") {
            hasValidationErrors = true;
            return {
              ...c.uiConfig,
              validationError,
            };
          }
        }

        return c.uiConfig;
      })
      .filter(Boolean)
  )) as UiElementApiResponses;

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
    },
    hasValidationErrors,
  };
}

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
type ExtractFormData<T extends UIElements> = {
  [K in Extract<T[number], InputElementResponse<string, any>>["name"]]: Extract<
    T[number],
    InputElementResponse<K, any>
  >["valueType"];
};
