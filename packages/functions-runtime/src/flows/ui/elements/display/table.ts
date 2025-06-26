import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementResponse,
} from "../..";

export type TableData<T extends Record<string, any>> = {
  data: T[];
  columns?: Array<Extract<keyof T, string>>;
};

export type UiElementTable = <const T extends Record<string, any>>(
  options: TableData<T>
) => DisplayElementResponse;

export type TableColumn = {
  name: string;
  index: number;
};

// The shape of the response over the API
export interface UiElementTableApiResponse
  extends BaseUiDisplayResponse<"ui.display.table"> {
  data: any[];
  columns: TableColumn[];
}

export const table: DisplayElementImplementation<
  UiElementTable,
  UiElementTableApiResponse
> = (options) => {
  const { data, columns } = processTableData(options.data, options.columns);

  return {
    uiConfig: {
      __type: "ui.display.table",
      data: data || [],
      columns: columns || [],
    } satisfies UiElementTableApiResponse,
  };
};

// Only send data for columns we need
export const processTableData = (data: any[], columnsConfig?: string[]) => {
  const filteredData = columnsConfig
    ? data.map((item) => {
        return Object.fromEntries(
          Object.entries(item).filter(
            ([key]) => columnsConfig?.includes(key as any)
          )
        );
      })
    : data;

  const cols = Object.keys(filteredData[0] || {});
  const columns: TableColumn[] = cols.map((column, index) => ({
    name: column,
    index,
  }));

  return { data: filteredData, columns };
};
