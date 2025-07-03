import {
  BaseUiDisplayResponse,
  BaseUiMinimalInputResponse,
  DisplayElementImplementation,
  InputElementImplementation,
  InputElementResponse,
  ValidateFn,
} from "../..";
import { ImageConfig } from "../common";

export type UiElementPickList = <
  N extends string,
  T extends Record<string, any>,
>(
  name: N,
  options: ListOptions<T>
) => InputElementResponse<
  N,
  {
    items: PickListResponseItem[];
  }
>;

type PickListResponseItem = {
  id: string;
  quantity: number;
  targetQuantity: number;
};

type PickListItem = {
  id: string;
  targetQuantity: number;
  title?: string;
  description?: string;
  image?: ImageConfig;
  barcodes?: string[];
};

type ListOptions<T> = {
  data: T[];
  render: (data: T) => PickListItem;
  validate?: ValidateFn<PickListResponseItem>;
};

// The shape of the response over the API
export interface UiElementPickListApiResponse
  extends BaseUiMinimalInputResponse<"ui.interactive.pickList"> {
  data: PickListItem[];
}

export const pickList: InputElementImplementation<
  any,
  UiElementPickList,
  UiElementPickListApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.interactive.pickList",
      name,
      data: options.data.map((item: any) => {
        const rendered = options.render(item);
        return {
          id: rendered.id,
          targetQuantity: rendered.targetQuantity,
          title: rendered.title,
          description: rendered.description,
          image: rendered.image ?? undefined,
          barcodes: rendered.barcodes ?? undefined,
        } satisfies PickListItem;
      }),
    } satisfies UiElementPickListApiResponse,
    validate: async (data) => {
      // Ensure the response is an object with an items array
      if (!("items" in data)) {
        return "Missing items in response";
      }

      if (!Array.isArray(data.items)) {
        return "Items must be an array";
      }

      if (
        data.items.some(
          (item: any) =>
            typeof item !== "object" ||
            typeof item.id !== "string" ||
            typeof item.qty !== "number"
        )
      ) {
        return "Invalid data";
      }

      return options?.validate?.(data) ?? null;
    },
    getData: (x: any) => x,
  };
};
