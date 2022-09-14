import fetch, { RequestInit } from "node-fetch";
import { Identity } from "@teamkeel/sdk";
import {
  FunctionCreateResponse,
  FunctionDeleteResponse,
  FunctionGetResponse,
  FunctionUpdateResponse,
  FunctionListResponse,
} from "@teamkeel/sdk/returnTypes";

interface ActionExecutorArgs {
  parentPort: number;
  host?: string;
  protocol?: string;
  identity?: Identity;
}

interface ExecuteArgs {
  actionName: string;
  identity?: Identity;
  payload: Record<string, any>;
}

const DEFAULT_HOST = "localhost";
const DEFAULT_PROTOCOL = "http";

// The return type of the execute function can be one of the operation return types
// the payload differs between different actions
export type ReturnTypes<T> =
  | FunctionCreateResponse<T>
  | FunctionDeleteResponse<T>
  | FunctionGetResponse<T>
  | FunctionUpdateResponse<T>
  | FunctionListResponse<T>;

// Makes a request to the testing runtime host with
export default class ActionExecutor {
  private readonly parentPort: number;
  private readonly host: string;
  private readonly protocol: string;

  constructor({ parentPort, host, protocol }: ActionExecutorArgs) {
    this.parentPort = parentPort;
    this.host = host || DEFAULT_HOST;
    this.protocol = protocol || DEFAULT_PROTOCOL;
  }

  execute = async <ReturnType>(args: ExecuteArgs): Promise<ReturnType> => {
    const requestInit: RequestInit = {
      method: "POST",
      body: JSON.stringify(args),
    };

    const res = await fetch(
      `${this.protocol}://${this.host}:${this.parentPort}/action`,
      requestInit
    );

    // the json from the /action endpoint will conform to the expected
    // return type for the given operation
    const json = (await res.json()) as ReturnType;

    return json;
  };
}
