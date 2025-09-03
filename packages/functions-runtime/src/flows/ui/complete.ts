import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  UiElementApiResponses,
  DisplayElementResponse,
} from ".";

type AllOptional<T> = {
  [K in keyof T]-?: {} extends Pick<T, K> ? never : K;
}[keyof T] extends never
  ? true
  : false;

type RestartConfig = {
  /**
   * If auto, the flow will restart automatically with a notification instead of a completion screen.
   * If manual, a button will be shown on the completion screen.
   * Default is manual.
   * */
  mode?: "manual" | "auto";
  buttonLabel?: string;
};

export type CompleteOptions<C extends FlowConfig, I> = {
  stage?: ExtractStageKeys<C>;
  title?: string;
  description?: string;

  /** Automatically close the flow once complete.
   * Title and description will be shown in a notification if provided.
   * If set, you cannot return content as this will not be shown. */
  autoClose?: boolean;

  /** Restart the flow once complete.
   * Title and description will be shown in a notification if provided.
   * If set, the flow will be restarted with the inputs provided.
   * If set, you cannot return content as this will not be shown. */
  allowRestart?: //
  // Quite messy types but this does the following:
  // This can be false to disable.
  // True if there are no inputs.
  // Otherwise the config object.
  // If the inputs are all optional then the inputs param is optional, otherwise required.
  | false
    | ([I] extends [never]
        ? boolean | RestartConfig
        : RestartConfig &
            (AllOptional<I> extends true
              ? {
                  inputs?: I;
                }
              : {
                  inputs: I;
                }));
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

export type Complete<C extends FlowConfig, I> = (
  options: CompleteOptions<C, I>
) => CompleteOptions<C, I> & { __type: "ui.complete" };

export interface UiCompleteApiResponse
  extends BaseUiDisplayResponse<"ui.complete"> {
  stage?: string;
  title?: string;
  description?: string;
  content: UiElementApiResponses;
  autoClose?: boolean;
  allowRestart?: {
    inputs?: any;
  } & RestartConfig;
}

export async function complete<C extends FlowConfig, I>(
  options: CompleteOptions<C, I>
): Promise<UiCompleteApiResponse> {
  // Turn these back into the actual response types
  const content =
    (options.content as unknown as ImplementationResponse<any, any>[]) || [];
  const contentUiConfig = (await Promise.all(
    content.map(async (c) => {
      return (await c).uiConfig;
    })
  )) as UiElementApiResponses;

  return {
    __type: "ui.complete",
    stage: options.stage,
    title: options.title,
    description: options.description,
    content: contentUiConfig || [],
    autoClose: options.autoClose,
    allowRestart:
      typeof options.allowRestart === "boolean"
        ? options.allowRestart
          ? { inputs: undefined }
          : undefined
        : options.allowRestart,
  } satisfies UiCompleteApiResponse;
}
