import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  InputElementResponse,
  UiElementApiResponses,
  UIElements,
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
  validate?: (data: ExtractFormData<T>) => Promise<true | string>;
  actions?: A;
};
export type UiPage<C extends FlowConfig> = <
  T extends UIElements,
  const A extends PageActions[] = [],
>(
  name: string,
  options: PageOptions<C, A, T>
) => A["length"] extends 0
  ? ExtractFormData<T>
  : { data: ExtractFormData<T>; action: ActionValue<A[number]> };

type PageActions =
  | string
  | {
      label: string;
      value: string;
      mode?: "primary" | "secondary" | "destructive";
    };

export interface UiPageApiResponse extends BaseUiDisplayResponse<"ui.page"> {
  stage?: string;
  title?: string;
  description?: string;
  actions?: PageActions[];
  content: UiElementApiResponses;
}

export const page = <
  C extends FlowConfig,
  A extends PageActions[],
  T extends UIElements,
>(
  options: PageOptions<C, A, T>
): UiPageApiResponse => {
  return {
    __type: "ui.page",
    stage: options.stage,
    title: options.title,
    description: options.description,
    content: options.content as unknown as UiElementApiResponses,
    actions: options.actions,
  };
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
type ExtractFormData<T extends UIElements> = {
  [K in Extract<T[number], InputElementResponse<string, any>>["name"]]: Extract<
    T[number],
    InputElementResponse<K, any>
  >["valueType"];
};
