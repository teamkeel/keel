
import { FlowConfig, ExtractStageKeys } from "..";
import {
  BaseUiDisplayResponse,
  ImplementationResponse,
  InputElementResponse,
  UiElementApiResponses,
  DisplayElementResponse,
  UIElements,
} from ".";

type CompletionOptions<C extends FlowConfig> = {
    stage?: ExtractStageKeys<C>;
    title?: string;
    description?: string;
    content: DisplayElementResponse[];
    data: any;
  };


export type Completion<C extends FlowConfig> = <
  T extends DisplayElementResponse[],
>(
  name: string,
  options: CompletionOptions<C>
) => void;


export interface CompleteApiResponse extends BaseUiDisplayResponse<"complete"> {
  stage?: string;
  title?: string;
  description?: string;
  content: UiElementApiResponses;
}

export async function complete<
  C extends FlowConfig,
  T extends UIElements,
>(
  options: CompletionOptions<C>
): Promise<{ complete: CompleteApiResponse }> {
  // Turn these back into the actual response types
  const content = options.content as unknown as ImplementationResponse<
    any,
    any
  >[];


  const contentUiConfig = (await Promise.all(
    content
      .map(async (c) => {
        return c.uiConfig;
      })
      .filter(Boolean)
  )) as UiElementApiResponses;

  return {
    complete: {
      __type: "complete",
      stage: options.stage,
      title: options.title,
      description: options.description,
      content: contentUiConfig,
    },
  };
}



// Extract the data from elements and return a key-value object based on the name of the element
type ExtractFormData<T extends UIElements> = {
    [K in Extract<T[number], InputElementResponse<string, any>>["name"]]: Extract<
      T[number],
      InputElementResponse<K, any>
    >["valueType"];
  };
  