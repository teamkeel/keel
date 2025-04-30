import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

type BannerMode = "info" | "warning" | "error" | "success";

export type UiElementBanner = DisplayElement<{
  title: string;
  description: string;
  mode?: BannerMode;
}>;

// The shape of the response over the API
export interface UiElementBannerApiResponse
  extends BaseUiDisplayResponse<"ui.display.banner"> {
  title: string;
  description: string;
  mode: BannerMode;
}

export const banner: DisplayElementImplementation<
  UiElementBanner,
  UiElementBannerApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.banner",
      title: options?.title || "",
      description: options?.description || "",
      mode: options?.mode || "info",
    } satisfies UiElementBannerApiResponse,
  };
};
