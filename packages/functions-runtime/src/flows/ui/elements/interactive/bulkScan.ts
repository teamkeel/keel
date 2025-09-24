import {
  BaseUiMinimalInputResponse,
  InputElementImplementation,
  InputElementResponse,
  ValidateFn,
} from "../..";

export type UiElementBulkScan = {
  // Overloaded so without a config we return the simple string array response
  <N extends string>(
    name: N
  ): InputElementResponse<
    N,
    {
      scans: string[];
    }
  >;
  <N extends string, const M extends BulkScanDuplicateMode>(
    name: N,
    options?: BulkScanOptions<M>
  ): InputElementResponse<
    N,
    {
      scans: BulkScanResponseItem<M>[];
    }
  >;
};

type BulkScanResponseItem<M> = M extends "trackQuantity"
  ? {
      value: string;
      quantity: number;
    }
  : string;

/**
 * Defines how duplicate scans should be handled.
 * @default "none"
 */
type BulkScanDuplicateMode =
  /** No duplicate handling - all scans are accepted as separate entries */
  | "none"
  /** Track quantity when duplicates are scanned */
  | "trackQuantity"
  /** Reject duplicate scans with an error message */
  | "rejectDuplicates";

type BulkScanOptions<M> = {
  /** The singular unit of the item being scanned. E.g. "box", "bottle", "product" etc */
  unit?: string;
  /** The title of the input block
   * @default "Scan {unit plural | 'items'}"
   */
  title?: string;
  description?: string;
  max?: number;
  min?: number;
  duplicateHandling?: M;

  validate?: ValidateFn<BulkScanResponseItem<M>>;
};

// The shape of the response over the API
export interface UiElementBulkScanApiResponse
  extends BaseUiMinimalInputResponse<"ui.interactive.bulkScan"> {
  title?: string;
  description?: string;
  unit?: string;
  max?: number;
  min?: number;
  duplicateHandling: BulkScanDuplicateMode;
}

export const bulkScan: InputElementImplementation<
  any,
  UiElementBulkScan,
  UiElementBulkScanApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.interactive.bulkScan",
      name,
      title: options?.title ?? undefined,
      description: options?.description ?? undefined,
      unit: options?.unit ?? undefined,
      max: options?.max ?? undefined,
      min: options?.min ?? undefined,
      duplicateHandling: options?.duplicateHandling ?? "none",
    } satisfies UiElementBulkScanApiResponse,
    validate: async (data) => {
      // Ensure the response is an object with an items array
      if (!("scans" in data)) {
        return "Missing scans in response";
      }

      if (!Array.isArray(data.scans)) {
        return "Scans must be an array";
      }

      return options?.validate?.(data) ?? true;
    },
    getData: (x: any) => x,
  };
};
