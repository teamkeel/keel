import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementResponse,
} from "../..";

type ImageConfig = {
  url: string;
  alt?: string;
  aspectRatio?: number;
  fit?: "cover" | "contain";
};

type GridItem = {
  title?: string;
  description?: string;
  image?: ImageConfig;
};

export type GridOptions<T> = {
  data: T[];
  render: (data: T) => GridItem;
};

// The types the user experiences
export type UiElementGrid = <T extends any>(
  options: GridOptions<T>
) => DisplayElementResponse;

// The shape of the response over the API
export interface UiElementGridApiResponse
  extends BaseUiDisplayResponse<"ui.display.grid"> {
  data: GridItem[];
}

// The implementation
export const grid: DisplayElementImplementation<
  UiElementGrid,
  UiElementGridApiResponse
> = (options) => {
  return {
    uiConfig: {
      __type: "ui.display.grid",
      data: options.data.map((item: any) => {
        const rendered = options.render(item);
        return {
          title: rendered.title,
          description: rendered.description,
          image: rendered.image,
        } satisfies GridItem;
      }),
    } satisfies UiElementGridApiResponse,
  };
};
