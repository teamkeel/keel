import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../..";

type ElementDataType = number;

export type UiElementInputNumber = InputElement<
  ElementDataType,
  {
    placeholder?: number;
    min?: number;
    max?: number;
  }
>;

// The shape of the response over the API
export interface UiElementInputNumberApiResponse
  extends BaseUiInputResponse<"ui.input.number", ElementDataType> {
  placeholder?: number;
  min?: number;
  max?: number;
}

export const numberInput: InputElementImplementation<
  ElementDataType,
  UiElementInputNumber,
  UiElementInputNumberApiResponse
> = (name, options) => {
  return {
    uiConfig: {
      __type: "ui.input.number",
      name,
      label: options?.label || name,
      defaultValue: options?.defaultValue,
      optional: options?.optional,
      placeholder: options?.placeholder,
      min: options?.min,
      max: options?.max,
    } satisfies UiElementInputNumberApiResponse,
    validate: options?.validate,
    getData: (x: any) => x,
  };
};
