import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../../..";

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
    uiConfig: {
      __type: "ui.input.boolean",
      name,
      label: options?.label || name,
      defaultValue: options?.defaultValue,
      optional: options?.optional,
      placeholder: "",
      mode: options?.mode || "checkbox",
    },
    validate: options?.validate,
    getData: (x: any) => x,
  };
};
