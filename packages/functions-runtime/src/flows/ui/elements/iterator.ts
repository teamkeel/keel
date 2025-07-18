import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  IteratorElementImplementation,
  IteratorElementResponse,
  UiElementApiResponses,
  UIElements,
} from "..";

export type UiElementIterator = <N extends string, T extends UIElements>(
  name: N,
  options: {
    content: T;
    // validate?: (data: ExtractFormData<T>[]) => Promise<null | string> | string | null;
    max?: number;
    min?: number;
  }
) => IteratorElementResponse<N, T>;

// The shape of the response over the API
export interface UiElementIteratorApiResponse
  extends BaseUiDisplayResponse<"ui.iterator"> {
  name: string;
  content: UiElementApiResponses;
  max?: number;
  min?: number;
  validationErrors?: Array<{
    index: number;
    name: string;
    validationError: string;
  }>;
}

export const iterator: IteratorElementImplementation<
  any,
  UiElementIterator,
  UiElementIteratorApiResponse
> = (name, options) => {
  return {
    __type: "iterator",
    uiConfig: {
      __type: "ui.iterator",
      name,
      content: options.content as unknown as UiElementApiResponses, // It's not actually in shape yet but will be transformed by the page
      max: options.max,
      min: options.min,
    } satisfies UiElementIteratorApiResponse,
   // validate: options?.validate,
    getData: (x: any) => x,
  };
};
