import { TestResultData } from "./output";
import fetch from "node-fetch";

export interface ReporterOptions {
  host: string;
  port: number;
}

export default class Reporter {
  private readonly opts: ReporterOptions;

  constructor(opts: ReporterOptions) {
    this.opts = opts;
  }

  clearDatabase = async () => {
    const res = await fetch(`${this.buildHostUri()}/reset`, {
      method: "POST"
    })

    if (!res.ok) {
      throw new Error('could not clear database')
    }
  }

  report = async (results: TestResultData[]): Promise<boolean> => {
    const response = await this.testResultsRequest(results);
    return response.ok;
  };

  private async testResultsRequest(results: TestResultData[]) {
    return await fetch(`${this.buildHostUri()}/report`, {
      method: "POST",
      body: JSON.stringify(results),
    });
  }

  private buildHostUri = () => {
    const { port, host } = this.opts;

    return `http://${host}:${port}`
  }
}
