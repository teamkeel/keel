import fetch from "node-fetch";

interface ActionExecutorArgs {
  parentPort: number;
  host: string;
}

interface ExecuteArgs {
  actionName: string;

  payload: Record<string, any>;
}

interface ActionFailure {
  message: string;
}

interface ActionResponse<T> {
  object: T;
  error?: ActionFailure;
}

// Makes a request to the testing runtime host with 
export default class ActionExecutor {
  private readonly parentPort: number;
  private readonly host: string;

  constructor({ parentPort, host }: ActionExecutorArgs) {
    this.parentPort = parentPort;
    this.host = host;
  }

  execute = async<ActionReturnType> ({ actionName, payload }: ExecuteArgs): Promise<ActionResponse<ActionReturnType>> => {
    const res = await fetch(`http://${this.host}:${this.parentPort}/action`, {
      method: "POST",
      body: JSON.stringify({ actionName, payload }),
    });
    
    const json = (await res.json()) as ActionResponse<ActionReturnType>;

    return json;
  };
}
