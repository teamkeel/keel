import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementResponse,
} from "../..";
import { ImageConfig } from "../common";

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
  data: ListItem[];
}

// The implementation
export const list: DisplayElementImplementation<
  UiElementList,
  UiElementListApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.list",
      data: options.data.map((item: any) => {
        const rendered = options.render(item);
        return {
          title: rendered.title,
          description: rendered.description,
          image: rendered.image,
        } satisfies ListItem;
      }),
    } satisfies UiElementListApiResponse,
  };
};
