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
  }[];
  autoPrint: boolean;
  autoContinue: boolean;
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
    } satisfies UiElementPrintApiResponse,
  };
};
