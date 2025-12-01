import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementWithRequiredConfig,
} from "../..";
import { Hardware, NullableHardware } from "../../../index";

export type UiElementPrint<H extends NullableHardware> =
  DisplayElementWithRequiredConfig<{
    jobs: PrintData<H>[] | PrintData<H>;
    title?: string;
    description?: string;
    /** If true, the jobs will be automatically printed. */
    autoPrint?: boolean;
    /** If true, the step will continue after the jobs are complete (pending validation). */
    autoContinue?: boolean;
    /** Control whether users can reprint jobs
     * @default true
     */
    allowReprint?: boolean;
  }>;

type PrintData<H extends NullableHardware> = {
  /** The name of the print job. */
  name?: string;
  /** The printer to use for the print job. Printers are defined in keelconfig.yaml. */
  printer?: H extends Hardware ? H["printers"][number]["name"] : never;
} & (PrintDataZpl | PrintDataRawPdf);

type PrintDataZpl = {
  type: "zpl";
} & (
  | {
      data: string | string[];
      url?: never;
    }
  | {
      data?: never;
      url: string;
    }
);

type PrintDataRawPdf = {
  type: "rawPdf";
  url: string;
  data?: never;
  /** The DPI of the PDF
   * @default 300
   */
  dpi?: number;
  /** The width of the page in dots.
   * e.g. 4" at 300 dpi is 1200 dots.
   * @default 1200
   */
  pageWidth?: number;
  /** The height of the page in dots.
   * e.g. 6" at 300 dpi is 1800 dots.
   * @default 1800
   */
  pageHeight?: number;
};

// The shape of the response over the API
export interface UiElementPrintApiResponse<>extends BaseUiDisplayResponse<"ui.interactive.print"> {
  title?: string;
  description?: string;
  data: {
    type: "zpl" | "rawPdf";
    name?: string;
    data?: string[];
    url?: string;
    printer?: string;
    dpi?: number;
    pageWidth?: number;
    pageHeight?: number;
  }[];
  autoPrint: boolean;
  autoContinue: boolean;
  allowReprint: boolean;
}

export const print: DisplayElementImplementation<
  UiElementPrint<NullableHardware>,
  UiElementPrintApiResponse
> = async (options) => {
  const dataConfig = Array.isArray(options.jobs)
    ? options.jobs
    : [options.jobs];

  const dataPromises = dataConfig.map(async (d) => {
    // if ("file" in d && d.file) {
    //   return {
    //     type: "url" as const,
    //     url: (await d.file.getPresignedUrl()).toString(),
    //   };
    // }

    return {
      type: d.type,
      name: d.name,
      data:
        "data" in d && d.data
          ? Array.isArray(d.data)
            ? d.data
            : [d.data]
          : undefined,
      printer: d.printer,
      url: "url" in d && d.url ? d.url : undefined,
      ...(d.type === "rawPdf"
        ? {
            dpi: d.dpi,
            pageWidth: d.pageWidth,
            pageHeight: d.pageHeight,
          }
        : {}),
    } satisfies UiElementPrintApiResponse["data"][number];
  });

  const data = (await Promise.all(dataPromises)).filter(
    (x): x is NonNullable<typeof x> => x !== null
  );

  return {
    uiConfig: {
      __type: "ui.interactive.print",
      title: options.title,
      description: options.description,
      data,
      autoPrint: options.autoPrint ?? false,
      autoContinue: options.autoContinue ?? false,
      allowReprint: options.allowReprint ?? true,
    } satisfies UiElementPrintApiResponse,
  };
};
