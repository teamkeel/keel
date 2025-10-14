import {
  BaseUiDisplayResponse,
  DisplayElementImplementation,
  DisplayElementWithRequiredConfig,
} from "../..";
import { File, FileDbRecord } from "../../../../File";

export type UiElementFile = DisplayElementWithRequiredConfig<{
  title?: string;
  file: File;
}>;

// The shape of the response over the API
export interface UiElementFileApiResponse
  extends BaseUiDisplayResponse<"ui.display.file"> {
  title: string;
  file?: FileDbRecord & { url: string };
}

export const file: DisplayElementImplementation<
  UiElementFile,
  UiElementFileApiResponse
> = async (options) => {
  const title = options.title || options.file.filename;
  const url = await options.file.getPresignedUrl();
  const metadata = await options.file.toJSON();
  return {
    uiConfig: {
      __type: "ui.display.file",
      title,
      file: metadata
        ? {
            url: url?.toString() || "",
            ...metadata,
          }
        : undefined,
    } satisfies UiElementFileApiResponse,
  };
};
