import {
  BaseInputConfig,
  BaseUiInputResponse,
  InputElementImplementation,
  InputElementResponse,
} from "../..";
import { processTableData, TableColumn, TableData } from "../display/table";

export type SelectMode = "single" | "multi";

// Can't use the whole InputElement type and also use a local
// bounded type parameter to infer the type of the value from the config options
// So having to duplicate the types of the inputs
export type UiElementSelectTable = <
  const T extends Record<string, any>,
  N extends string,
  const M extends SelectMode = "multi",
  const O extends boolean = false,
>(
  name: N,
  options: TableOptions<T, O, M>
) => InputElementResponse<
  N,
  M extends "single" ? (O extends true ? T | undefined : T) : T[]
>;

export type TableOptions<
  T extends Record<string, any>,
  O extends boolean,
  M extends SelectMode,
> = Omit<
  BaseInputConfig<
    M extends "single" ? (O extends true ? T | undefined : T) : T[],
    O
  >,
  "defaultValue" | "label"
> &
  TableData<T> &
  (
    | {
        mode?: M;
      }
    | {
        mode: "multi";
        max?: number;
        min?: number;
      }
  );

// The shape of the response over the API
export interface UiElementSelectTableApiResponse
  extends Omit<
    BaseUiInputResponse<"ui.select.table", any>,
    "defaultValue" | "label"
  > {
  data: Record<string, any>[];
  columns: TableColumn[];
  mode: SelectMode;
}

export const selectTable: InputElementImplementation<
  any,
  UiElementSelectTable,
  UiElementSelectTableApiResponse
> = (name, options) => {
  const { data, columns } = processTableData(options.data, options.columns);

  return {
    __type: "input",
    uiConfig: {
      __type: "ui.select.table",
      name,
      data,
      columns,
      mode: options?.mode || "multi",
      optional: options?.optional || false,
      disabled: options?.disabled || false,
      helpText: options?.helpText,
    } satisfies UiElementSelectTableApiResponse,
    validate: options?.validate,
    getData: (x: any) => x,
  };
};
