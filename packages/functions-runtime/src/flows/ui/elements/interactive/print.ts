import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementWithRequiredConfig,
} from "../..";
import { File } from "../../../../File";

export type UiElementPrint = DisplayElementWithRequiredConfig<{
  jobs: PrintData[] | PrintData;
  title?: string;
  description?: string;
  /** If true, the jobs will be automatically printed. */
  autoPrint?: boolean;
  /** If true, the flow will continue after the jobs are complete. */
  autoContinue?: boolean;
}>;

type PrintData = {
  type: "zpl";
  name?: string;
  data: string | string[];
};

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
export interface UiElementPrintApiResponse
  extends BaseUiDisplayResponse<"ui.interactive.print"> {
  title?: string;
  description?: string;
  data: {
    type: "url" | "text" | "html" | "zpl";
    data?: string[];
    url?: string;
  }[];
  autoPrint: boolean;
}

export const print: DisplayElementImplementation<
  UiElementPrint,
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

    // if ("url" in d && d.url) {
    //   return {
    //     type: "url" as const,
    //     url: d.url,
    //   };
    // }

    if ("type" in d && d.type) {
      return {
        type: d.type,
        name: d.name,
        data: Array.isArray(d.data) ? d.data : [d.data],
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
    } satisfies UiElementPrintApiResponse,
  };
};
