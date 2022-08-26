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

  report = async (results: TestResultData[]): Promise<boolean> => {
    const response = await this.testResultsRequest(results);
    return response.ok
  };

  private async testResultsRequest(results: TestResultData[]) {
    const { port, host } = this.opts;

    return await fetch(`http://${host}:${port}/report`, {
      method: "POST",
      body: JSON.stringify(results),
    });
  }
}
