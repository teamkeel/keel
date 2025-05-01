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
export type UiElementSelectOne = <
  TValue extends ElementDataType,
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
) => InputElementResponse<N, TValue>;

// The shape of the response over the API
export interface UiElementSelectOneApiResponse
  extends BaseUiInputResponse<"ui.select.single", ElementDataType> {
  options: {
    label: string;
    value: string;
  }[];
}

export const selectOne: InputElementImplementation<
  ElementDataType,
  UiElementSelectOne,
  UiElementSelectOneApiResponse
> = (name, options) => {
  return {
    uiConfig: {
      __type: "ui.select.single",
      name,
      label: options?.label || name,
      defaultValue: options?.defaultValue,
      optional: options?.optional,
      options: [],
    } satisfies UiElementSelectOneApiResponse,
    validate: options?.validate,
    getData: (x: any) => x,
  };
};
