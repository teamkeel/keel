import fetch, { RequestInit } from "node-fetch";
import { Identity, Logger, LogLevel } from '@teamkeel/sdk';

const logger = new Logger({ colorize: true })

interface ActionExecutorArgs {
  parentPort: number;
  host?: string;
  protocol?: string;
  identity?: Identity;
  debug?: boolean;
}

interface ExecuteArgs {
  actionName: string;
  identity?: Identity;
  payload: Record<string, any>;
}

interface ActionFailure {
  message: string;
}

interface ActionResponse<T> {
  result?: T;
  error?: ActionFailure;
}

const DEFAULT_HOST = 'localhost';
const DEFAULT_PROTOCOL = 'http';

// Makes a request to the testing runtime host with 
export default class ActionExecutor {
  private readonly parentPort: number;
  private readonly host: string;
  private readonly protocol: string;
  private readonly debug?: boolean;

  constructor({ parentPort, host, protocol, debug }: ActionExecutorArgs) {
    this.parentPort = parentPort;
    this.host = host || DEFAULT_HOST;
    this.protocol = protocol || DEFAULT_PROTOCOL;
    this.debug = debug || false;
  }

  execute = async<ActionReturnType> (args: ExecuteArgs): Promise<ActionResponse<ActionReturnType>> => {
    const requestInit : RequestInit = {
      method: "POST",
      body: JSON.stringify(args)
    }
    const url = `${this.protocol}://${this.host}:${this.parentPort}/action`
    const res = await fetch(url, requestInit);

    if (this.debug) {
      logger.log(`Request to ${url}`, LogLevel.Debug)
    }

    const json = (await res.json()) as ActionResponse<ActionReturnType>;
  
    if (this.debug) {
      logger.log(json, LogLevel.Debug)
    }

    return json;
  };
}
