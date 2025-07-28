import {
  BaseInputConfig,
  BaseUiInputResponse,
  InputElementImplementation,
  InputElementResponse,
} from "../..";

type ElementDataType = string | number | boolean | Date;

// Annoyingly can't use the whole InputElement type and also use a local
// bounded type parameter to infer the type of the value from the config options
// So having to duplicate the types of the inputs
export type UiElementSelectMany = <
  const TValue extends ElementDataType,
  N extends string,
>(
  name: N,
  options?: BaseInputConfig<TValue> & {
    options: (
      | {
          label: string;
          value: TValue;
        }
      | TValue
    )[];
  }
) => InputElementResponse<N, TValue[]>;

// The shape of the response over the API
export interface UiElementSelectManyApiResponse
  extends BaseUiInputResponse<"ui.select.many", ElementDataType> {
  options: (
    | {
        label: string;
        value: ElementDataType;
      }
    | ElementDataType
  )[];
}

export const selectMany: InputElementImplementation<
  ElementDataType,
  UiElementSelectMany,
  UiElementSelectManyApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.select.many",
      name,
      label: options?.label || name,
      defaultValue: options?.defaultValue,
      optional: options?.optional || false,
      disabled: options?.disabled || false,
      helpText: options?.helpText,
      options: options?.options || [],
    } satisfies UiElementSelectManyApiResponse,
    validate: options?.validate,
    getData: (x: ElementDataType) => x,
  };
};
