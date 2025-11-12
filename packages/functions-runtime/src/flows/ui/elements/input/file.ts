import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
  CallbackFn,
} from "../..";

import { File, FileDbRecord } from "../../../../File";

type ElementDataType = Partial<FileDbRecord>;

export type UiElementInputFile = InputElement<ElementDataType, {}, File>;

/**
 * key: string;
 * filename: string;
 * contentType: string;
 */
type PresignedUrlCallbackInput = Partial<FileDbRecord>;
type PresignedUrlCallbackResponse = {
  key: string;
  url: string;
};

// The shape of the response over the API
export interface UiElementInputFileApiResponse
  extends BaseUiInputResponse<"ui.input.file", ElementDataType> {}

export const fileInput: InputElementImplementation<
  ElementDataType,
  UiElementInputFile,
  UiElementInputFileApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.input.file",
      name,
      label: options?.label || name,
      optional: options?.optional || false,
      disabled: options?.disabled || false,
      helpText: options?.helpText,
    } satisfies UiElementInputFileApiResponse,
    validate: options?.validate,
    getData: (x: ElementDataType) => x,
    getPresignedUploadURL: (async (input: PresignedUrlCallbackInput) => {
      const file = new File(input);
      const url = await file.getPresignedUploadUrl();
      return {
        url: url.toString(),
        key: file.key,
      };
    }) as CallbackFn<PresignedUrlCallbackInput, PresignedUrlCallbackResponse>,
  };
};
