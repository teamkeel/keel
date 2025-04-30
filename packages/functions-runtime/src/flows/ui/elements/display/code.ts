import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

export type UiElementCode = DisplayElement<{
  code: string;
  language?: string; // TODO: type the supported languages
}>;

// The shape of the response over the API
export interface UiElementCodeApiResponse
  extends BaseUiDisplayResponse<"ui.display.code"> {
  code: string;
  language?: string;
}

export const code: DisplayElementImplementation<
  UiElementCode,
  UiElementCodeApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.code",
      code: options?.code || "",
      language: options?.language,
    } satisfies UiElementCodeApiResponse,
  };
};
