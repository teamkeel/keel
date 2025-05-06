import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

export type UiElementHeader = DisplayElement<{
  level: 1 | 2 | 3;
  title?: string;
  description?: string;
}>;

// The shape of the response over the API
export interface UiElementHeaderApiResponse
  extends BaseUiDisplayResponse<"ui.display.header"> {
  level: number;
  title: string;
  description: string;
}

export const header: DisplayElementImplementation<
  UiElementHeader,
  UiElementHeaderApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.header",
      level: options?.level || 1,
      title: options?.title || "",
      description: options?.description || "",
    } satisfies UiElementHeaderApiResponse,
  };
};
