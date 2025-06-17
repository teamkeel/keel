
import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  UiElementApiResponses,
  DisplayElementResponse,
} from ".";

export type CompleteOptions<C extends FlowConfig> = {
    stage?: ExtractStageKeys<C>;
    title?: string;
    description?: string;
    content: DisplayElementResponse[];
    data: any;
  };

export type Complete<C extends FlowConfig> = (
  options: CompleteOptions<C>
) => CompleteOptions<C>;;

export interface CompleteApiResponse extends BaseUiDisplayResponse<"ui.complete"> {
  stage?: string;
  title?: string;
  description?: string;
  content: UiElementApiResponses;
}

export async function complete<
  C extends FlowConfig,
>(
  options: CompleteOptions<C>
): Promise<{ complete: CompleteApiResponse }> {
  // Turn these back into the actual response types
  const content = options.content as unknown as ImplementationResponse<
    any,
    any
  >[];
console.log(content)
  const contentUiConfig = (await Promise.all(
    content
      .map(async (c) => {
        return c.uiConfig;
      })
  )) as UiElementApiResponses;

  return {
    complete: {
      __type: "ui.complete",
      stage: options.stage,
      title: options.title,
      description: options.description,
      content: contentUiConfig,
    },
  };
}