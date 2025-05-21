import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

export type UiElementHeader = DisplayElement<{
  /**
   * The visual level of the header.
   *
   * @default 2
   */
  level?: 1 | 2 | 3;
  /**
   * The title of the header.
   */
  title?: string;
  /**
   * The description of the header.
   */
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
      level: options?.level || 2,
      title: options?.title || "",
      description: options?.description || "",
    } satisfies UiElementHeaderApiResponse,
  };
};
