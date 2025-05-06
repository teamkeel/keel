import {
  BaseUiDisplayResponse,
  DisplayElement,
  DisplayElementImplementation,
} from "../..";

export type UiElementDivider = DisplayElement<{}>;

// The shape of the response over the API
export interface UiElementDividerApiResponse
  extends BaseUiDisplayResponse<"ui.display.divider"> {}

export const divider: DisplayElementImplementation<
  UiElementDivider,
  UiElementDividerApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.divider",
    } satisfies UiElementDividerApiResponse,
  };
};
