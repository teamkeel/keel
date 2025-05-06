import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementResponse,
} from "../..";

type TableData<T extends Record<string, any>> = {
  data: T[];
  columns?: Array<Extract<keyof T, string>>;
};

export type UiElementTable = <const T extends Record<string, any>>(
  options: TableData<T>
) => DisplayElementResponse;

// The shape of the response over the API
export interface UiElementTableApiResponse
  extends BaseUiDisplayResponse<"ui.display.table"> {
  data: any[];
  columns?: string[]; // Todo: support for an object form with extra context on the data type
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

  return {
    uiConfig: {
      __type: "ui.display.table",
      data: filteredData || [],
      columns: options.columns,
    } satisfies UiElementTableApiResponse,
  };
};
