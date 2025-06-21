import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

type KeyValueMode = "list" | "grid" | "card";

// The types the user experiences
export type UiElementKeyValue = DisplayElement<{
  data: {
    key: string;
    value: string | number | Date | boolean; // TODO: support for an object with richer type info / linking
  }[];
  mode: KeyValueMode;
}>;

// The shape of the response over the API
export interface UiElementKeyValueApiResponse
  extends BaseUiDisplayResponse<"ui.display.keyValue"> {
  data: {
    key: string;
    value: string | number | Date | boolean;
    index: number;
  }[];
  mode: KeyValueMode;
}

// The implementation
export const keyValue: DisplayElementImplementation<
  UiElementKeyValue,
  UiElementKeyValueApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.keyValue",
      data:
        options?.data?.map((item, index) => ({
          ...item,
          index,
        })) || [],
      mode: options?.mode || "list",
    } satisfies UiElementKeyValueApiResponse,
  };
};
