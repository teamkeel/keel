import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../..";

type ElementDataType = boolean;

export type UiElementInputBoolean = InputElement<
  ElementDataType,
  {
    mode?: "checkbox" | "switch";
  }
>;

// The shape of the response over the API
export interface UiElementInputBooleanApiResponse
  extends BaseUiInputResponse<"ui.input.boolean", ElementDataType> {
  mode: "checkbox" | "switch";
}

export const booleanInput: InputElementImplementation<
  ElementDataType,
  UiElementInputBoolean,
  UiElementInputBooleanApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.input.boolean",
      name,
      label: options?.label || name,
      optional: options?.optional || false,
      disabled: options?.disabled || false,
      helpText: options?.helpText,
      defaultValue: options?.defaultValue,
      mode: options?.mode || "checkbox",
    } satisfies UiElementInputBooleanApiResponse,
    validate: options?.validate,
    onLeave: options?.onLeave,
    getData: (x: any) => x,
  };
};
