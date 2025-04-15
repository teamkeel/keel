import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../..";

type ElementDataType = number;

export type UiElementInputNumber = InputElement<
  ElementDataType,
  {
    placeholder?: string;
    min?: number;
    max?: number;
  }
>;

// The shape of the response over the API
export interface UiElementInputNumberApiResponse
  extends BaseUiInputResponse<"ui.input.number", ElementDataType> {
  placeholder: string;
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
      placeholder: "",
    },
    validate: options?.validate,
    getData: (x: any) => x,
  };
};
