import { sentenceCase } from "change-case";
import {
  BaseInputConfig,
  BaseUiInputResponse,
  InputElementImplementation,
  InputElementResponse,
} from "../..";
import { processTableData } from "../display/table";

// Can't use the whole InputElement type and also use a local
// bounded type parameter to infer the type of the value from the config options
// So having to duplicate the types of the inputs
export type UiElementInputDataGrid = <
  N extends string,
  T extends Record<string, any>,
  const Cols extends Extract<keyof T, string>[],
  C extends ColumnConfig<Cols>[] | undefined = undefined,
>(
  name: N,
  options: DataGridOptions<T, C>
) => InputElementResponse<
  N,
  C extends ColumnConfig<Cols>[]
    ? {
        // If the column has a type then we need to cast the value to the type of the column
        // Otherwise we just use the type of the value
        [K in C[number] as K["key"]]: MapDataType<K["type"], T[K["key"]]>;
      }[]
    : T[]
>;

type MapDataType<U, Fallback> = U extends "text"
  ? string
  : U extends "number"
  ? number
  : U extends "boolean"
  ? boolean
  : U extends "id"
  ? string
  : U extends "hidden"
  ? Fallback
  : Fallback;

type ColumnConfig<Cols extends string[]> = {
  key: Cols[number];
  label?: string;
  type?: DataTypes;
  editable?: boolean;
};

export type DataGridOptions<T extends Record<string, any>, C> = Omit<
  BaseInputConfig<T>,
  "defaultValue" | "label" | "optional" | "disabled"
> & {
  data: T[];
  columns?: C;
  allowAddRows?: boolean;
  allowDeleteRows?: boolean;
};

// The shape of the response over the API
export interface UiElementInputDataGridApiResponse
  extends Omit<
    BaseUiInputResponse<"ui.input.dataGrid", any>,
    "defaultValue" | "label" | "optional" | "disabled"
  > {
  data: any[];
  columns: DataGridColumn[];
  allowAddRows: boolean;
  allowDeleteRows: boolean;
}

type DataTypes = "text" | "number" | "boolean" | "id" | "hidden";

export type DataGridColumn = {
  key: string;
  label: string;
  index: number;
  type: DataTypes;
  editable: boolean;
};

export const dataGridInput: InputElementImplementation<
  any,
  UiElementInputDataGrid,
  UiElementInputDataGridApiResponse
> = (name, options) => {
  const { data } = processTableData(
    options.data,
    options.columns?.map((c) => c.key)
  );

  const inferType = (key: string) => {
    const inferredTypeRaw = typeof data[0][key];
    const inferredTypeMap: Record<typeof inferredTypeRaw, DataTypes> = {
      string: "text",
      number: "number",
      boolean: "boolean",
      bigint: "number",
      symbol: "text",
      undefined: "text",
      object: "text",
      function: "text",
    };
    return inferredTypeMap[inferredTypeRaw] ?? "text";
  };

  const columns = options.columns
    ? options.columns?.map((column, idx) => ({
        key: column.key,
        label: column.label || column.key,
        index: idx,
        type: column.type || inferType(column.key),
        editable:
          column.editable === undefined
            ? column.type === "id"
              ? false
              : true
            : column.editable,
      }))
    : Object.keys(data[0]).map((key, idx) => {
        return {
          index: idx,
          key,
          label: sentenceCase(key),
          type: inferType(key),
          editable: true,
        };
      });

  return {
    __type: "input",
    uiConfig: {
      __type: "ui.input.dataGrid",
      name,
      data,
      columns,
      helpText: options?.helpText,
      allowAddRows: options?.allowAddRows ?? false,
      allowDeleteRows: options?.allowDeleteRows ?? false,
    } satisfies UiElementInputDataGridApiResponse,
    validate: options?.validate, // TODO have some built in validation that checks the types of the response
    getData: (x: any) => x,
  };
};
