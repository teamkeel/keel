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
  type: "zpl";
  name?: string;
  printer?: H extends Hardware ? H["printers"][number]["name"] : never;
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

// Future format support
// type PrintData =
//   | {
//       type: "zpl" | "text" | "html";
//       data: string | string[];
//       file: never;
//       url: never;
//     }
//   | {
//       file: File;
//       data: never;
//       url: never;
//       type: never;
//     }
//   | {
//       url: string;
//       data: never;
//       file: never;
//       type: never;
//     };

// The shape of the response over the API
export interface UiElementPrintApiResponse<>extends BaseUiDisplayResponse<"ui.interactive.print"> {
  title?: string;
  description?: string;
  data: {
    type: "url" | "text" | "html" | "zpl";
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

    if ("type" in d && d.type) {
      return {
        type: d.type,
        name: d.name,
        data: d.data ? (Array.isArray(d.data) ? d.data : [d.data]) : undefined,
        printer: d.printer,
        url: d.url,
      };
    }

    return null;
  });

  const data: UiElementPrintApiResponse["data"] = (
    await Promise.all(dataPromises)
  ).filter((x): x is NonNullable<typeof x> => x !== null);

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
