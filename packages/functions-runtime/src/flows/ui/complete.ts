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

  /** Automatically close the flow once complete.
   * If set, the title and description will be shown in a notification.
   * If set, you cannot return content as this will not be shown. */
  autoClose?: boolean;
  data?: any;
} & (
  | {
      autoClose: true;
      content?: never;
    }
  | {
      autoClose?: false;
      content?: DisplayElementResponse[];
    }
);

export type Complete<C extends FlowConfig> = (
  options: CompleteOptions<C>
) => CompleteOptions<C> & { __type: "ui.complete" };

export interface UiCompleteApiResponse
  extends BaseUiDisplayResponse<"ui.complete"> {
  stage?: string;
  title?: string;
  description?: string;
  content: UiElementApiResponses;
  autoClose?: boolean;
}

export async function complete<C extends FlowConfig>(
  options: CompleteOptions<C>
): Promise<UiCompleteApiResponse> {
  // Turn these back into the actual response types
  const content =
    (options.content as unknown as ImplementationResponse<any, any>[]) || [];
  const contentUiConfig = (await Promise.all(
    content.map(async (c) => {
      return c.uiConfig;
    })
  )) as UiElementApiResponses;

  return {
    __type: "ui.complete",
    stage: options.stage,
    title: options.title,
    description: options.description,
    content: contentUiConfig || [],
    autoClose: options.autoClose,
  } satisfies UiCompleteApiResponse;
}
