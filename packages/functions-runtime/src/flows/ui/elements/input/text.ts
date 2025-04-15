import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../../..";

type ElementDataType = string;

export type UiElementInputText = InputElement<
  ElementDataType,
  {
    placeholder?: string;
    multiline?: boolean;
    maxLength?: number;
    minLength?: number;
  }
>;

// The shape of the response over the API
export interface UiElementInputTextApiResponse
  extends BaseUiInputResponse<"ui.input.text", ElementDataType> {
  placeholder?: string;
  multiline?: boolean;
  maxLength?: number;
  minLength?: number;
}

export const textInput: InputElementImplementation<
  ElementDataType,
  UiElementInputText,
  UiElementInputTextApiResponse
> = (name, options) => {
  return {
    uiConfig: {
      __type: "ui.input.text",
      name,
      label: options?.label || name,
      defaultValue: options?.defaultValue,
      optional: options?.optional,
    },
    validate: options?.validate,
    getData: (x: any) => x,
  };
};
