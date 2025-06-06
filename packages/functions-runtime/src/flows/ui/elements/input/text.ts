import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../..";

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
    __type: "input",
    uiConfig: {
      __type: "ui.input.text",
      name,
      label: options?.label || name,
      optional: options?.optional || false,
      disabled: options?.disabled || false,
      helpText: options?.helpText,
      defaultValue: options?.defaultValue,
      placeholder: options?.placeholder,
      multiline: options?.multiline,
      maxLength: options?.maxLength,
      minLength: options?.minLength,
    } satisfies UiElementInputTextApiResponse,
    validate: options?.validate,
    getData: (x: ElementDataType) => x,
  };
};
