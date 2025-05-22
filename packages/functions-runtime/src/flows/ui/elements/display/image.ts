import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

export type UiElementImage = DisplayElement<{
  url: string;
  alt?: string;
  size?: "thumbnail" | "small" | "medium" | "large" | "full";
  caption?: string;
}>;

// The shape of the response over the API
export interface UiElementImageApiResponse
  extends BaseUiDisplayResponse<"ui.display.image"> {
  url: string;
  alt?: string;
  size?: "thumbnail" | "small" | "medium" | "large" | "full";
  caption?: string;
}

export const image: DisplayElementImplementation<
  UiElementImage,
  UiElementImageApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.image",
      url: options?.url || "",
      alt: options?.alt,
      size: options?.size,
      caption: options?.caption,
    } satisfies UiElementImageApiResponse,
  };
};
