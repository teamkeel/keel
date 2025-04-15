import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../../..";

export type UiElementTable = DisplayElement<{
  data: any[];
  columns: string[];
}>;

// The shape of the response over the API
export interface UiElementTableApiResponse
  extends BaseUiDisplayResponse<"ui.display.table"> {
  data: any[];
  columns: string[];
}

export const table: DisplayElementImplementation<
  UiElementTable,
  UiElementTableApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.table",
      data: options?.data || [],
      columns: options?.columns || [],
    },
  };
};
