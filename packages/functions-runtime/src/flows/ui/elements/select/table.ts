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
>(
  name: N,
  options: TableOptions<T, M>
) => InputElementResponse<N, M extends "single" ? T : T[]>;

export type TableOptions<
  T extends Record<string, any>,
  M extends SelectMode = "multi",
> = Omit<BaseInputConfig<T>, "defaultValue" | "label"> &
  TableData<T> & {
    mode?: M;
  };

// The shape of the response over the API
export interface UiElementSelectTableApiResponse
  extends Omit<
    BaseUiInputResponse<"ui.select.table", any>,
    "defaultValue" | "label"
  > {
  data: any[];
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
