import fetch, { RequestInit } from "node-fetch";
import { Identity } from "@teamkeel/sdk";

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

interface ActionFailure {
  message: string;
}

// todo: update with proper types from sdk package
interface ActionResponse<T> {
  object?: T;
  error?: ActionFailure;
}

const DEFAULT_HOST = "localhost";
const DEFAULT_PROTOCOL = "http";

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

  execute = async <ActionReturnType>(
    args: ExecuteArgs
  ): Promise<ActionResponse<ActionReturnType>> => {
    const requestInit: RequestInit = {
      method: "POST",
      body: JSON.stringify(args),
    };

    const res = await fetch(
      `${this.protocol}://${this.host}:${this.parentPort}/action`,
      requestInit
    );
    const json = (await res.json()) as ActionResponse<ActionReturnType>;

    return json;
  };
}
