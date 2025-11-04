import {
  BaseUiMinimalInputResponse,
  InputElementImplementation,
  InputElementResponse,
  ValidateFn,
} from "../..";

export type UiElementScan = <
  N extends string,
  const M extends scanMode = "single",
  const D extends ScanDuplicateMode = "none",
>(
  name: N,
  options?: ScanOptions<M, D>
) => InputElementResponse<
  N,
  M extends "single" ? string : ScanResponseItem<D>[]
>;

type ScanResponseItem<D> = D extends "trackQuantity"
  ? {
    value: string;
    quantity: number;
  }
  : string;

/**
 * Defines how duplicate scans should be handled.
 * @default "none"
 */
type ScanDuplicateMode =
  /** No duplicate handling - all scans are accepted as separate entries */
  | "none"
  /** Track quantity when duplicates are scanned */
  | "trackQuantity"
  /** Reject duplicate scans with an error message */
  | "rejectDuplicates";

type scanMode = "single" | "multi";

type ScanOptions<M, D> = {
  /** The title of the input block
   * @default "Scan {unit plural | 'items'}"
   */
  title?: string;
  description?: string;
  /** The singular unit of the item being scanned. E.g. "box", "bottle", "product" etc */
  unit?: string;
  /** The mode of the scan input
   * @default "single"
   */
  mode: M | scanMode;

  validate?: ValidateFn<ScanResponseItem<D>>;
} & (M extends "multi"
  ? {
    max?: number;
    min?: number;
    duplicateHandling?: D | ScanDuplicateMode;
  }
  : {
    /** If true, the step will continue after a scan (pending validation).
     * @default false
     */
    autoContinue?: boolean;
  });

// The shape of the response over the API
export interface UiElementInputScanApiResponse
  extends BaseUiMinimalInputResponse<"ui.input.scan"> {
  mode: "single" | "multi";
  title?: string;
  description?: string;
  unit?: string;
  max?: number;
  min?: number;
  duplicateHandling: ScanDuplicateMode;
  autoContinue: boolean;
}

const isMultiMode = (opts: any): opts is ScanOptions<"multi", any> =>
  opts && opts.mode === "multi";
const isSingleMode = (opts: any): opts is ScanOptions<"single", any> =>
  opts && opts.mode === "single";

export const scan: InputElementImplementation<
  any,
  UiElementScan,
  UiElementInputScanApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.input.scan",
      name,
      title: options?.title ?? undefined,
      description: options?.description ?? undefined,
      unit: options?.unit ?? undefined,
      mode: options?.mode ?? "single",
      duplicateHandling: "none",
      autoContinue: false,
      ...(isMultiMode(options)
        ? {
          max: options.max ?? undefined,
          min: options.min ?? undefined,
          duplicateHandling: options.duplicateHandling ?? "none",
        }
        : {}),
      ...(isSingleMode(options)
        ? {
          autoContinue: options.autoContinue,
        }
        : {}),
    } satisfies UiElementInputScanApiResponse,
    validate: async (data, action) => {
      if (options?.validate) {
        return (options.validate as any)(data, action);
      }
      return true;
    },
    getData: (x: any) => x,
  };
};
