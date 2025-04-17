import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

export type UiElementMarkdown = DisplayElement<{
  content: string;
}>;

// The shape of the response over the API
export interface UiElementMarkdownApiResponse
  extends BaseUiDisplayResponse<"ui.display.markdown"> {
  content: string;
}

export const markdown: DisplayElementImplementation<
  UiElementMarkdown,
  UiElementMarkdownApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.markdown",
      content: options?.content || "",
    },
  };
};
