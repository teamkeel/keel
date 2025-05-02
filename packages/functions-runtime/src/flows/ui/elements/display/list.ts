import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementResponse,
} from "../..";

type ImageConfig = {
  url: string;
  alt?: string;
  fit?: "cover" | "contain";
};

type ListItem = {
  title?: string;
  description?: string;
  image?: ImageConfig;
};

type ListOptions<T> = {
  data: T[];
  render: (data: T) => ListItem;
};

// The types the user experiences
export type UiElementList = <T extends any>(
  options: ListOptions<T>
) => DisplayElementResponse;

// The shape of the response over the API
export interface UiElementListApiResponse
  extends BaseUiDisplayResponse<"ui.display.list"> {
  data: any[];
}

// The implementation
export const list: DisplayElementImplementation<
  UiElementList,
  UiElementListApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.list",
      data: options.data.map(options.render),
    } satisfies UiElementListApiResponse,
  };
};
