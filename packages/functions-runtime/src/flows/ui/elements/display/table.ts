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
  // Only send data for columns we need
  const filteredData = options.columns
    ? options.data.map((item) => {
        return Object.fromEntries(
          Object.entries(item).filter(
            ([key]) => options.columns?.includes(key as any)
          )
        );
      })
    : options.data;

  const cols = Object.keys(filteredData[0] || {});
  const columns = cols.map((column, index) => ({
    name: column,
    index,
  }));

  return {
    uiConfig: {
      __type: "ui.display.table",
      data: filteredData || [],
      columns: columns || [],
    } satisfies UiElementTableApiResponse,
  };
};
