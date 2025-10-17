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
  const M extends PickListInputModes = { scanner: true; manual: true },
>(
  name: N,
  options: PickListOptions<M, T>
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
  scannedBarcodes?: string[];
};

type PickListItem = {
  id: string;
  targetQuantity: number;
  title?: string;
  description?: string;
  image?: ImageConfig;
  barcodes?: string[];
};

/**
 * Defines how duplicate scans should be handled.
 * @default "increaseQuantity"
 */
type ScanDuplicateMode =
  /** Increase quantity when duplicates are scanned */
  | "increaseQuantity"
  /** Reject duplicate scans with an error message */
  | "rejectDuplicates";

/**
 * Defines how picking items should be handled. By default, all modes are enabled.
 */
type PickListInputModes = {
  /** Picking items can be done by scanning barcodes */
  scanner: boolean;
  /** Picking items can be done by using the add/remove buttons */
  manual: boolean;
};

type PickListOptions<M extends PickListInputModes, T> = {
  data: T[];
  render: (data: T) => PickListItem;
  supportedInputs?: M;
  validate?: ValidateFn<{ items: PickListResponseItem[] }>;
  duplicateHandling?: ScanDuplicateMode | undefined;
  /** If true, the step will continue after all items reach the target quantity (pending validation)
   * Only applied for scanner inputs.
   */
  autoContinue?: boolean;
};

// The shape of the response over the API
export interface UiElementPickListApiResponse
  extends BaseUiMinimalInputResponse<"ui.interactive.pickList"> {
  data: PickListItem[];
  supportedInputs: PickListInputModes;
  duplicateHandling: ScanDuplicateMode | undefined;
  autoContinue: boolean;
}

export const pickList: InputElementImplementation<
  { items: PickListResponseItem[] },
  UiElementPickList,
  UiElementPickListApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.interactive.pickList",
      name,
      supportedInputs: options.supportedInputs || {
        scanner: true,
        manual: true,
      },
      duplicateHandling: options.duplicateHandling,
      autoContinue: options.autoContinue ?? false,
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
            typeof item.quantity !== "number" ||
            typeof item.targetQuantity !== "number" ||
            (item.scannedBarcodes && !Array.isArray(item.scannedBarcodes))
        )
      ) {
        return "Invalid data";
      }

      return options?.validate?.(data) ?? true;
    },
    getData: (x: any) => x,
  };
};
