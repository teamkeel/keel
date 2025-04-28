import { InputElementResponse } from ".";

import { UIElements } from ".";
import { FlowConfig } from "..";

export type UiPage<C extends FlowConfig> = <
  T extends UIElements,
  A extends PageActions[] = []
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

// Extract the stage keys from the flow config supporting either a string or an object with a key property
type ExtractStageKeys<T extends FlowConfig> = T extends { stages: infer S }
  ? S extends ReadonlyArray<infer U>
    ? U extends string
      ? U
      : U extends { key: infer K extends string }
      ? K
      : never
    : never
  : never;
